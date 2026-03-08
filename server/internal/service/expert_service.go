package service

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"

	"github.com/cashyalla/aquaverse/internal/domain"
	"github.com/jmoiron/sqlx"
)

// ExpertService Expert Connect 기능을 제공한다.
type ExpertService struct {
	db *sqlx.DB
}

// NewExpertService ExpertService를 생성한다.
func NewExpertService(db *sqlx.DB) *ExpertService {
	return &ExpertService{db: db}
}

// ListExperts 전문가 목록 조회 (타입 필터 + 페이지네이션)
func (s *ExpertService) ListExperts(ctx context.Context, expertType string, page, limit int) ([]domain.ExpertProfile, int, error) {
	if limit <= 0 || limit > 50 {
		limit = 20
	}
	if page <= 0 {
		page = 1
	}
	offset := (page - 1) * limit

	var experts []domain.ExpertProfile
	var total int
	var err error

	if expertType != "" {
		err = s.db.SelectContext(ctx, &experts, `
			SELECT ep.*, u.username
			FROM expert_profiles ep
			JOIN users u ON u.id = ep.user_id
			WHERE ep.expert_type = $1
			ORDER BY ep.is_verified DESC, ep.rating DESC NULLS LAST, ep.review_count DESC
			LIMIT $2 OFFSET $3
		`, expertType, limit, offset)
		if err != nil {
			return nil, 0, err
		}
		s.db.QueryRowContext(ctx,
			`SELECT COUNT(*) FROM expert_profiles WHERE expert_type = $1`, expertType,
		).Scan(&total)
	} else {
		err = s.db.SelectContext(ctx, &experts, `
			SELECT ep.*, u.username
			FROM expert_profiles ep
			JOIN users u ON u.id = ep.user_id
			ORDER BY ep.is_verified DESC, ep.rating DESC NULLS LAST, ep.review_count DESC
			LIMIT $1 OFFSET $2
		`, limit, offset)
		if err != nil {
			return nil, 0, err
		}
		s.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM expert_profiles`).Scan(&total)
	}

	// JSONB specialties 언마샬
	for i := range experts {
		experts[i].UnmarshalSpecialties()
	}

	return experts, total, nil
}

// GetExpert 전문가 상세 조회
func (s *ExpertService) GetExpert(ctx context.Context, id int64) (*domain.ExpertProfile, error) {
	var ep domain.ExpertProfile
	err := s.db.QueryRowContext(ctx, `
		SELECT ep.id, ep.user_id, ep.expert_type, ep.bio, ep.specialties, ep.hourly_rate,
		       ep.is_verified, ep.verified_at, ep.rating, ep.review_count, ep.created_at,
		       u.username
		FROM expert_profiles ep
		JOIN users u ON u.id = ep.user_id
		WHERE ep.id = $1
	`, id).Scan(
		&ep.ID, &ep.UserID, &ep.ExpertType, &ep.Bio, &ep.SpecialtiesRaw, &ep.HourlyRate,
		&ep.IsVerified, &ep.VerifiedAt, &ep.Rating, &ep.ReviewCount, &ep.CreatedAt,
		&ep.Username,
	)
	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("전문가를 찾을 수 없습니다")
	}
	if err != nil {
		return nil, err
	}

	ep.UnmarshalSpecialties()
	return &ep, nil
}

// UpsertProfile 내 전문가 프로필 등록 또는 수정 (ON CONFLICT)
func (s *ExpertService) UpsertProfile(ctx context.Context, userID string, req domain.ExpertProfileRequest) error {
	if req.ExpertType == "" {
		return fmt.Errorf("expert_type 필수")
	}
	validTypes := map[string]bool{"vet": true, "breeder": true, "aquarist": true, "trainer": true}
	if !validTypes[req.ExpertType] {
		return fmt.Errorf("expert_type은 vet, breeder, aquarist, trainer 중 하나여야 합니다")
	}

	specialtiesJSON, err := json.Marshal(req.Specialties)
	if err != nil {
		specialtiesJSON = []byte("[]")
	}

	_, err = s.db.ExecContext(ctx, `
		INSERT INTO expert_profiles (user_id, expert_type, bio, specialties, hourly_rate)
		VALUES ($1::uuid, $2, $3, $4, $5)
		ON CONFLICT (user_id) DO UPDATE
		SET expert_type = EXCLUDED.expert_type,
		    bio         = EXCLUDED.bio,
		    specialties = EXCLUDED.specialties,
		    hourly_rate = EXCLUDED.hourly_rate
	`, userID, req.ExpertType, req.Bio, specialtiesJSON, req.HourlyRate)
	return err
}

// CreateConsultation 상담 예약 생성
func (s *ExpertService) CreateConsultation(ctx context.Context, userID string, expertID int64, req domain.ConsultationRequest) (*domain.Consultation, error) {
	// 전문가 존재 확인
	var exists bool
	s.db.QueryRowContext(ctx, `SELECT EXISTS(SELECT 1 FROM expert_profiles WHERE id = $1)`, expertID).Scan(&exists)
	if !exists {
		return nil, fmt.Errorf("전문가를 찾을 수 없습니다")
	}

	// 자기 자신에게 예약 불가
	var expertUserID string
	s.db.QueryRowContext(ctx, `SELECT user_id FROM expert_profiles WHERE id = $1`, expertID).Scan(&expertUserID)
	if expertUserID == userID {
		return nil, fmt.Errorf("본인에게 상담을 예약할 수 없습니다")
	}

	durationMin := req.DurationMin
	if durationMin <= 0 {
		durationMin = 30
	}

	// 전문가의 hourly_rate 기반 결제금액 계산
	var hourlyRate *int64
	s.db.QueryRowContext(ctx, `SELECT hourly_rate FROM expert_profiles WHERE id = $1`, expertID).Scan(&hourlyRate)
	var paymentAmount *int64
	if hourlyRate != nil {
		amt := *hourlyRate * int64(durationMin) / 60
		paymentAmount = &amt
	}

	var c domain.Consultation
	err := s.db.QueryRowContext(ctx, `
		INSERT INTO consultations (user_id, expert_id, scheduled_at, duration_min, question, payment_amount)
		VALUES ($1::uuid, $2, $3, $4, $5, $6)
		RETURNING id, user_id, expert_id, scheduled_at, duration_min, status, question, answer, payment_amount, created_at
	`, userID, expertID, req.ScheduledAt, durationMin, req.Question, paymentAmount,
	).Scan(
		&c.ID, &c.UserID, &c.ExpertID, &c.ScheduledAt, &c.DurationMin,
		&c.Status, &c.Question, &c.Answer, &c.PaymentAmount, &c.CreatedAt,
	)
	if err != nil {
		return nil, err
	}
	return &c, nil
}

// GetMyConsultations 내 상담 목록 조회
func (s *ExpertService) GetMyConsultations(ctx context.Context, userID string) ([]domain.Consultation, error) {
	var consultations []domain.Consultation
	err := s.db.SelectContext(ctx, &consultations, `
		SELECT c.id, c.user_id, c.expert_id, c.scheduled_at, c.duration_min,
		       c.status, c.question, c.answer, c.payment_amount, c.created_at,
		       u.username as expert_username
		FROM consultations c
		JOIN expert_profiles ep ON ep.id = c.expert_id
		JOIN users u ON u.id = ep.user_id
		WHERE c.user_id = $1::uuid
		ORDER BY c.created_at DESC
		LIMIT 50
	`, userID)
	return consultations, err
}

// UpdateConsultationStatus 상담 상태 변경
func (s *ExpertService) UpdateConsultationStatus(ctx context.Context, id int64, userID, status string) error {
	validStatuses := map[string]bool{
		"confirmed": true, "completed": true, "cancelled": true,
	}
	if !validStatuses[status] {
		return fmt.Errorf("유효하지 않은 상태: %s", status)
	}

	// 해당 상담의 전문가 또는 상담 신청자만 변경 가능
	result, err := s.db.ExecContext(ctx, `
		UPDATE consultations SET status = $1
		WHERE id = $2
		  AND (
		      user_id = $3::uuid
		      OR expert_id IN (SELECT id FROM expert_profiles WHERE user_id = $3::uuid)
		  )
	`, status, id, userID)
	if err != nil {
		return err
	}
	n, _ := result.RowsAffected()
	if n == 0 {
		return fmt.Errorf("상담을 찾을 수 없거나 권한이 없습니다")
	}
	return nil
}

// CreateReview 상담 완료 후 리뷰 작성
func (s *ExpertService) CreateReview(ctx context.Context, consultationID int64, reviewerID string, rating int, comment string) error {
	if rating < 1 || rating > 5 {
		return fmt.Errorf("rating은 1-5 사이여야 합니다")
	}

	// 완료된 상담인지, 본인 상담인지 확인
	var expertID int64
	var status string
	err := s.db.QueryRowContext(ctx, `
		SELECT expert_id, status FROM consultations
		WHERE id = $1 AND user_id = $2::uuid
	`, consultationID, reviewerID).Scan(&expertID, &status)
	if err == sql.ErrNoRows {
		return fmt.Errorf("상담을 찾을 수 없거나 권한이 없습니다")
	}
	if err != nil {
		return err
	}
	if status != "completed" {
		return fmt.Errorf("완료된 상담에만 리뷰를 작성할 수 있습니다")
	}

	// 이미 리뷰가 있는지 확인
	var alreadyReviewed bool
	s.db.QueryRowContext(ctx,
		`SELECT EXISTS(SELECT 1 FROM expert_reviews WHERE consultation_id = $1)`, consultationID,
	).Scan(&alreadyReviewed)
	if alreadyReviewed {
		return fmt.Errorf("이미 리뷰를 작성했습니다")
	}

	tx, err := s.db.BeginTxx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	// 리뷰 삽입
	var commentPtr *string
	if comment != "" {
		commentPtr = &comment
	}
	_, err = tx.ExecContext(ctx, `
		INSERT INTO expert_reviews (consultation_id, reviewer_id, expert_id, rating, comment)
		VALUES ($1, $2::uuid, $3, $4, $5)
	`, consultationID, reviewerID, expertID, rating, commentPtr)
	if err != nil {
		return err
	}

	// 전문가 프로필 rating 및 review_count 갱신
	_, err = tx.ExecContext(ctx, `
		UPDATE expert_profiles
		SET review_count = review_count + 1,
		    rating = (
		        SELECT ROUND(AVG(r.rating)::numeric, 2)
		        FROM expert_reviews r
		        WHERE r.expert_id = $1
		    )
		WHERE id = $1
	`, expertID)
	if err != nil {
		return err
	}

	return tx.Commit()
}
