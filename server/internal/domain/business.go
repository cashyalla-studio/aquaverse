package domain

// BusinessProfile 업체 프로필
type BusinessProfile struct {
	ID          int64    `db:"id" json:"id"`
	UserID      string   `db:"user_id" json:"user_id"`
	StoreName   string   `db:"store_name" json:"store_name"`
	Description string   `db:"description" json:"description,omitempty"`
	Address     string   `db:"address" json:"address,omitempty"`
	City        string   `db:"city" json:"city,omitempty"`
	Phone       string   `db:"phone" json:"phone,omitempty"`
	Website     string   `db:"website" json:"website,omitempty"`
	LogoURL     string   `db:"logo_url" json:"logo_url,omitempty"`
	IsVerified  bool     `db:"is_verified" json:"is_verified"`
	Lat         *float64 `db:"lat" json:"lat,omitempty"`
	Lng         *float64 `db:"lng" json:"lng,omitempty"`
	CreatedAt   string   `db:"created_at" json:"created_at"`
	UpdatedAt   string   `db:"updated_at" json:"updated_at"`
	// 집계 필드 (JOIN)
	AvgRating   float64 `db:"avg_rating" json:"avg_rating,omitempty"`
	ReviewCount int     `db:"review_count" json:"review_count,omitempty"`
	DistanceKm  float64 `db:"distance_km" json:"distance_km,omitempty"`
}

// BusinessReview 업체 리뷰
type BusinessReview struct {
	ID           int64  `db:"id" json:"id"`
	BusinessID   int64  `db:"business_id" json:"business_id"`
	ReviewerID   string `db:"reviewer_id" json:"reviewer_id"`
	ReviewerName string `db:"reviewer_name" json:"reviewer_name,omitempty"`
	Rating       int    `db:"rating" json:"rating"`
	Content      string `db:"content" json:"content,omitempty"`
	CreatedAt    string `db:"created_at" json:"created_at"`
}
