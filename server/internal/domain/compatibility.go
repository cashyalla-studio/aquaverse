package domain

// FishCompatibility 어종 호환성
type FishCompatibility struct {
	FishAID    int64  `db:"fish_a_id" json:"fish_a_id"`
	FishBID    int64  `db:"fish_b_id" json:"fish_b_id"`
	Compatible bool   `db:"compatible" json:"compatible"`
	Caution    bool   `db:"caution" json:"caution"`
	Reason     string `db:"reason" json:"reason,omitempty"`
}

// CompatibilityResult 호환성 체크 결과
type CompatibilityResult struct {
	Compatible bool   `json:"compatible"`
	Caution    bool   `json:"caution"`
	Reason     string `json:"reason,omitempty"`
	Source     string `json:"source"` // "database" or "rule"
}

// FishRecommendation 추천 어종
type FishRecommendation struct {
	FishID     int64   `json:"fish_id"`
	FishName   string  `json:"fish_name"`
	ScientName string  `json:"scientific_name,omitempty"`
	Score      float64 `json:"score"`
	Reason     string  `json:"reason"`
	ImageURL   string  `json:"image_url,omitempty"`
}
