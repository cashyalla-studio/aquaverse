package repository

import (
	"context"
	"database/sql"
	"errors"

	"github.com/cashyalla/aquaverse/internal/domain"
	"github.com/jmoiron/sqlx"
)

type CommunityRepository struct {
	db *sqlx.DB
}

func NewCommunityRepository(db *sqlx.DB) *CommunityRepository {
	return &CommunityRepository{db: db}
}

func (r *CommunityRepository) GetBoard(ctx context.Context, boardID int64) (*domain.Board, error) {
	var board domain.Board
	q := `SELECT id, locale, category, name, description, is_rtl, sort_order, post_count, is_active, created_at FROM boards WHERE id = $1 AND is_active = TRUE`
	if err := r.db.GetContext(ctx, &board, q, boardID); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, errors.New("board not found")
		}
		return nil, err
	}
	return &board, nil
}

func (r *CommunityRepository) ListBoards(ctx context.Context, locale domain.Locale) ([]domain.Board, error) {
	var boards []domain.Board
	q := `
		SELECT id, locale, category, name, description, is_rtl, sort_order, post_count, is_active, created_at
		FROM boards
		WHERE locale = $1 AND is_active = TRUE
		ORDER BY sort_order ASC
	`
	if err := r.db.SelectContext(ctx, &boards, q, string(locale)); err != nil {
		return nil, err
	}
	return boards, nil
}

func (r *CommunityRepository) ListPosts(ctx context.Context, boardID int64, page, limit int) ([]domain.Post, int, error) {
	offset := (page - 1) * limit

	var total int
	cq := `SELECT COUNT(*) FROM posts WHERE board_id = $1 AND is_deleted = FALSE`
	if err := r.db.GetContext(ctx, &total, cq, boardID); err != nil {
		return nil, 0, err
	}

	q := `
		SELECT
			p.id, p.board_id, p.author_id, p.locale, p.title,
			left(p.content, 300) AS content,
			p.content_type, p.image_urls, p.fish_data_id,
			p.view_count, p.like_count, p.comment_count,
			p.is_pinned, p.is_locked, p.is_deleted,
			p.created_at, p.updated_at
		FROM posts p
		WHERE p.board_id = $1 AND p.is_deleted = FALSE
		ORDER BY p.is_pinned DESC, p.created_at DESC
		LIMIT $2 OFFSET $3
	`
	var posts []domain.Post
	if err := r.db.SelectContext(ctx, &posts, q, boardID, limit, offset); err != nil {
		return nil, 0, err
	}
	return posts, total, nil
}

func (r *CommunityRepository) GetPost(ctx context.Context, postID int64) (*domain.Post, error) {
	var post domain.Post
	q := `
		SELECT
			p.id, p.board_id, p.author_id, p.locale, p.title, p.content,
			p.content_type, p.image_urls, p.fish_data_id,
			p.view_count, p.like_count, p.comment_count,
			p.is_pinned, p.is_locked, p.is_deleted,
			p.created_at, p.updated_at
		FROM posts p
		WHERE p.id = $1 AND p.is_deleted = FALSE
	`
	if err := r.db.GetContext(ctx, &post, q, postID); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, errors.New("post not found")
		}
		return nil, err
	}

	// 댓글 로드 (최상위만, 대댓글은 별도 요청)
	var comments []domain.Comment
	cq := `
		SELECT id, post_id, author_id, parent_id, content, like_count, is_deleted, created_at, updated_at
		FROM comments
		WHERE post_id = $1 AND parent_id IS NULL AND is_deleted = FALSE
		ORDER BY created_at ASC
		LIMIT 50
	`
	if err := r.db.SelectContext(ctx, &comments, cq, postID); err == nil {
		post.Replies = nil // Post에 Replies 필드 없음, Comment에 있음
		_ = comments      // TODO: 응답 구조체에서 활용
	}

	return &post, nil
}

