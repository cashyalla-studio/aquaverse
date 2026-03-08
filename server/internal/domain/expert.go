package domain

import (
	"encoding/json"
	"time"
)

// ExpertType 전문가 유형
type ExpertType string

const (
	ExpertTypeVet      ExpertType = "vet"
	ExpertTypeBreeder  ExpertType = "breeder"
	ExpertTypeAquarist ExpertType = "aquarist"
	ExpertTypeTrainer  ExpertType = "trainer"
)

// ConsultationStatus 상담 상태
type ConsultationStatus string

const (
	ConsultationStatusPending   ConsultationStatus = "pending"
	ConsultationStatusConfirmed ConsultationStatus = "confirmed"
	ConsultationStatusCompleted ConsultationStatus = "completed"
	ConsultationStatusCancelled ConsultationStatus = "cancelled"
)

// ExpertProfile 전문가 프로필
type ExpertProfile struct {
	ID             int64           `db:"id" json:"id"`
	UserID         string          `db:"user_id" json:"user_id"`
	ExpertType     string          `db:"expert_type" json:"expert_type"`
	Bio            *string         `db:"bio" json:"bio,omitempty"`
	SpecialtiesRaw json.RawMessage `db:"specialties" json:"-"`
	Specialties    []string        `db:"-" json:"specialties,omitempty"`
	HourlyRate     *int64          `db:"hourly_rate" json:"hourly_rate,omitempty"`
	IsVerified     bool            `db:"is_verified" json:"is_verified"`
	VerifiedAt     *time.Time      `db:"verified_at" json:"verified_at,omitempty"`
	Rating         *float64        `db:"rating" json:"rating,omitempty"`
	ReviewCount    int             `db:"review_count" json:"review_count"`
	CreatedAt      time.Time       `db:"created_at" json:"created_at"`
	// 조인 필드
	Username string `db:"username" json:"username,omitempty"`
}

// UnmarshalSpecialties SpecialtiesRaw (JSONB)를 Specialties []string으로 파싱한다.
func (ep *ExpertProfile) UnmarshalSpecialties() {
	if len(ep.SpecialtiesRaw) > 0 {
		_ = json.Unmarshal(ep.SpecialtiesRaw, &ep.Specialties)
	}
}

// ExpertProfileRequest 프로필 등록/수정 요청
type ExpertProfileRequest struct {
	ExpertType  string   `json:"expert_type"`
	Bio         string   `json:"bio"`
	Specialties []string `json:"specialties"`
	HourlyRate  *int64   `json:"hourly_rate"`
}

// Consultation 상담 예약
type Consultation struct {
	ID            int64      `db:"id" json:"id"`
	UserID        string     `db:"user_id" json:"user_id"`
	ExpertID      int64      `db:"expert_id" json:"expert_id"`
	ScheduledAt   *time.Time `db:"scheduled_at" json:"scheduled_at,omitempty"`
	DurationMin   int        `db:"duration_min" json:"duration_min"`
	Status        string     `db:"status" json:"status"`
	Question      *string    `db:"question" json:"question,omitempty"`
	Answer        *string    `db:"answer" json:"answer,omitempty"`
	PaymentAmount *int64     `db:"payment_amount" json:"payment_amount,omitempty"`
	CreatedAt     time.Time  `db:"created_at" json:"created_at"`
	// 조인 필드
	ExpertUsername string `db:"expert_username" json:"expert_username,omitempty"`
}

// ConsultationRequest 상담 예약 요청
type ConsultationRequest struct {
	ScheduledAt *time.Time `json:"scheduled_at"`
	DurationMin int        `json:"duration_min"`
	Question    string     `json:"question"`
}

// ExpertReview 전문가 리뷰
type ExpertReview struct {
	ID             int64     `db:"id" json:"id"`
	ConsultationID int64     `db:"consultation_id" json:"consultation_id"`
	ReviewerID     string    `db:"reviewer_id" json:"reviewer_id"`
	ExpertID       int64     `db:"expert_id" json:"expert_id"`
	Rating         int       `db:"rating" json:"rating"`
	Comment        *string   `db:"comment" json:"comment,omitempty"`
	CreatedAt      time.Time `db:"created_at" json:"created_at"`
}

// ShopInventory 업체 재고 항목
type ShopInventory struct {
	ID          int64      `db:"id" json:"id"`
	BusinessID  int64      `db:"business_id" json:"business_id"`
	FishDataID  *int64     `db:"fish_data_id" json:"fish_data_id,omitempty"`
	CustomName  *string    `db:"custom_name" json:"custom_name,omitempty"`
	Quantity    int        `db:"quantity" json:"quantity"`
	Price       *int64     `db:"price" json:"price,omitempty"`
	CitesStatus *string    `db:"cites_status" json:"cites_status,omitempty"`
	IsAvailable bool       `db:"is_available" json:"is_available"`
	UpdatedAt   time.Time  `db:"updated_at" json:"updated_at"`
	// 조인 필드
	FishName string `db:"fish_name" json:"fish_name,omitempty"`
}

// ShopInventoryRequest 재고 추가/수정 요청
type ShopInventoryRequest struct {
	FishDataID  *int64  `json:"fish_data_id"`
	CustomName  string  `json:"custom_name"`
	Quantity    int     `json:"quantity"`
	Price       *int64  `json:"price"`
	IsAvailable *bool   `json:"is_available"`
}

// BusinessStats 업체 통계
type BusinessStats struct {
	BusinessID     int64 `json:"business_id"`
	TotalInventory int   `json:"total_inventory"`
	AvailableItems int   `json:"available_items"`
	TotalValue     int64 `json:"total_value"`
	ReviewCount    int   `json:"review_count"`
	AvgRating      float64 `json:"avg_rating"`
}
