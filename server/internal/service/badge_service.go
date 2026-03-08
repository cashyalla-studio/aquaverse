package service

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/cashyalla/aquaverse/internal/domain"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
)

// BadgeService handles badge definitions, user badge awards, and challenges.
type BadgeService struct {
	db *sqlx.DB
}

func NewBadgeService(db *sqlx.DB) *BadgeService {
	return &BadgeService{db: db}
}

// ─── Badge definitions ─────────────────────────────────────────────────────

// ListBadgeDefinitions returns all active badge definitions.
func (s *BadgeService) ListBadgeDefinitions(ctx context.Context) ([]domain.BadgeDefinition, error) {
	var badges []domain.BadgeDefinition
	err := s.db.SelectContext(ctx, &badges, `
		SELECT code, name, COALESCE(description, '') AS description,
		       COALESCE(icon_emoji, '') AS icon_emoji,
		       category, condition_type, condition_value, is_active
		FROM badge_definitions
		WHERE is_active = TRUE
		ORDER BY category, code
	`)
	return badges, err
}

// ─── User badges ───────────────────────────────────────────────────────────

// GetUserBadges returns all badges earned by the given user, joined with
// the badge name, icon, and category for display convenience.
func (s *BadgeService) GetUserBadges(ctx context.Context, userID uuid.UUID) ([]domain.UserBadge, error) {
	var badges []domain.UserBadge
	err := s.db.SelectContext(ctx, &badges, `
		SELECT ub.id, ub.user_id, ub.badge_code, ub.earned_at,
		       bd.name, COALESCE(bd.icon_emoji, '') AS icon_emoji, bd.category
		FROM user_badges ub
		JOIN badge_definitions bd ON bd.code = ub.badge_code
		WHERE ub.user_id = $1
		ORDER BY ub.earned_at DESC
	`, userID)
	return badges, err
}

// AwardBadge grants badgeCode to userID. It is idempotent: if the user
// already holds the badge the call succeeds without modifying anything.
// This function is exported so that other services (e.g. community,
// marketplace) can call it after a user reaches an activity milestone.
func AwardBadge(ctx context.Context, db *sqlx.DB, userID uuid.UUID, badgeCode string) error {
	_, err := db.ExecContext(ctx, `
		INSERT INTO user_badges (user_id, badge_code)
		VALUES ($1, $2)
		ON CONFLICT (user_id, badge_code) DO NOTHING
	`, userID, badgeCode)
	return err
}

// ─── Challenges ────────────────────────────────────────────────────────────

// ListActiveChallenges returns challenges that are active and have not yet ended.
func (s *BadgeService) ListActiveChallenges(ctx context.Context) ([]domain.Challenge, error) {
	var challenges []domain.Challenge
	err := s.db.SelectContext(ctx, &challenges, `
		SELECT id, title, COALESCE(description, '') AS description,
		       badge_code, starts_at, ends_at,
		       condition_type, condition_value, is_active
		FROM challenges
		WHERE is_active = TRUE AND ends_at > NOW()
		ORDER BY ends_at ASC
	`)
	return challenges, err
}

// JoinChallenge registers the user as a participant in the given challenge.
// Returns an error if the challenge does not exist, is inactive, or has ended.
func (s *BadgeService) JoinChallenge(ctx context.Context, challengeID int64, userID uuid.UUID) error {
	var ch domain.Challenge
	err := s.db.GetContext(ctx, &ch, `
		SELECT id, title, COALESCE(description, '') AS description,
		       badge_code, starts_at, ends_at,
		       condition_type, condition_value, is_active
		FROM challenges
		WHERE id = $1
	`, challengeID)
	if errors.Is(err, sql.ErrNoRows) {
		return errors.New("challenge not found")
	}
	if err != nil {
		return err
	}
	if !ch.IsActive {
		return errors.New("challenge is not active")
	}
	if time.Now().After(ch.EndsAt) {
		return errors.New("challenge has ended")
	}

	_, err = s.db.ExecContext(ctx, `
		INSERT INTO challenge_participants (challenge_id, user_id)
		VALUES ($1, $2)
		ON CONFLICT (challenge_id, user_id) DO NOTHING
	`, challengeID, userID)
	return err
}

// GetProgress returns the caller's participation record for a challenge.
// Returns an error with message "not joined" when the user has not joined.
func (s *BadgeService) GetProgress(ctx context.Context, challengeID int64, userID uuid.UUID) (*domain.ChallengeParticipant, error) {
	var p domain.ChallengeParticipant
	err := s.db.GetContext(ctx, &p, `
		SELECT challenge_id, user_id, progress, completed, joined_at, completed_at
		FROM challenge_participants
		WHERE challenge_id = $1 AND user_id = $2
	`, challengeID, userID)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, errors.New("not joined")
	}
	return &p, err
}
