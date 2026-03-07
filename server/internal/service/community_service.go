package service

import (
	"context"
	"errors"

	"github.com/cashyalla/aquaverse/internal/domain"
)

var ErrLocaleMismatch = errors.New("locale mismatch: board and post locale must match")

type CreatePostRequest struct {
	BoardID    int64
	AuthorID   string
	Locale     domain.Locale
	Title      string   `json:"title"`
	Content    string   `json:"content"`
	FishDataID *int64   `json:"fish_data_id,omitempty"`
	ImageURLs  []string `json:"image_urls,omitempty"`
}

type CreateCommentRequest struct {
	PostID   int64
	AuthorID string
	ParentID *int64 `json:"parent_id,omitempty"`
	Content  string `json:"content"`
}

type PostListResult struct {
	Items      []domain.Post `json:"items"`
	TotalCount int           `json:"total_count"`
	Page       int           `json:"page"`
	Limit      int           `json:"limit"`
}

type CommunityRepository interface {
	GetBoard(ctx context.Context, boardID int64) (*domain.Board, error)
	ListBoards(ctx context.Context, locale domain.Locale) ([]domain.Board, error)
	ListPosts(ctx context.Context, boardID int64, page, limit int) ([]domain.Post, int, error)
	GetPost(ctx context.Context, postID int64) (*domain.Post, error)
	CreatePost(ctx context.Context, post *domain.Post) error
	CreateComment(ctx context.Context, comment *domain.Comment) error
	ToggleLike(ctx context.Context, postID int64, userID string) (bool, error)
	IncrementViewCount(ctx context.Context, postID int64) error
}

type CommunityService struct {
	repo CommunityRepository
}

func NewCommunityService(repo CommunityRepository) *CommunityService {
	return &CommunityService{repo: repo}
}

func (s *CommunityService) ListBoards(ctx context.Context, locale domain.Locale) ([]domain.Board, error) {
	return s.repo.ListBoards(ctx, locale)
}

func (s *CommunityService) ListPosts(ctx context.Context, boardID int64, locale domain.Locale, page, limit int) (*PostListResult, error) {
	// 게시판 로케일 엄격 검증
	board, err := s.repo.GetBoard(ctx, boardID)
	if err != nil {
		return nil, err
	}
	if board.Locale != locale {
		return nil, ErrLocaleMismatch
	}

	posts, total, err := s.repo.ListPosts(ctx, boardID, page, limit)
	if err != nil {
		return nil, err
	}
	return &PostListResult{Items: posts, TotalCount: total, Page: page, Limit: limit}, nil
}

func (s *CommunityService) GetPost(ctx context.Context, boardID, postID int64, locale domain.Locale) (*domain.Post, error) {
	board, err := s.repo.GetBoard(ctx, boardID)
	if err != nil {
		return nil, err
	}
	if board.Locale != locale {
		return nil, ErrLocaleMismatch
	}

	post, err := s.repo.GetPost(ctx, postID)
	if err != nil {
		return nil, err
	}
	if post.Locale != locale {
		return nil, ErrLocaleMismatch
	}

	// 조회수 비동기 증가
	go func() {
		_ = s.repo.IncrementViewCount(context.Background(), postID)
	}()

	return post, nil
}

func (s *CommunityService) CreatePost(ctx context.Context, req CreatePostRequest) (*domain.Post, error) {
	board, err := s.repo.GetBoard(ctx, req.BoardID)
	if err != nil {
		return nil, err
	}
	// 게시판 로케일 엄격 검증: 작성자 로케일이 게시판 로케일과 일치해야 함
	if board.Locale != req.Locale {
		return nil, ErrLocaleMismatch
	}
	if len(req.Title) < 2 || len(req.Title) > 200 {
		return nil, errors.New("title must be 2-200 characters")
	}
	if len(req.Content) < 10 {
		return nil, errors.New("content too short")
	}

	// 사용자 UUID 파싱
	// 실제 구현에서는 UUID 파싱 추가
	post := &domain.Post{
		BoardID:    req.BoardID,
		Locale:     req.Locale,
		Title:      req.Title,
		Content:    req.Content,
		ContentType: "MARKDOWN",
		FishDataID: req.FishDataID,
		ImageURLs:  req.ImageURLs,
	}

	if err := s.repo.CreatePost(ctx, post); err != nil {
		return nil, err
	}
	return post, nil
}

func (s *CommunityService) CreateComment(ctx context.Context, req CreateCommentRequest) (*domain.Comment, error) {
	if len(req.Content) < 1 {
		return nil, errors.New("comment cannot be empty")
	}
	comment := &domain.Comment{
		PostID:   req.PostID,
		ParentID: req.ParentID,
		Content:  req.Content,
	}
	if err := s.repo.CreateComment(ctx, comment); err != nil {
		return nil, err
	}
	return comment, nil
}

func (s *CommunityService) ToggleLike(ctx context.Context, postID int64, userID string) (bool, error) {
	return s.repo.ToggleLike(ctx, postID, userID)
}
