package service

import (
	"context"
	"fmt"
	"time"

	"github.com/cashyalla/aquaverse/internal/domain"
	"github.com/jmoiron/sqlx"
	"github.com/minio/minio-go/v7"
)

type VideoService struct {
	db    *sqlx.DB
	minio *minio.Client
}

func NewVideoService(db *sqlx.DB, minioClient *minio.Client) *VideoService {
	return &VideoService{db: db, minio: minioClient}
}

const videoBucket = "videos"

// CreatePost 영상 게시물 등록
func (s *VideoService) CreatePost(ctx context.Context, userID, title, description, videoKey, thumbnailKey string, durationSec int) (*domain.VideoPost, error) {
	var id int64
	err := s.db.QueryRowContext(ctx, `
        INSERT INTO video_posts (user_id, title, description, video_key, thumbnail_key, duration_sec)
        VALUES ($1,$2,$3,$4,$5,$6)
        RETURNING id`,
		userID, title, description, videoKey, thumbnailKey, durationSec,
	).Scan(&id)
	if err != nil {
		return nil, err
	}
	return &domain.VideoPost{
		ID: id, UserID: userID, Title: title,
		Description: description, VideoKey: videoKey,
		ThumbnailKey: thumbnailKey, DurationSec: durationSec,
		Status: "ACTIVE",
	}, nil
}

// GetFeed 영상 피드 (최신순)
func (s *VideoService) GetFeed(ctx context.Context, userID string, limit, offset int) ([]domain.VideoPost, error) {
	if limit <= 0 || limit > 30 {
		limit = 10
	}
	var posts []domain.VideoPost
	var err error
	if userID != "" {
		err = s.db.SelectContext(ctx, &posts, `
            SELECT vp.*, u.username,
                   CASE WHEN vl.user_id IS NOT NULL THEN true ELSE false END as is_liked
            FROM video_posts vp
            JOIN users u ON u.id = vp.user_id
            LEFT JOIN video_likes vl ON vl.video_id = vp.id AND vl.user_id=$1::uuid
            WHERE vp.status='ACTIVE'
            ORDER BY vp.created_at DESC
            LIMIT $2 OFFSET $3
        `, userID, limit, offset)
	} else {
		err = s.db.SelectContext(ctx, &posts, `
            SELECT vp.*, u.username, false as is_liked
            FROM video_posts vp
            JOIN users u ON u.id = vp.user_id
            WHERE vp.status='ACTIVE'
            ORDER BY vp.created_at DESC
            LIMIT $1 OFFSET $2
        `, limit, offset)
	}
	if err != nil {
		return nil, err
	}

	// presigned URL 생성 (MinIO)
	for i := range posts {
		if s.minio != nil && posts[i].VideoKey != "" {
			url, err := s.minio.PresignedGetObject(ctx, videoBucket, posts[i].VideoKey, 24*time.Hour, nil)
			if err == nil {
				posts[i].VideoURL = url.String()
			}
		}
	}
	return posts, nil
}

// LikePost 좋아요 토글
func (s *VideoService) LikePost(ctx context.Context, videoID int64, userID string) (bool, error) {
	var existing int
	s.db.QueryRowContext(ctx, `SELECT 1 FROM video_likes WHERE video_id=$1 AND user_id=$2::uuid`, videoID, userID).Scan(&existing)

	if existing == 1 {
		// 좋아요 취소
		s.db.ExecContext(ctx, `DELETE FROM video_likes WHERE video_id=$1 AND user_id=$2::uuid`, videoID, userID)
		s.db.ExecContext(ctx, `UPDATE video_posts SET like_count=GREATEST(0,like_count-1) WHERE id=$1`, videoID)
		return false, nil
	}
	// 좋아요
	s.db.ExecContext(ctx, `INSERT INTO video_likes (video_id, user_id) VALUES ($1,$2::uuid) ON CONFLICT DO NOTHING`, videoID, userID)
	s.db.ExecContext(ctx, `UPDATE video_posts SET like_count=like_count+1 WHERE id=$1`, videoID)
	return true, nil
}

// IncrementView 조회수 증가
func (s *VideoService) IncrementView(ctx context.Context, videoID int64) {
	s.db.ExecContext(ctx, `UPDATE video_posts SET view_count=view_count+1 WHERE id=$1`, videoID)
}

// DeletePost 영상 삭제 (소프트 삭제)
func (s *VideoService) DeletePost(ctx context.Context, videoID int64, userID string) error {
	result, err := s.db.ExecContext(ctx, `
        UPDATE video_posts SET status='DELETED' WHERE id=$1 AND user_id=$2::uuid
    `, videoID, userID)
	if err != nil {
		return err
	}
	n, _ := result.RowsAffected()
	if n == 0 {
		return fmt.Errorf("권한 없음 또는 게시물 없음")
	}
	return nil
}
