package notification

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"time"

	"github.com/cashyalla/aquaverse/internal/domain"
)

// Service 알림 발송 서비스 (FCM 푸시 + 인앱)
type Service struct {
	fcmServerKey string
	client       *http.Client
	logger       *slog.Logger
	cache        NotifyQueue
}

type NotifyQueue interface {
	PushNotifyJob(ctx context.Context, job []byte) error
	PopNotifyJob(ctx context.Context, timeout time.Duration) ([]byte, error)
}

type NotifyJob struct {
	UserID           string    `json:"user_id"`
	SubscriptionID   *int64    `json:"subscription_id,omitempty"`
	ListingID        *int64    `json:"listing_id,omitempty"`
	NotificationType string    `json:"notification_type"`
	Channel          string    `json:"channel"` // PUSH, IN_APP
	Title            string    `json:"title"`
	Body             string    `json:"body"`
	CreatedAt        time.Time `json:"created_at"`
}

func NewService(fcmKey string, cache NotifyQueue, logger *slog.Logger) *Service {
	return &Service{
		fcmServerKey: fcmKey,
		client:       &http.Client{Timeout: 10 * time.Second},
		logger:       logger,
		cache:        cache,
	}
}

// NotifyNewListing 신규 분양글 매칭 구독자에게 알림 (비동기)
func (s *Service) NotifyNewListing(ctx context.Context, listing *domain.Listing) {
	// 실제로는 DB에서 매칭 구독자 조회 후 큐에 넣음
	// 여기서는 큐 푸시 패턴만 보여줌
	job := NotifyJob{
		ListingID:        &listing.ID,
		NotificationType: "FISH_WATCH",
		Channel:          "PUSH",
		Title:            fmt.Sprintf("New listing: %s", listing.CommonName),
		Body:             fmt.Sprintf("%s is now available - %s", listing.CommonName, listing.LocationText),
		CreatedAt:        time.Now(),
	}
	b, _ := json.Marshal(job)
	if err := s.cache.PushNotifyJob(ctx, b); err != nil {
		s.logger.Error("failed to push notify job", "err", err)
	}
}

// NotifyTradeUpdate 거래 상태 변경 알림
func (s *Service) NotifyTradeUpdate(ctx context.Context, trade *domain.Trade, targetUserID string) {
	statusMessages := map[domain.TradeStatus]string{
		domain.TradeStatusConfirmed:   "Trade confirmed! Please proceed with the exchange.",
		domain.TradeStatusInDelivery:  "Your fish is on the way!",
		domain.TradeStatusDelivered:   "Fish arrived - please confirm receipt.",
		domain.TradeStatusCompleted:   "Trade completed successfully!",
		domain.TradeStatusCancelled:   "Trade has been cancelled.",
		domain.TradeStatusDisputed:    "A dispute has been filed. Admin will review.",
	}

	msg, ok := statusMessages[trade.Status]
	if !ok {
		return
	}

	job := NotifyJob{
		UserID:           targetUserID,
		NotificationType: "TRADE_UPDATE",
		Channel:          "PUSH",
		Title:            "Trade Update",
		Body:             msg,
		CreatedAt:        time.Now(),
	}
	b, _ := json.Marshal(job)
	_ = s.cache.PushNotifyJob(ctx, b)
}

// StartWorker 알림 큐 워커 (고루틴으로 실행)
func (s *Service) StartWorker(ctx context.Context) {
	go func() {
		for {
			select {
			case <-ctx.Done():
				s.logger.Info("notification worker stopped")
				return
			default:
				data, err := s.cache.PopNotifyJob(ctx, 5*time.Second)
				if err != nil || data == nil {
					continue
				}
				var job NotifyJob
				if err := json.Unmarshal(data, &job); err != nil {
					s.logger.Error("invalid notify job", "err", err)
					continue
				}
				s.dispatch(ctx, job)
			}
		}
	}()
}

func (s *Service) dispatch(ctx context.Context, job NotifyJob) {
	switch job.Channel {
	case "PUSH":
		s.sendFCM(ctx, job)
	case "IN_APP":
		// DB notification_log 저장 (여기선 생략)
		s.logger.Info("in-app notification", "user", job.UserID, "title", job.Title)
	}
}

// sendFCM FCM v1 API 푸시 발송
func (s *Service) sendFCM(ctx context.Context, job NotifyJob) {
	if s.fcmServerKey == "" || job.UserID == "" {
		return
	}

	payload := map[string]interface{}{
		"message": map[string]interface{}{
			"token": job.UserID, // 실제로는 FCM 토큰 (users_fcm_tokens 테이블 필요)
			"notification": map[string]string{
				"title": job.Title,
				"body":  job.Body,
			},
			"data": map[string]string{
				"type":       job.NotificationType,
				"listing_id": fmt.Sprintf("%v", job.ListingID),
			},
		},
	}

	b, _ := json.Marshal(payload)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost,
		"https://fcm.googleapis.com/v1/projects/aquaverse/messages:send",
		bytes.NewReader(b),
	)
	if err != nil {
		return
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+s.fcmServerKey)

	resp, err := s.client.Do(req)
	if err != nil {
		s.logger.Error("FCM send failed", "err", err)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		s.logger.Warn("FCM non-200", "status", resp.StatusCode, "user", job.UserID)
	}
}
