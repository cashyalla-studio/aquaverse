package service

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"

	"github.com/cashyalla/aquaverse/internal/domain"
	"github.com/jmoiron/sqlx"
)

// TossPaymentsConfig 토스페이먼츠 설정
type TossPaymentsConfig struct {
	SecretKey string // 환경변수 TOSS_SECRET_KEY
	BaseURL   string // https://api.tosspayments.com/v1
	IsSandbox bool
}

type PaymentService struct {
	db         *sqlx.DB
	httpClient *http.Client
	cfg        TossPaymentsConfig
}

func NewPaymentService(db *sqlx.DB) *PaymentService {
	secretKey := os.Getenv("TOSS_SECRET_KEY")
	isSandbox := secretKey == "" || secretKey == "test"
	return &PaymentService{
		db:         db,
		httpClient: &http.Client{Timeout: 30 * time.Second},
		cfg: TossPaymentsConfig{
			SecretKey: secretKey,
			BaseURL:   "https://api.tosspayments.com/v1",
			IsSandbox: isSandbox,
		},
	}
}

// DB returns the underlying database connection.
func (s *PaymentService) DB() *sqlx.DB { return s.db }

// InitiatePayment 결제 초기화 → 결제창 URL 반환
func (s *PaymentService) InitiatePayment(ctx context.Context, tradeID int64, buyerID string) (*domain.PaymentInitResult, error) {
	// 에스크로 정보 조회
	var escrow struct {
		Amount   int64  `db:"amount"`
		Currency string `db:"currency"`
		Status   string `db:"status"`
	}
	if err := s.db.GetContext(ctx, &escrow,
		`SELECT amount, currency, status FROM escrow_transactions WHERE trade_id=$1`, tradeID); err != nil {
		return nil, fmt.Errorf("에스크로 정보 없음")
	}
	if escrow.Status != "PENDING" {
		return nil, fmt.Errorf("이미 처리된 거래입니다 (상태: %s)", escrow.Status)
	}

	orderID := fmt.Sprintf("AV-TRADE-%d-%d", tradeID, time.Now().Unix())

	// 샌드박스 모드: 실제 API 호출 없이 모의 결제창 URL 생성
	if s.cfg.IsSandbox || s.cfg.SecretKey == "" {
		checkoutURL := fmt.Sprintf(
			"https://sandbox.tosspayments.com/widget/sdk-v2/checkout?orderId=%s&amount=%d&orderName=AquaVerse%%20Trade%%20%%23%d",
			orderID, escrow.Amount, tradeID,
		)

		// DB에 pg_order_id, checkout_url 저장
		_, err := s.db.ExecContext(ctx, `
            UPDATE escrow_transactions
            SET pg_order_id=$1, pg_provider='toss', checkout_url=$2
            WHERE trade_id=$3
        `, orderID, checkoutURL, tradeID)
		if err != nil {
			return nil, err
		}

		return &domain.PaymentInitResult{
			TradeID:     tradeID,
			OrderID:     orderID,
			CheckoutURL: checkoutURL,
			Amount:      escrow.Amount,
			Currency:    escrow.Currency,
			IsSandbox:   true,
		}, nil
	}

	// 실제 토스페이먼츠 Payment Intent 생성
	reqBody := map[string]interface{}{
		"amount":    escrow.Amount,
		"orderId":   orderID,
		"orderName": fmt.Sprintf("AquaVerse 거래 #%d", tradeID),
		"currency":  "KRW",
		"method":    "카드",
	}
	bodyBytes, _ := json.Marshal(reqBody)

	req, _ := http.NewRequestWithContext(ctx, "POST",
		s.cfg.BaseURL+"/payments/key-in", bytes.NewReader(bodyBytes))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Basic "+base64.StdEncoding.EncodeToString([]byte(s.cfg.SecretKey+":")))

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)

	var result map[string]interface{}
	json.Unmarshal(body, &result) //nolint:errcheck

	checkoutURL := fmt.Sprintf("https://pay.toss.im/widget/v2?orderId=%s&amount=%d", orderID, escrow.Amount)
	if url, ok := result["checkoutUrl"].(string); ok {
		checkoutURL = url
	}

	rawJSON, _ := json.Marshal(result)
	_, err = s.db.ExecContext(ctx, `
        UPDATE escrow_transactions
        SET pg_order_id=$1, pg_provider='toss', checkout_url=$2, pg_raw_response=$3
        WHERE trade_id=$4
    `, orderID, checkoutURL, rawJSON, tradeID)
	if err != nil {
		return nil, err
	}

	return &domain.PaymentInitResult{
		TradeID:     tradeID,
		OrderID:     orderID,
		CheckoutURL: checkoutURL,
		Amount:      escrow.Amount,
		Currency:    escrow.Currency,
		IsSandbox:   false,
	}, nil
}

// ConfirmPayment 웹훅 수신 후 결제 확인 처리
func (s *PaymentService) ConfirmPayment(ctx context.Context, orderID, paymentKey, amount string) error {
	// 에스크로 조회
	var escrowID int64
	var tradeID int64
	if err := s.db.QueryRowContext(ctx,
		`SELECT id, trade_id FROM escrow_transactions WHERE pg_order_id=$1 AND status='PENDING'`,
		orderID).Scan(&escrowID, &tradeID); err != nil {
		return fmt.Errorf("주문 없음 또는 이미 처리됨: %s", orderID)
	}

	// 토스페이먼츠 결제 확인 API 호출 (샌드박스 모드는 스킵)
	if s.cfg.SecretKey != "" && s.cfg.SecretKey != "test" {
		reqBody := map[string]interface{}{
			"paymentKey": paymentKey,
			"orderId":    orderID,
			"amount":     amount,
		}
		bodyBytes, _ := json.Marshal(reqBody)
		req, _ := http.NewRequestWithContext(ctx, "POST",
			s.cfg.BaseURL+"/payments/confirm", bytes.NewReader(bodyBytes))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Basic "+base64.StdEncoding.EncodeToString([]byte(s.cfg.SecretKey+":")))
		resp, err := s.httpClient.Do(req)
		if err != nil {
			return err
		}
		defer resp.Body.Close()
		if resp.StatusCode != http.StatusOK {
			body, _ := io.ReadAll(resp.Body)
			return fmt.Errorf("결제 확인 실패: %s", string(body))
		}
	}

	// DB 업데이트: PENDING → FUNDED
	_, err := s.db.ExecContext(ctx, `
        UPDATE escrow_transactions
        SET status='FUNDED', pg_payment_key=$1, updated_at=NOW()
        WHERE id=$2
    `, paymentKey, escrowID)
	if err != nil {
		return err
	}

	// 거래 상태도 CONFIRMED로 업데이트
	_, err = s.db.ExecContext(ctx, `
        UPDATE trades SET status='CONFIRMED', updated_at=NOW() WHERE id=$1
    `, tradeID)
	return err
}
