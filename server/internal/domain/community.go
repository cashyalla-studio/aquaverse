package domain

import (
	"time"

	"github.com/google/uuid"
)

// BoardCategory 게시판 카테고리 (언어별 엄격 분리)
type BoardCategory string

const (
	BoardGeneral    BoardCategory = "GENERAL"    // 자유게시판
	BoardQuestion   BoardCategory = "QUESTION"   // 질문/답변
	BoardShowcase   BoardCategory = "SHOWCASE"   // 수조 자랑
	BoardBreeding   BoardCategory = "BREEDING"   // 번식 일지
	BoardDiseases   BoardCategory = "DISEASES"   // 질병/치료
	BoardEquipment  BoardCategory = "EQUIPMENT"  // 장비/용품
	BoardNews       BoardCategory = "NEWS"        // 뉴스/공지
)

// Board 게시판 (로케일당 1개씩, 엄격 분리)
type Board struct {
	ID          int64         `db:"id"`
	Locale      Locale        `db:"locale"`       // 게시판 소속 로케일 (엄격 분리)
	Category    BoardCategory `db:"category"`
	Name        string        `db:"name"`
	Description *string       `db:"description"`
	IsRTL       bool          `db:"is_rtl"`
	SortOrder   int           `db:"sort_order"`
	PostCount   int           `db:"post_count"`
	IsActive    bool          `db:"is_active"`
	CreatedAt   time.Time     `db:"created_at"`
}

// Post 게시글
type Post struct {
	ID          int64         `db:"id"`
	BoardID     int64         `db:"board_id"`
	AuthorID    uuid.UUID     `db:"author_id"`
	Locale      Locale        `db:"locale"`       // 게시글 로케일 (보드 로케일과 일치해야 함)
	Title       string        `db:"title"`
	Content     string        `db:"content"`
	ContentType string        `db:"content_type"` // MARKDOWN, HTML
	ImageURLs   []string      `db:"image_urls"`

	// 연관 어종 (선택적)
	FishDataID  *int64        `db:"fish_data_id"`

	// 집계
	ViewCount   int           `db:"view_count"`
	LikeCount   int           `db:"like_count"`
	CommentCount int          `db:"comment_count"`

	// 상태
	IsPinned    bool          `db:"is_pinned"`
	IsLocked    bool          `db:"is_locked"`
	IsDeleted   bool          `db:"is_deleted"`

	CreatedAt   time.Time     `db:"created_at"`
	UpdatedAt   time.Time     `db:"updated_at"`
	DeletedAt   *time.Time    `db:"deleted_at"`

	// 조인 데이터
	Author      *UserProfile  `db:"-"`
}

// Comment 댓글
type Comment struct {
	ID         int64      `db:"id"`
	PostID     int64      `db:"post_id"`
	AuthorID   uuid.UUID  `db:"author_id"`
	ParentID   *int64     `db:"parent_id"` // 대댓글
	Content    string     `db:"content"`
	LikeCount  int        `db:"like_count"`
	IsDeleted  bool       `db:"is_deleted"`
	CreatedAt  time.Time  `db:"created_at"`
	UpdatedAt  time.Time  `db:"updated_at"`

	Author   *UserProfile `db:"-"`
	Replies  []Comment    `db:"-"`
}

// PostLike 게시글 좋아요
type PostLike struct {
	PostID    int64     `db:"post_id"`
	UserID    uuid.UUID `db:"user_id"`
	CreatedAt time.Time `db:"created_at"`
}

// Tank 수조 (사육 일지 연동)
type Tank struct {
	ID          int64     `db:"id"`
	OwnerID     uuid.UUID `db:"owner_id"`
	Name        string    `db:"name"`
	SizeLiters  int       `db:"size_liters"`
	SetupDate   *time.Time `db:"setup_date"`
	Description *string   `db:"description"`
	ImageURL    *string   `db:"image_url"`

	// 수질 파라미터
	CurrentPH      *float64 `db:"current_ph"`
	CurrentTempC   *float64 `db:"current_temp_c"`
	CurrentNH3     *float64 `db:"current_nh3"`    // 암모니아 ppm
	CurrentNO2     *float64 `db:"current_no2"`    // 아질산 ppm
	CurrentNO3     *float64 `db:"current_no3"`    // 질산 ppm
	LastWaterChange *time.Time `db:"last_water_change"`

	IsPublic  bool      `db:"is_public"`
	CreatedAt time.Time `db:"created_at"`
	UpdatedAt time.Time `db:"updated_at"`

	// 조인
	Inhabitants []TankInhabitant `db:"-"`
}

// TankInhabitant 수조 입주 어종
type TankInhabitant struct {
	ID         int64      `db:"id"`
	TankID     int64      `db:"tank_id"`
	FishDataID *int64     `db:"fish_data_id"`
	CustomName string     `db:"custom_name"` // fish_data 없을 때
	Quantity   int        `db:"quantity"`
	AddedAt    time.Time  `db:"added_at"`
	RemovedAt  *time.Time `db:"removed_at"`
	Notes      *string    `db:"notes"`
}
