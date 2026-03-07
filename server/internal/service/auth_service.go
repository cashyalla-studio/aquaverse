package service

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"time"

	"github.com/cashyalla/aquaverse/internal/config"
	"github.com/cashyalla/aquaverse/internal/domain"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

type RegisterRequest struct {
	Email          string        `json:"email"`
	Username       string        `json:"username"`
	Password       string        `json:"password"`
	DisplayName    string        `json:"display_name"`
	PreferredLocale domain.Locale `json:"preferred_locale"`
}

type LoginRequest struct {
	Email      string `json:"email"`
	Password   string `json:"password"`
	DeviceInfo string `json:"device_info,omitempty"`
}

type TokenPair struct {
	AccessToken  string    `json:"access_token"`
	RefreshToken string    `json:"refresh_token"`
	ExpiresAt    time.Time `json:"expires_at"`
}

type UserRepository interface {
	Create(ctx context.Context, user *domain.User) error
	GetByEmail(ctx context.Context, email string) (*domain.User, error)
	GetByID(ctx context.Context, id uuid.UUID) (*domain.User, error)
	ExistsEmail(ctx context.Context, email string) (bool, error)
	ExistsUsername(ctx context.Context, username string) (bool, error)
	UpdateLastLogin(ctx context.Context, userID uuid.UUID) error
	SaveRefreshToken(ctx context.Context, token *domain.RefreshToken) error
	GetRefreshToken(ctx context.Context, tokenHash string) (*domain.RefreshToken, error)
	RevokeRefreshToken(ctx context.Context, tokenHash string) error
	RevokeAllRefreshTokens(ctx context.Context, userID uuid.UUID) error
}

type AuthService struct {
	userRepo UserRepository
	cfg      config.AuthConfig
}

func NewAuthService(userRepo UserRepository, cfg config.AuthConfig) *AuthService {
	return &AuthService{userRepo: userRepo, cfg: cfg}
}

func (s *AuthService) Register(ctx context.Context, req RegisterRequest) (*domain.User, error) {
	// 유효성 검사
	if len(req.Email) < 5 || len(req.Email) > 320 {
		return nil, errors.New("invalid email")
	}
	if len(req.Username) < 3 || len(req.Username) > 50 {
		return nil, errors.New("username must be 3-50 characters")
	}
	if len(req.Password) < 8 {
		return nil, errors.New("password must be at least 8 characters")
	}
	if !req.PreferredLocale.IsValid() {
		req.PreferredLocale = domain.LocaleENUS
	}

	// 중복 확인
	if exists, _ := s.userRepo.ExistsEmail(ctx, req.Email); exists {
		return nil, errors.New("email already registered")
	}
	if exists, _ := s.userRepo.ExistsUsername(ctx, req.Username); exists {
		return nil, errors.New("username already taken")
	}

	// 비밀번호 해시
	hash, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, fmt.Errorf("password hashing failed: %w", err)
	}

	user := &domain.User{
		Email:           req.Email,
		Username:        req.Username,
		PasswordHash:    string(hash),
		DisplayName:     req.DisplayName,
		PreferredLocale: req.PreferredLocale,
		Role:            domain.RoleUser,
		IsActive:        true,
	}
	if user.DisplayName == "" {
		user.DisplayName = req.Username
	}

	if err := s.userRepo.Create(ctx, user); err != nil {
		return nil, fmt.Errorf("user creation failed: %w", err)
	}

	return user, nil
}

func (s *AuthService) Login(ctx context.Context, req LoginRequest) (*TokenPair, error) {
	user, err := s.userRepo.GetByEmail(ctx, req.Email)
	if err != nil {
		return nil, errors.New("invalid email or password")
	}
	if user.IsBanned {
		return nil, errors.New("account is banned")
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.Password)); err != nil {
		return nil, errors.New("invalid email or password")
	}

	// 토큰 발급
	pair, err := s.issueTokenPair(ctx, user, req.DeviceInfo)
	if err != nil {
		return nil, err
	}

	go s.userRepo.UpdateLastLogin(context.Background(), user.ID)
	return pair, nil
}

func (s *AuthService) Refresh(ctx context.Context, rawRefreshToken string) (*TokenPair, error) {
	tokenHash := hashToken(rawRefreshToken)
	stored, err := s.userRepo.GetRefreshToken(ctx, tokenHash)
	if err != nil {
		return nil, errors.New("invalid or expired refresh token")
	}

	user, err := s.userRepo.GetByID(ctx, stored.UserID)
	if err != nil {
		return nil, errors.New("user not found")
	}

	// 기존 토큰 폐기 (Refresh Token Rotation)
	_ = s.userRepo.RevokeRefreshToken(ctx, tokenHash)

	return s.issueTokenPair(ctx, user, "")
}

func (s *AuthService) Logout(ctx context.Context, rawRefreshToken string) error {
	tokenHash := hashToken(rawRefreshToken)
	return s.userRepo.RevokeRefreshToken(ctx, tokenHash)
}

func (s *AuthService) LogoutAll(ctx context.Context, userID uuid.UUID) error {
	return s.userRepo.RevokeAllRefreshTokens(ctx, userID)
}

// issueTokenPair 액세스 + 리프레시 토큰 발급
func (s *AuthService) issueTokenPair(ctx context.Context, user *domain.User, deviceInfo string) (*TokenPair, error) {
	// Access Token (JWT)
	expiresAt := time.Now().Add(time.Duration(s.cfg.AccessTokenExpiry) * time.Minute)
	claims := jwt.MapClaims{
		"uid":  user.ID.String(),
		"role": string(user.Role),
		"exp":  expiresAt.Unix(),
		"iat":  time.Now().Unix(),
	}
	accessToken, err := jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString([]byte(s.cfg.JWTSecret))
	if err != nil {
		return nil, fmt.Errorf("access token signing failed: %w", err)
	}

	// Refresh Token (랜덤 32바이트)
	rawRefresh := generateSecureToken()
	refreshHash := hashToken(rawRefresh)
	refreshExpiry := time.Now().AddDate(0, 0, s.cfg.RefreshTokenExpiry)

	var di *string
	if deviceInfo != "" {
		di = &deviceInfo
	}
	rt := &domain.RefreshToken{
		UserID:     user.ID,
		TokenHash:  refreshHash,
		DeviceInfo: di,
		ExpiresAt:  refreshExpiry,
	}
	if err := s.userRepo.SaveRefreshToken(ctx, rt); err != nil {
		return nil, fmt.Errorf("refresh token save failed: %w", err)
	}

	return &TokenPair{
		AccessToken:  accessToken,
		RefreshToken: rawRefresh,
		ExpiresAt:    expiresAt,
	}, nil
}

func hashToken(raw string) string {
	h := sha256.Sum256([]byte(raw))
	return hex.EncodeToString(h[:])
}

func generateSecureToken() string {
	b := make([]byte, 32)
	rand.Read(b)
	return hex.EncodeToString(b)
}
