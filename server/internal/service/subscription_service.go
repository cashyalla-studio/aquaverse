package service

import (
	"context"
	"fmt"
	"time"

	"github.com/cashyalla/aquaverse/internal/domain"
	"github.com/jmoiron/sqlx"
)

type SubscriptionService struct {
	db *sqlx.DB
}

func NewSubscriptionService(db *sqlx.DB) *SubscriptionService {
	return &SubscriptionService{db: db}
}

// GetSubscription 사용자 구독 정보 조회 (없으면 FREE 반환)
func (s *SubscriptionService) GetSubscription(ctx context.Context, userID string) (*domain.Subscription, error) {
	var sub domain.Subscription
	err := s.db.GetContext(ctx, &sub, `
		SELECT id, user_id, plan, status, started_at, expires_at, billing_amount, created_at
		FROM user_subscriptions
		WHERE user_id=$1::uuid
	`, userID)
	if err != nil {
		// 구독 정보 없으면 FREE 반환
		return &domain.Subscription{
			UserID: userID,
			Plan:   "FREE",
			Status: "ACTIVE",
		}, nil
	}
	// 만료 체크
	if sub.ExpiresAt != nil {
		expTime, _ := time.Parse(time.RFC3339, *sub.ExpiresAt)
		if time.Now().After(expTime) {
			sub.Plan = "FREE"
			sub.Status = "EXPIRED"
		}
	}
	return &sub, nil
}

// IsPro 사용자가 PRO 구독 중인지 확인
func (s *SubscriptionService) IsPro(ctx context.Context, userID string) bool {
	var plan string
	var expiresAt *time.Time
	err := s.db.QueryRowContext(ctx, `
		SELECT plan, expires_at FROM user_subscriptions
		WHERE user_id=$1::uuid AND status='ACTIVE'
	`, userID).Scan(&plan, &expiresAt)
	if err != nil || plan != "PRO" {
		return false
	}
	if expiresAt != nil && time.Now().After(*expiresAt) {
		return false
	}
	return true
}

// Subscribe PRO 구독 활성화 (토스페이먼츠 빌링키 기반)
func (s *SubscriptionService) Subscribe(ctx context.Context, userID, billingKey string) (*domain.Subscription, error) {
	expiresAt := time.Now().AddDate(0, 1, 0) // 1개월 후 만료
	expiresAtStr := expiresAt.Format(time.RFC3339)

	_, err := s.db.ExecContext(ctx, `
		INSERT INTO user_subscriptions (user_id, plan, status, expires_at, billing_key, billing_amount)
		VALUES ($1::uuid, 'PRO', 'ACTIVE', $2, $3, 9900)
		ON CONFLICT (user_id) DO UPDATE
		SET plan='PRO', status='ACTIVE', expires_at=$2, billing_key=$3, billing_amount=9900, updated_at=NOW()
	`, userID, expiresAtStr, billingKey)
	if err != nil {
		return nil, err
	}

	// 이력 저장
	s.db.ExecContext(ctx, `
		INSERT INTO subscription_history (user_id, plan, event, amount)
		VALUES ($1::uuid, 'PRO', 'SUBSCRIBED', 9900)
	`, userID)

	return &domain.Subscription{
		UserID:        userID,
		Plan:          "PRO",
		Status:        "ACTIVE",
		ExpiresAt:     &expiresAtStr,
		BillingAmount: 9900,
	}, nil
}

// SubscribeFree 무료 체험 (빌링키 없이 1개월 PRO)
func (s *SubscriptionService) SubscribeFree(ctx context.Context, userID string) (*domain.Subscription, error) {
	return s.Subscribe(ctx, userID, "")
}

// Cancel 구독 취소
func (s *SubscriptionService) Cancel(ctx context.Context, userID string) error {
	result, err := s.db.ExecContext(ctx, `
		UPDATE user_subscriptions SET status='CANCELLED', updated_at=NOW()
		WHERE user_id=$1::uuid AND status='ACTIVE'
	`, userID)
	if err != nil {
		return err
	}
	n, _ := result.RowsAffected()
	if n == 0 {
		return fmt.Errorf("활성 구독이 없습니다")
	}
	s.db.ExecContext(ctx, `
		INSERT INTO subscription_history (user_id, plan, event)
		VALUES ($1::uuid, 'PRO', 'CANCELLED')
	`, userID)
	return nil
}

// GetPlans 플랜 목록
func (s *SubscriptionService) GetPlans() []domain.SubscriptionPlan {
	return domain.PredefinedPlans
}
