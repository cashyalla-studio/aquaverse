package service

import (
	"context"
	"encoding/json"

	"github.com/cashyalla/aquaverse/internal/domain"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
)

type SocialService struct {
	db *sqlx.DB
}

func NewSocialService(db *sqlx.DB) *SocialService {
	return &SocialService{db: db}
}

func (s *SocialService) Follow(ctx context.Context, followerID, followingID uuid.UUID) error {
	_, err := s.db.ExecContext(ctx, `
		INSERT INTO user_follows (follower_id, following_id) VALUES ($1, $2)
		ON CONFLICT DO NOTHING
	`, followerID, followingID)
	return err
}

func (s *SocialService) Unfollow(ctx context.Context, followerID, followingID uuid.UUID) error {
	_, err := s.db.ExecContext(ctx, `
		DELETE FROM user_follows WHERE follower_id = $1 AND following_id = $2
	`, followerID, followingID)
	return err
}

func (s *SocialService) GetFollowing(ctx context.Context, userID uuid.UUID, limit, offset int) ([]domain.UserFollow, error) {
	var follows []domain.UserFollow
	err := s.db.SelectContext(ctx, &follows, `
		SELECT follower_id::text, following_id::text, created_at
		FROM user_follows WHERE follower_id = $1
		ORDER BY created_at DESC LIMIT $2 OFFSET $3
	`, userID, limit, offset)
	return follows, err
}

func (s *SocialService) GetFollowers(ctx context.Context, userID uuid.UUID, limit, offset int) ([]domain.UserFollow, error) {
	var follows []domain.UserFollow
	err := s.db.SelectContext(ctx, &follows, `
		SELECT follower_id::text, following_id::text, created_at
		FROM user_follows WHERE following_id = $1
		ORDER BY created_at DESC LIMIT $2 OFFSET $3
	`, userID, limit, offset)
	return follows, err
}

func (s *SocialService) GetFeed(ctx context.Context, userID uuid.UUID, limit, offset int) ([]domain.ActivityItem, error) {
	var items []domain.ActivityItem
	err := s.db.SelectContext(ctx, &items, `
		SELECT af.id, af.actor_id::text, u.username AS actor_name,
		       af.verb, COALESCE(af.object_type, '') AS object_type,
		       af.object_id, af.object_data, af.created_at
		FROM activity_feed af
		JOIN users u ON u.id = af.actor_id
		WHERE af.actor_id IN (
		    SELECT following_id FROM user_follows WHERE follower_id = $1
		)
		ORDER BY af.created_at DESC
		LIMIT $2 OFFSET $3
	`, userID, limit, offset)
	return items, err
}

func (s *SocialService) GetSuggestions(ctx context.Context, userID uuid.UUID) ([]domain.FollowSuggestion, error) {
	// 같은 어종 관심사를 가진 사용자 (watches 테이블 기반)
	var suggestions []domain.FollowSuggestion
	err := s.db.SelectContext(ctx, &suggestions, `
		SELECT DISTINCT u.id::text AS user_id, u.username, u.trust_score,
		       COALESCE(fd.name, '') AS common_fish
		FROM users u
		LEFT JOIN watches w ON w.user_id = u.id
		LEFT JOIN fish_data fd ON fd.id = w.fish_data_id
		WHERE u.id != $1
		AND u.id NOT IN (SELECT following_id FROM user_follows WHERE follower_id = $1)
		AND u.is_banned = false
		ORDER BY u.trust_score DESC
		LIMIT 10
	`, userID)
	return suggestions, err
}

func (s *SocialService) PublishActivity(ctx context.Context, actorID uuid.UUID, verb, objectType string, objectID *int64, data interface{}) error {
	var jsonData []byte
	if data != nil {
		jsonData, _ = json.Marshal(data)
	}
	_, err := s.db.ExecContext(ctx, `
		INSERT INTO activity_feed (actor_id, verb, object_type, object_id, object_data)
		VALUES ($1, $2, $3, $4, $5)
	`, actorID, verb, objectType, objectID, jsonData)
	return err
}
