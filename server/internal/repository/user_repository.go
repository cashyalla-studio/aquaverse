package repository

import (
	"context"
	"database/sql"
	"errors"

	"github.com/cashyalla/aquaverse/internal/domain"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
)

type UserRepository struct {
	db *sqlx.DB
}

func NewUserRepository(db *sqlx.DB) *UserRepository {
	return &UserRepository{db: db}
}

func (r *UserRepository) Create(ctx context.Context, user *domain.User) error {
	q := `
		INSERT INTO users (email, username, password_hash, display_name, preferred_locale, role)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING id, created_at, updated_at
	`
	return r.db.QueryRowContext(ctx, q,
		user.Email, user.Username, user.PasswordHash,
		user.DisplayName, string(user.PreferredLocale), string(user.Role),
	).Scan(&user.ID, &user.CreatedAt, &user.UpdatedAt)
}

func (r *UserRepository) GetByEmail(ctx context.Context, email string) (*domain.User, error) {
	var user domain.User
	q := `SELECT * FROM users WHERE email = $1 AND is_active = TRUE`
	if err := r.db.GetContext(ctx, &user, q, email); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, errors.New("user not found")
		}
		return nil, err
	}
	return &user, nil
}

func (r *UserRepository) GetByID(ctx context.Context, id uuid.UUID) (*domain.User, error) {
	var user domain.User
	q := `SELECT * FROM users WHERE id = $1 AND is_active = TRUE`
	if err := r.db.GetContext(ctx, &user, q, id); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, errors.New("user not found")
		}
		return nil, err
	}
	return &user, nil
}

func (r *UserRepository) GetByUsername(ctx context.Context, username string) (*domain.User, error) {
	var user domain.User
	q := `SELECT * FROM users WHERE username = $1 AND is_active = TRUE`
	if err := r.db.GetContext(ctx, &user, q, username); err != nil {
		return nil, errors.New("user not found")
	}
	return &user, nil
}

func (r *UserRepository) GetProfile(ctx context.Context, userID uuid.UUID) (*domain.UserProfile, error) {
	var profile domain.UserProfile
	q := `
		SELECT
			u.id, u.username, u.display_name, u.avatar_url, u.bio, u.trust_score, u.role, u.created_at,
			COALESCE(ts.total_trades, 0) AS total_trades,
			COALESCE(ts.completed_trades, 0) AS completed_trades,
			COALESCE(ts.avg_rating, 0) AS avg_rating,
			COALESCE(ts.badges, '[]'::jsonb) AS badges
		FROM users u
		LEFT JOIN user_trust_scores ts ON ts.user_id = u.id
		WHERE u.id = $1 AND u.is_active = TRUE
	`
	if err := r.db.GetContext(ctx, &profile, q, userID); err != nil {
		return nil, errors.New("user not found")
	}
	return &profile, nil
}

func (r *UserRepository) UpdateLastLogin(ctx context.Context, userID uuid.UUID) error {
	_, err := r.db.ExecContext(ctx, `UPDATE users SET last_login_at = NOW() WHERE id = $1`, userID)
	return err
}

func (r *UserRepository) ExistsEmail(ctx context.Context, email string) (bool, error) {
	var exists bool
	err := r.db.QueryRowContext(ctx, `SELECT EXISTS(SELECT 1 FROM users WHERE email = $1)`, email).Scan(&exists)
	return exists, err
}

func (r *UserRepository) ExistsUsername(ctx context.Context, username string) (bool, error) {
	var exists bool
	err := r.db.QueryRowContext(ctx, `SELECT EXISTS(SELECT 1 FROM users WHERE username = $1)`, username).Scan(&exists)
	return exists, err
}

// RefreshToken 관련
func (r *UserRepository) SaveRefreshToken(ctx context.Context, token *domain.RefreshToken) error {
	q := `
		INSERT INTO refresh_tokens (user_id, token_hash, device_info, expires_at)
		VALUES ($1, $2, $3, $4)
		RETURNING id, created_at
	`
	return r.db.QueryRowContext(ctx, q, token.UserID, token.TokenHash, token.DeviceInfo, token.ExpiresAt).
		Scan(&token.ID, &token.CreatedAt)
}

func (r *UserRepository) GetRefreshToken(ctx context.Context, tokenHash string) (*domain.RefreshToken, error) {
	var token domain.RefreshToken
	q := `SELECT * FROM refresh_tokens WHERE token_hash = $1 AND revoked_at IS NULL AND expires_at > NOW()`
	if err := r.db.GetContext(ctx, &token, q, tokenHash); err != nil {
		return nil, errors.New("invalid or expired refresh token")
	}
	return &token, nil
}

func (r *UserRepository) RevokeRefreshToken(ctx context.Context, tokenHash string) error {
	_, err := r.db.ExecContext(ctx, `UPDATE refresh_tokens SET revoked_at = NOW() WHERE token_hash = $1`, tokenHash)
	return err
}

func (r *UserRepository) RevokeAllRefreshTokens(ctx context.Context, userID uuid.UUID) error {
	_, err := r.db.ExecContext(ctx, `UPDATE refresh_tokens SET revoked_at = NOW() WHERE user_id = $1 AND revoked_at IS NULL`, userID)
	return err
}
