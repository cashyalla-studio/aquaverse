package domain

// Tank 수조 정보
type Tank struct {
	ID        int64  `db:"id" json:"id"`
	UserID    string `db:"user_id" json:"user_id"`
	Name      string `db:"name" json:"name"`
	VolumeL   *int   `db:"volume_l" json:"volume_l,omitempty"`
	CreatedAt string `db:"created_at" json:"created_at"`
}

// TankInhabitant 수조 입주 어종
type TankInhabitant struct {
	TankID     int64  `db:"tank_id" json:"tank_id"`
	FishDataID int64  `db:"fish_data_id" json:"fish_data_id"`
	Quantity   int    `db:"quantity" json:"quantity"`
	FishName   string `db:"fish_name" json:"fish_name,omitempty"` // JOIN 결과
}

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
