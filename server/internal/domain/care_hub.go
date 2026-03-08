package domain

import (
	"time"

	"github.com/google/uuid"
)

// CareSchedule 케어 일정
type CareSchedule struct {
	ID           int64      `db:"id"            json:"id"`
	TankID       int64      `db:"tank_id"       json:"tank_id"`
	UserID       uuid.UUID  `db:"user_id"       json:"user_id"`
	ScheduleType string     `db:"schedule_type" json:"schedule_type"`
	Title        string     `db:"title"         json:"title"`
	Description  *string    `db:"description"   json:"description,omitempty"`
	Frequency    string     `db:"frequency"     json:"frequency"`
	IntervalDays *int       `db:"interval_days" json:"interval_days,omitempty"`
	NextDueAt    time.Time  `db:"next_due_at"   json:"next_due_at"`
	LastDoneAt   *time.Time `db:"last_done_at"  json:"last_done_at,omitempty"`
	IsActive     bool       `db:"is_active"     json:"is_active"`
	CreatedAt    time.Time  `db:"created_at"    json:"created_at"`
}

// CareLog 케어 기록
type CareLog struct {
	ID         int64      `db:"id"          json:"id"`
	ScheduleID *int64     `db:"schedule_id" json:"schedule_id,omitempty"`
	TankID     int64      `db:"tank_id"     json:"tank_id"`
	UserID     uuid.UUID  `db:"user_id"     json:"user_id"`
	CareType   string     `db:"care_type"   json:"care_type"`
	Notes      *string    `db:"notes"       json:"notes,omitempty"`
	PhotoURL   *string    `db:"photo_url"   json:"photo_url,omitempty"`
	DoneAt     time.Time  `db:"done_at"     json:"done_at"`
}

// CareStreak 케어 스트릭
type CareStreak struct {
	UserID        uuid.UUID  `db:"user_id"        json:"user_id"`
	CurrentStreak int        `db:"current_streak" json:"current_streak"`
	LongestStreak int        `db:"longest_streak" json:"longest_streak"`
	LastCareDate  *string    `db:"last_care_date" json:"last_care_date,omitempty"`
	UpdatedAt     time.Time  `db:"updated_at"     json:"updated_at"`
}

// CreateScheduleRequest 일정 생성 요청
type CreateScheduleRequest struct {
	ScheduleType string  `json:"schedule_type"`
	Title        string  `json:"title"`
	Description  *string `json:"description,omitempty"`
	Frequency    string  `json:"frequency"`
	IntervalDays *int    `json:"interval_days,omitempty"`
	NextDueAt    string  `json:"next_due_at"` // RFC3339
}

// UpdateScheduleRequest 일정 수정 요청
type UpdateScheduleRequest struct {
	Title        *string `json:"title,omitempty"`
	Description  *string `json:"description,omitempty"`
	Frequency    *string `json:"frequency,omitempty"`
	IntervalDays *int    `json:"interval_days,omitempty"`
	NextDueAt    *string `json:"next_due_at,omitempty"` // RFC3339
	IsActive     *bool   `json:"is_active,omitempty"`
}

// CompleteScheduleRequest 케어 완료 요청
type CompleteScheduleRequest struct {
	Notes    *string `json:"notes,omitempty"`
	PhotoURL *string `json:"photo_url,omitempty"`
}
