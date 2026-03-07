package domain

import (
	"time"

	"github.com/google/uuid"
)

// Locale 목록 - 지원하는 13개 로케일
type Locale string

const (
	LocaleKO   Locale = "ko"
	LocaleENUS Locale = "en-US"
	LocaleENGB Locale = "en-GB"
	LocaleENAU Locale = "en-AU"
	LocaleJA   Locale = "ja"
	LocaleZHCN Locale = "zh-CN"
	LocaleZHTW Locale = "zh-TW"
	LocaleDE   Locale = "de"
	LocaleFRFR Locale = "fr-FR"
	LocaleFRCA Locale = "fr-CA"
	LocaleES   Locale = "es"
	LocalePT   Locale = "pt"
	LocaleAR   Locale = "ar" // RTL
	LocaleHE   Locale = "he" // RTL
)

// RTL 언어 여부
var RTLLocales = map[Locale]bool{
	LocaleAR: true,
	LocaleHE: true,
}

func (l Locale) IsRTL() bool {
	return RTLLocales[l]
}

func (l Locale) IsValid() bool {
	validLocales := []Locale{
		LocaleKO, LocaleENUS, LocaleENGB, LocaleENAU,
		LocaleJA, LocaleZHCN, LocaleZHTW, LocaleDE,
		LocaleFRFR, LocaleFRCA, LocaleES, LocalePT,
		LocaleAR, LocaleHE,
	}
	for _, v := range validLocales {
		if l == v {
			return true
		}
	}
	return false
}

// User Role
type UserRole string

const (
	RoleUser      UserRole = "USER"
	RoleModerator UserRole = "MODERATOR"
	RoleAdmin     UserRole = "ADMIN"
)

// User 핵심 도메인 모델
type User struct {
	ID             uuid.UUID  `db:"id"`
	Email          string     `db:"email"`
	Username       string     `db:"username"`
	PasswordHash   string     `db:"password_hash"`
	DisplayName    string     `db:"display_name"`
	AvatarURL      *string    `db:"avatar_url"`
	Bio            *string    `db:"bio"`
	PreferredLocale Locale    `db:"preferred_locale"`
	Role           UserRole   `db:"role"`

	// 인증 상태
	EmailVerified    bool       `db:"email_verified"`
	PhoneVerified    bool       `db:"phone_verified"`
	PhoneNumber      *string    `db:"phone_number"`

	// 신뢰도 (마켓플레이스)
	TrustScore     float64    `db:"trust_score"`

	// 위치 (마켓플레이스 - 암호화 저장)
	CountryCode    *string    `db:"country_code"`
	City           *string    `db:"city"`

	// 상태
	IsActive       bool       `db:"is_active"`
	IsBanned       bool       `db:"is_banned"`
	BanReason      *string    `db:"ban_reason"`

	CreatedAt      time.Time  `db:"created_at"`
	UpdatedAt      time.Time  `db:"updated_at"`
	LastLoginAt    *time.Time `db:"last_login_at"`
}

// UserProfile - 공개 프로필 (민감 정보 제외)
type UserProfile struct {
	ID          uuid.UUID `json:"id"`
	Username    string    `json:"username"`
	DisplayName string    `json:"display_name"`
	AvatarURL   *string   `json:"avatar_url,omitempty"`
	Bio         *string   `json:"bio,omitempty"`
	TrustScore  float64   `json:"trust_score"`
	Role        UserRole  `json:"role"`
	CreatedAt   time.Time `json:"created_at"`

	// 마켓플레이스 통계 (JOIN)
	TotalTrades     int     `json:"total_trades"`
	CompletedTrades int     `json:"completed_trades"`
	AvgRating       float64 `json:"avg_rating"`
	Badges          []string `json:"badges"`
}

// RefreshToken
type RefreshToken struct {
	ID        uuid.UUID `db:"id"`
	UserID    uuid.UUID `db:"user_id"`
	TokenHash string    `db:"token_hash"`
	DeviceInfo *string  `db:"device_info"`
	ExpiresAt time.Time `db:"expires_at"`
	CreatedAt time.Time `db:"created_at"`
	RevokedAt *time.Time `db:"revoked_at"`
}

// TOTPStatus TOTP 2단계 인증 상태
type TOTPStatus struct {
	Enabled    bool       `json:"totp_enabled" db:"totp_enabled"`
	VerifiedAt *time.Time `json:"verified_at" db:"totp_verified_at"`
}
