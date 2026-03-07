package domain

import "time"

type CitesAppendix string

const (
	CitesAppendixI   CitesAppendix = "I"
	CitesAppendixII  CitesAppendix = "II"
	CitesAppendixIII CitesAppendix = "III"
)

type CitesFish struct {
	ID             int64         `db:"id"`
	ScientificName string        `db:"scientific_name"`
	CommonNames    []string      `db:"common_names"`
	Appendix       CitesAppendix `db:"appendix"`
	IsBlocked      bool          `db:"is_blocked"` // true: 등록 차단, false: 경고만
	Notes          *string       `db:"notes"`
	CreatedAt      time.Time     `db:"created_at"`
}

type InvasiveSpeciesKR struct {
	ID             int64     `db:"id"`
	ScientificName string    `db:"scientific_name"`
	KoreanName     *string   `db:"korean_name"`
	CommonNames    []string  `db:"common_names"`
	Notes          *string   `db:"notes"`
	CreatedAt      time.Time `db:"created_at"`
}

// CitesCheckResult: 어종 등록 시 CITES/교란종 체크 결과
type CitesCheckResult struct {
	IsBlocked    bool          `json:"is_blocked"`              // 차단 여부
	HasWarning   bool          `json:"has_warning"`             // 경고 여부
	Appendix     CitesAppendix `json:"appendix,omitempty"`
	Message      string        `json:"message,omitempty"`
	IsInvasiveKR bool          `json:"is_invasive_kr"` // 한국 교란종 여부
}
