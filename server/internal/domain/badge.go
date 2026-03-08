package domain

import (
	"time"

	"github.com/google/uuid"
)

// BadgeDefinition represents a badge that can be earned.
type BadgeDefinition struct {
	Code           string `json:"code"            db:"code"`
	Name           string `json:"name"            db:"name"`
	Description    string `json:"description"     db:"description"`
	IconEmoji      string `json:"icon_emoji"      db:"icon_emoji"`
	Category       string `json:"category"        db:"category"`
	ConditionType  string `json:"condition_type"  db:"condition_type"`
	ConditionValue int    `json:"condition_value" db:"condition_value"`
	IsActive       bool   `json:"is_active"       db:"is_active"`
}

// UserBadge records a badge that has been awarded to a user.
type UserBadge struct {
	ID        int64     `json:"id"         db:"id"`
	UserID    uuid.UUID `json:"user_id"    db:"user_id"`
	BadgeCode string    `json:"badge_code" db:"badge_code"`
	EarnedAt  time.Time `json:"earned_at"  db:"earned_at"`

	// Joined from badge_definitions when fetched together.
	Name      string `json:"name,omitempty"       db:"name"`
	IconEmoji string `json:"icon_emoji,omitempty" db:"icon_emoji"`
	Category  string `json:"category,omitempty"   db:"category"`
}

// Challenge is a time-limited activity challenge.
type Challenge struct {
	ID             int64      `json:"id"              db:"id"`
	Title          string     `json:"title"           db:"title"`
	Description    string     `json:"description"     db:"description"`
	BadgeCode      *string    `json:"badge_code"      db:"badge_code"`
	StartsAt       time.Time  `json:"starts_at"       db:"starts_at"`
	EndsAt         time.Time  `json:"ends_at"         db:"ends_at"`
	ConditionType  string     `json:"condition_type"  db:"condition_type"`
	ConditionValue int        `json:"condition_value" db:"condition_value"`
	IsActive       bool       `json:"is_active"       db:"is_active"`
}

// ChallengeParticipant holds a user's participation state in a challenge.
type ChallengeParticipant struct {
	ChallengeID int64      `json:"challenge_id" db:"challenge_id"`
	UserID      uuid.UUID  `json:"user_id"      db:"user_id"`
	Progress    int        `json:"progress"     db:"progress"`
	Completed   bool       `json:"completed"    db:"completed"`
	JoinedAt    time.Time  `json:"joined_at"    db:"joined_at"`
	CompletedAt *time.Time `json:"completed_at" db:"completed_at"`
}