func (r *CommunityRepository) CreatePost(ctx context.Context, post *domain.Post) error {
	q := `
		INSERT INTO posts (board_id, author_id, locale, title, content, content_type, image_urls, fish_data_id)
		VALUES ($1, $2, $3, $4, $5, $6, $7::jsonb, $8)
		RETURNING id, created_at, updated_at
	`
	imageURLsJSON := "[]"
	if len(post.ImageURLs) > 0 {
		imageURLsJSON = toJSONArray(post.ImageURLs)
	}

	return r.db.QueryRowContext(ctx, q,
		post.BoardID, post.AuthorID, string(post.Locale),
		post.Title, post.Content, post.ContentType,
		imageURLsJSON, post.FishDataID,
	).Scan(&post.ID, &post.CreatedAt, &post.UpdatedAt)
}

func (r *CommunityRepository) CreateComment(ctx context.Context, comment *domain.Comment) error {
	q := `
		INSERT INTO comments (post_id, author_id, parent_id, content)
		VALUES ($1, $2, $3, $4)
		RETURNING id, created_at, updated_at
	`
	if err := r.db.QueryRowContext(ctx, q,
		comment.PostID, comment.AuthorID, comment.ParentID, comment.Content,
	).Scan(&comment.ID, &comment.CreatedAt, &comment.UpdatedAt); err != nil {
		return err
	}

	// comment_count 증가
	_, err := r.db.ExecContext(ctx,
		`UPDATE posts SET comment_count = comment_count + 1, updated_at = NOW() WHERE id = $1`,
		comment.PostID,
	)
	return err
}

func (r *CommunityRepository) ToggleLike(ctx context.Context, postID int64, userID string) (bool, error) {
	tx, err := r.db.BeginTxx(ctx, nil)
	if err != nil {
		return false, err
	}
	defer tx.Rollback()

	// 이미 좋아요했는지 확인
	var exists bool
	tx.QueryRowContext(ctx, `SELECT EXISTS(SELECT 1 FROM post_likes WHERE post_id=$1 AND user_id=$2)`, postID, userID).Scan(&exists)

	if exists {
		// 좋아요 취소
		tx.ExecContext(ctx, `DELETE FROM post_likes WHERE post_id=$1 AND user_id=$2`, postID, userID)
		tx.ExecContext(ctx, `UPDATE posts SET like_count = GREATEST(0, like_count - 1) WHERE id=$1`, postID)
		tx.Commit()
		return false, nil
	}

	// 좋아요 추가
	tx.ExecContext(ctx, `INSERT INTO post_likes (post_id, user_id) VALUES ($1, $2) ON CONFLICT DO NOTHING`, postID, userID)
	tx.ExecContext(ctx, `UPDATE posts SET like_count = like_count + 1 WHERE id=$1`, postID)
	tx.Commit()
	return true, nil
}

func (r *CommunityRepository) IncrementViewCount(ctx context.Context, postID int64) error {
	_, err := r.db.ExecContext(ctx, `UPDATE posts SET view_count = view_count + 1 WHERE id = $1`, postID)
	return err
}

// Tank 관련
func (r *CommunityRepository) ListTanks(ctx context.Context, ownerID string, publicOnly bool) ([]domain.Tank, error) {
	q := `
		SELECT id, owner_id, name, size_liters, setup_date, description, image_url,
		       current_ph, current_temp_c, current_nh3, current_no2, current_no3,
		       last_water_change, is_public, created_at, updated_at
		FROM tanks WHERE owner_id = $1
	`
	if publicOnly {
		q += " AND is_public = TRUE"
	}
	q += " ORDER BY created_at DESC"

	var tanks []domain.Tank
	return tanks, r.db.SelectContext(ctx, &tanks, q, ownerID)
}

func (r *CommunityRepository) CreateTank(ctx context.Context, tank *domain.Tank) error {
	q := `
		INSERT INTO tanks (owner_id, name, size_liters, setup_date, description, is_public)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING id, created_at, updated_at
	`
	return r.db.QueryRowContext(ctx, q,
		tank.OwnerID, tank.Name, tank.SizeLiters, tank.SetupDate, tank.Description, tank.IsPublic,
	).Scan(&tank.ID, &tank.CreatedAt, &tank.UpdatedAt)
}

func toJSONArray(items []string) string {
	if len(items) == 0 {
		return "[]"
	}
	parts := make([]string, len(items))
	for i, s := range items {
		parts[i] = `"` + s + `"`
	}
	return "[" + join(parts, ",") + "]"
}

func join(parts []string, sep string) string {
	result := ""
	for i, p := range parts {
		if i > 0 {
			result += sep
		}
		result += p
	}
	return result
}
