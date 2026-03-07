package domain

// WaterParams 수질 파라미터
type WaterParams struct {
	ID         int64    `db:"id" json:"id"`
	TankID     int64    `db:"tank_id" json:"tank_id"`
	TempC      *float64 `db:"temp_c" json:"temp_c,omitempty"`
	PH         *float64 `db:"ph" json:"ph,omitempty"`
	AmmoniaPPM *float64 `db:"ammonia_ppm" json:"ammonia_ppm,omitempty"`
	NitritePPM *float64 `db:"nitrite_ppm" json:"nitrite_ppm,omitempty"`
	NitratePPM *float64 `db:"nitrate_ppm" json:"nitrate_ppm,omitempty"`
	GhDgh      *float64 `db:"gh_dgh" json:"gh_dgh,omitempty"`
	KhDkh      *float64 `db:"kh_dkh" json:"kh_dkh,omitempty"`
	RecordedAt string   `db:"recorded_at" json:"recorded_at"`
	Notes      string   `db:"notes" json:"notes,omitempty"`
}

// FishStatus 어종별 수질 상태
type FishStatus struct {
	FishName   string `json:"fish_name"`
	Status     string `json:"status"` // "good", "warning", "danger"
	Issue      string `json:"issue,omitempty"`
	Suggestion string `json:"suggestion,omitempty"`
}

// TankDiagnosis AI 진단 결과
type TankDiagnosis struct {
	TankID     int64        `json:"tank_id"`
	Summary    string       `json:"summary"`
	FishStates []FishStatus `json:"fish_states"`
	Actions    []string     `json:"actions"`
	CreatedAt  string       `json:"created_at"`
}
