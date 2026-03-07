package notification

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"

	"github.com/jmoiron/sqlx"
)

// FCMService Firebase Cloud Messaging v1 HTTP API 클라이언트
type FCMService struct {
	db         *sqlx.DB
	httpClient *http.Client
	projectID  string
	serverKey  string // FCM Legacy (v1 전환 전 임시 사용)
}

func NewFCMService(db *sqlx.DB) *FCMService {
	return &FCMService{
		db:         db,
		httpClient: &http.Client{Timeout: 10 * time.Second},
		projectID:  os.Getenv("FIREBASE_PROJECT_ID"),
		serverKey:  os.Getenv("FCM_SERVER_KEY"), // FCM v1: OAuth2 필요, 임시로 legacy key 사용
	}
}

// NotifyPayload 알림 페이로드
type NotifyPayload struct {
	Title string
	Body  string
	Data  map[string]string
}

// SendToUser 특정 사용자의 모든 기기에 알림 전송
func (s *FCMService) SendToUser(ctx context.Context, userID string, payload NotifyPayload) error {
	if s.serverKey == "" && s.projectID == "" {
		// FCM 설정 없으면 로그만
		return nil
	}

	// 사용자 FCM 토큰 조회
	var tokens []string
	if err := s.db.SelectContext(ctx, &tokens,
		`SELECT token FROM users_fcm_tokens WHERE user_id=$1`, userID); err != nil || len(tokens) == 0 {
		return nil // 토큰 없으면 스킵 (앱 미설치)
	}

	for _, token := range tokens {
		if err := s.sendFCM(ctx, token, payload); err != nil {
			// 개별 전송 실패는 무시 (토큰 만료 등)
			continue
		}
	}

	// 알림 이력 저장
	dataJSON, _ := json.Marshal(payload.Data)
	s.db.ExecContext(ctx, `
		INSERT INTO notification_logs (user_id, title, body, data, success)
		VALUES ($1,$2,$3,$4,true)
	`, userID, payload.Title, payload.Body, dataJSON)

	return nil
}

func (s *FCMService) sendFCM(ctx context.Context, token string, payload NotifyPayload) error {
	// FCM Legacy HTTP API (서버키 기반)
	// v1 API: https://fcm.googleapis.com/v1/projects/{project}/messages:send (OAuth2 필요)
	// 임시: Legacy HTTP API 사용
	msg := map[string]interface{}{
		"to": token,
		"notification": map[string]string{
			"title": payload.Title,
			"body":  payload.Body,
		},
	}
	if len(payload.Data) > 0 {
		msg["data"] = payload.Data
	}

	bodyBytes, _ := json.Marshal(msg)
	req, err := http.NewRequestWithContext(ctx, "POST",
		"https://fcm.googleapis.com/fcm/send", bytes.NewReader(bodyBytes))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "key="+s.serverKey)

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("FCM 오류: %d %s", resp.StatusCode, string(body))
	}
	return nil
}

// RegisterToken FCM 토큰 등록/갱신
func (s *FCMService) RegisterToken(ctx context.Context, userID, token, platform string) error {
	_, err := s.db.ExecContext(ctx, `
		INSERT INTO users_fcm_tokens (user_id, token, platform, updated_at)
		VALUES ($1,$2,$3,NOW())
		ON CONFLICT (token) DO UPDATE SET user_id=$1, platform=$3, updated_at=NOW()
	`, userID, token, platform)
	return err
}

// UnregisterToken FCM 토큰 제거
func (s *FCMService) UnregisterToken(ctx context.Context, userID, token string) error {
	_, err := s.db.ExecContext(ctx, `
		DELETE FROM users_fcm_tokens WHERE user_id=$1 AND token=$2
	`, userID, token)
	return err
}

// --- 알림 트리거 헬퍼 ---

// NotifyChatMessage 채팅 메시지 알림 (수신자에게)
func (s *FCMService) NotifyChatMessage(ctx context.Context, recipientID, senderName, message string, tradeID int64) {
	go s.SendToUser(context.Background(), recipientID, NotifyPayload{
		Title: fmt.Sprintf("💬 %s", senderName),
		Body:  truncate(message, 100),
		Data: map[string]string{
			"type":     "chat",
			"trade_id": fmt.Sprintf("%d", tradeID),
			"url":      fmt.Sprintf("/trades/%d/chat", tradeID),
		},
	})
}

// NotifyEscrowStatus 에스크로 상태 변경 알림
func (s *FCMService) NotifyEscrowStatus(ctx context.Context, userID, status string, tradeID int64) {
	var title, body string
	switch status {
	case "FUNDED":
		title = "💰 에스크로 입금 완료"
		body = "거래 상대방이 에스크로에 입금했습니다. 어종을 전달해주세요."
	case "RELEASED":
		title = "✅ 거래 완료"
		body = "구매자가 수령을 확인했습니다. 대금이 지급됩니다."
	case "REFUNDED":
		title = "↩️ 환불 처리"
		body = "에스크로 금액이 환불됐습니다."
	case "DISPUTED":
		title = "⚠️ 분쟁 신청"
		body = "거래에 분쟁이 접수됐습니다."
	default:
		return
	}
	go s.SendToUser(context.Background(), userID, NotifyPayload{
		Title: title, Body: body,
		Data: map[string]string{"type": "escrow", "trade_id": fmt.Sprintf("%d", tradeID)},
	})
}

func truncate(s string, n int) string {
	runes := []rune(s)
	if len(runes) > n {
		return string(runes[:n]) + "..."
	}
	return s
}
