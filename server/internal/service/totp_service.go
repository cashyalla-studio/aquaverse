package service

import (
	"context"
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha1"
	"crypto/sha256"
	"encoding/base32"
	"encoding/base64"
	"encoding/binary"
	"fmt"
	"math"
	"time"

	"github.com/cashyalla/aquaverse/internal/domain"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
)

type TOTPService struct {
	db *sqlx.DB
}

func NewTOTPService(db *sqlx.DB) *TOTPService {
	return &TOTPService{db: db}
}

// GenerateSecret TOTP 비밀키 생성 (Base32 인코딩)
func (s *TOTPService) GenerateSecret() (string, error) {
	secret := make([]byte, 20)
	if _, err := rand.Read(secret); err != nil {
		return "", err
	}
	return base32.StdEncoding.WithPadding(base32.NoPadding).EncodeToString(secret), nil
}

// GenerateOTPAuthURL QR코드용 URI 생성
func (s *TOTPService) GenerateOTPAuthURL(userEmail, secret string) string {
	return fmt.Sprintf(
		"otpauth://totp/Finara:%s?secret=%s&issuer=Finara&algorithm=SHA1&digits=6&period=30",
		userEmail, secret,
	)
}

// ValidateTOTP TOTP 코드 검증 (±1 step 허용)
func (s *TOTPService) ValidateTOTP(secret, code string) bool {
	now := time.Now().Unix()
	for _, delta := range []int64{-1, 0, 1} {
		counter := uint64((now/30) + delta)
		expected := s.generateHOTP(secret, counter)
		if code == expected {
			return true
		}
	}
	return false
}

func (s *TOTPService) generateHOTP(secret string, counter uint64) string {
	keyBytes, err := base32.StdEncoding.WithPadding(base32.NoPadding).DecodeString(secret)
	if err != nil {
		return ""
	}
	counterBytes := make([]byte, 8)
	binary.BigEndian.PutUint64(counterBytes, counter)

	mac := hmac.New(sha1.New, keyBytes)
	mac.Write(counterBytes)
	h := mac.Sum(nil)

	offset := h[len(h)-1] & 0x0F
	code := binary.BigEndian.Uint32(h[offset:offset+4]) & 0x7FFFFFFF
	otp := int(code) % int(math.Pow10(6))
	return fmt.Sprintf("%06d", otp)
}

// EnableTOTP TOTP 활성화 시작 (비밀키를 임시 저장)
func (s *TOTPService) EnableTOTP(ctx context.Context, userID uuid.UUID) (string, string, error) {
	secret, err := s.GenerateSecret()
	if err != nil {
		return "", "", err
	}

	// 비밀키를 DB에 저장 (아직 enabled=false)
	_, err = s.db.ExecContext(ctx, `
		UPDATE users SET totp_secret = $1, totp_enabled = FALSE WHERE id = $2
	`, secret, userID)
	if err != nil {
		return "", "", err
	}

	var email string
	s.db.QueryRowContext(ctx, `SELECT email FROM users WHERE id = $1`, userID).Scan(&email)

	qrURL := s.GenerateOTPAuthURL(email, secret)
	return secret, qrURL, nil
}

// VerifyAndActivateTOTP 코드 확인 후 TOTP 활성화
func (s *TOTPService) VerifyAndActivateTOTP(ctx context.Context, userID uuid.UUID, code string) ([]string, error) {
	var secret string
	err := s.db.QueryRowContext(ctx, `SELECT COALESCE(totp_secret, '') FROM users WHERE id = $1`, userID).Scan(&secret)
	if err != nil || secret == "" {
		return nil, fmt.Errorf("totp not initialized")
	}

	if !s.ValidateTOTP(secret, code) {
		return nil, fmt.Errorf("invalid code")
	}

	// TOTP 활성화
	_, err = s.db.ExecContext(ctx, `
		UPDATE users SET totp_enabled = TRUE, totp_verified_at = NOW() WHERE id = $1
	`, userID)
	if err != nil {
		return nil, err
	}

	// 백업 코드 8개 생성
	codes := make([]string, 8)
	for i := range codes {
		raw := make([]byte, 6)
		rand.Read(raw)
		code := base64.RawURLEncoding.EncodeToString(raw)[:8]
		codes[i] = code
		hash := fmt.Sprintf("%x", sha256.Sum256([]byte(code)))
		s.db.ExecContext(ctx, `
			INSERT INTO totp_backup_codes (user_id, code_hash) VALUES ($1, $2)
		`, userID, hash)
	}

	return codes, nil
}

// DisableTOTP TOTP 비활성화
func (s *TOTPService) DisableTOTP(ctx context.Context, userID uuid.UUID, code string) error {
	var secret string
	var enabled bool
	err := s.db.QueryRowContext(ctx, `
		SELECT COALESCE(totp_secret, ''), totp_enabled FROM users WHERE id = $1
	`, userID).Scan(&secret, &enabled)
	if err != nil {
		return err
	}
	if !enabled {
		return fmt.Errorf("totp not enabled")
	}
	if !s.ValidateTOTP(secret, code) {
		return fmt.Errorf("invalid code")
	}
	_, err = s.db.ExecContext(ctx, `
		UPDATE users SET totp_secret = NULL, totp_enabled = FALSE, totp_verified_at = NULL WHERE id = $1
	`, userID)
	return err
}

// GetTOTPStatus 현재 TOTP 상태 조회
func (s *TOTPService) GetTOTPStatus(ctx context.Context, userID uuid.UUID) (*domain.TOTPStatus, error) {
	var status domain.TOTPStatus
	err := s.db.QueryRowContext(ctx, `
		SELECT totp_enabled, totp_verified_at FROM users WHERE id = $1
	`, userID).Scan(&status.Enabled, &status.VerifiedAt)
	return &status, err
}
