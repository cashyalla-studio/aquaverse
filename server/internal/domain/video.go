package domain

// VideoPost 수조 영상 게시물
type VideoPost struct {
	ID           int64  `db:"id" json:"id"`
	UserID       string `db:"user_id" json:"user_id"`
	Username     string `db:"username" json:"username,omitempty"` // JOIN
	Title        string `db:"title" json:"title"`
	Description  string `db:"description" json:"description,omitempty"`
	VideoKey     string `db:"video_key" json:"video_key"`
	VideoURL     string `json:"video_url,omitempty"`     // presigned URL
	ThumbnailKey string `db:"thumbnail_key" json:"thumbnail_key,omitempty"`
	ThumbnailURL string `json:"thumbnail_url,omitempty"` // presigned URL
	DurationSec  int    `db:"duration_sec" json:"duration_sec,omitempty"`
	ViewCount    int    `db:"view_count" json:"view_count"`
	LikeCount    int    `db:"like_count" json:"like_count"`
	IsLiked      bool   `db:"is_liked" json:"is_liked"` // 현재 사용자 좋아요 여부
	Status       string `db:"status" json:"status"`
	CreatedAt    string `db:"created_at" json:"created_at"`
}
