package service

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/jmoiron/sqlx"
)

// PaymentServiceV2 다중 PSP를 지원하는 결제 서비스.
// 기존 PaymentService를 교체하지 않고 점진적 마이그레이션을 위해 별도로 제공한다.
// PSP 선택 로직: currency가 "KRW"이면 Toss, 그 외(USD 등)이면 Stripe.
type PaymentServiceV2 struct {
	db        *sqlx.DB
	providers map[string]PaymentProvider
}

// NewPaymentServiceV2 PaymentServiceV2를 생성한다.
// tossKey: TOSS_SECRET_KEY 환경변수 값
// stripeKey: STRIPE_SECRET_KEY 환경변수 값
// stripeWebhookSecret: STRIPE_WEBHOOK_SECRET 환경변수 값
func NewPaymentServiceV2(db *sqlx.DB, tossKey, stripeKey, stripeWebhookSecret string) *PaymentServiceV2 {
	providers := map[string]PaymentProvider{
		"toss":   NewTossProvider(tossKey),
		"stripe": NewStripeProvider(stripeKey, stripeWebhookSecret),
	}
	return &PaymentServiceV2{
		db:        db,
		providers: providers,
	}
}

// DB 내부 DB 연결을 반환한다 (핸들러 호환성 유지).
func (s *PaymentServiceV2) DB() *sqlx.DB { return s.db }

// selectProvider currency에 따라 PSP를 선택한다.
// KRW → Toss, 그 외 → Stripe.
func (s *PaymentServiceV2) selectProvider(currency string) PaymentProvider {
	if strings.ToUpper(currency) == "KRW" || currency == "" {
		return s.providers["toss"]
	}
	return s.providers["stripe"]
}

// InitiatePayment 거래에 대한 결제를 초기화하고 결제창 URL을 반환한다.
// 적절한 PSP를 자동 선택하고 payment_logs에 기록한다.
func (s *PaymentServiceV2) InitiatePayment(ctx context.Context, tradeID int64, buyerID string) (*PaymentInitResultV2, error) {
	// 에스크로 정보 조회
	var escrow struct {
		Amount   int64  `db:"amount"`
		Currency string `db:"currency"`
		Status   string `db:"status"`
	}
	if err := s.db.GetContext(ctx, &escrow,
		`SELECT amount, currency, status FROM escrow_transactions WHERE trade_id=$1`, tradeID); err != nil {
		return nil, fmt.Errorf("에스크로 정보 없음: trade_id=%d", tradeID)
	}
	if escrow.Status != "PENDING" {
		return nil, fmt.Errorf("이미 처리된 거래입니다 (상태: %s)", escrow.Status)
	}

	currency := escrow.Currency
	if currency == "" {
		currency = "KRW"
	}

	provider := s.selectProvider(currency)
	orderID := fmt.Sprintf("AV-TRADE-%d-%d", tradeID, time.Now().Unix())

	req := PaymentRequest{
		OrderID:     orderID,
		Amount:      escrow.Amount,
		Currency:    currency,
		Description: fmt.Sprintf("Finara 거래 #%d", tradeID),
		SuccessURL:  "https://finara.app/payment/success",
		FailURL:     "https://finara.app/payment/fail",
	}

	resp, err := provider.Initiate(ctx, req)
	if err != nil {
		// payment_logs에 실패 기록
		s.logPayment(ctx, tradeID, provider.Name(), "", escrow.Amount, currency, "failed", nil) //nolint:errcheck
		return nil, fmt.Errorf("결제 초기화 실패 (%s): %w", provider.Name(), err)
	}

	// escrow_transactions 업데이트
	payload, _ := json.Marshal(resp)
	_, err = s.db.ExecContext(ctx, `
        UPDATE escrow_transactions
        SET pg_order_id=$1,
            pg_provider=$2,
            provider=$2,
            checkout_url=$3,
            pg_raw_response=$4
        WHERE trade_id=$5
    `, orderID, provider.Name(), resp.CheckoutURL, payload, tradeID)
	if err != nil {
		return nil, fmt.Errorf("에스크로 업데이트 실패: %w", err)
	}

	// payment_logs 기록
	s.logPayment(ctx, tradeID, provider.Name(), resp.ProviderTxnID, escrow.Amount, currency, "initiated", payload) //nolint:errcheck

	return &PaymentInitResultV2{
		TradeID:       tradeID,
		OrderID:       orderID,
		Provider:      provider.Name(),
		ProviderTxnID: resp.ProviderTxnID,
		CheckoutURL:   resp.CheckoutURL,
		ClientKey:     resp.ClientKey,
		Amount:        escrow.Amount,
		Currency:      currency,
		IsMock:        resp.IsMock,
	}, nil
}

// ConfirmPayment PSP 결제를 확인하고 에스크로 상태를 FUNDED로 전환한다.
func (s *PaymentServiceV2) ConfirmPayment(ctx context.Context, orderID, providerTxnID, providerName string, amount int64) error {
	// 에스크로 조회
	var row struct {
		EscrowID int64  `db:"id"`
		TradeID  int64  `db:"trade_id"`
		Currency string `db:"currency"`
	}
	if err := s.db.GetContext(ctx, &row,
		`SELECT id, trade_id, COALESCE(currency,'KRW') AS currency
         FROM escrow_transactions
         WHERE pg_order_id=$1 AND status='PENDING'`, orderID); err != nil {
		return fmt.Errorf("주문 없음 또는 이미 처리됨: %s", orderID)
	}

	// providerName이 명시되지 않으면 currency로 추론
	if providerName == "" {
		providerName = s.selectProvider(row.Currency).Name()
	}

	provider, ok := s.providers[providerName]
	if !ok {
		return fmt.Errorf("알 수 없는 PSP: %s", providerName)
	}

	// PSP 확인
	result, err := provider.Confirm(ctx, providerTxnID, amount)
	if err != nil {
		s.logPayment(ctx, row.TradeID, providerName, providerTxnID, amount, row.Currency, "failed", nil) //nolint:errcheck
		return fmt.Errorf("PSP 결제 확인 실패: %w", err)
	}
	if !result.Success {
		s.logPayment(ctx, row.TradeID, providerName, providerTxnID, amount, row.Currency, "failed", nil) //nolint:errcheck
		return fmt.Errorf("결제가 완료되지 않았습니다")
	}

	// 에스크로 상태 → FUNDED
	_, err = s.db.ExecContext(ctx, `
        UPDATE escrow_transactions
        SET status='FUNDED', pg_payment_key=$1, updated_at=NOW()
        WHERE id=$2
    `, providerTxnID, row.EscrowID)
	if err != nil {
		return fmt.Errorf("에스크로 업데이트 실패: %w", err)
	}

	// 거래 상태 → CONFIRMED
	_, err = s.db.ExecContext(ctx, `
        UPDATE trades SET status='CONFIRMED', updated_at=NOW() WHERE id=$1
    `, row.TradeID)
	if err != nil {
		return err
	}

	// payment_logs 기록
	s.logPayment(ctx, row.TradeID, providerName, providerTxnID, result.PaidAmount, row.Currency, "confirmed", nil) //nolint:errcheck
	return nil
}

// Refund PSP를 통해 환불을 처리하고 에스크로 상태를 REFUNDED로 전환한다.
func (s *PaymentServiceV2) Refund(ctx context.Context, tradeID int64, amount int64) error {
	var row struct {
		EscrowID      int64  `db:"id"`
		ProviderTxnID string `db:"pg_payment_key"`
		Provider      string `db:"provider"`
		Currency      string `db:"currency"`
		Status        string `db:"status"`
	}
	if err := s.db.GetContext(ctx, &row, `
        SELECT id, COALESCE(pg_payment_key,'') AS pg_payment_key,
               COALESCE(provider,'toss') AS provider,
               COALESCE(currency,'KRW') AS currency,
               status
        FROM escrow_transactions WHERE trade_id=$1`, tradeID); err != nil {
		return fmt.Errorf("에스크로 정보 없음: trade_id=%d", tradeID)
	}
	if row.Status != "FUNDED" {
		return fmt.Errorf("환불 불가 상태: %s", row.Status)
	}

	provider, ok := s.providers[row.Provider]
	if !ok {
		return fmt.Errorf("알 수 없는 PSP: %s", row.Provider)
	}

	if err := provider.Refund(ctx, row.ProviderTxnID, amount); err != nil {
		return fmt.Errorf("PSP 환불 실패: %w", err)
	}

	// 에스크로 상태 → REFUNDED
	_, err := s.db.ExecContext(ctx, `
        UPDATE escrow_transactions SET status='REFUNDED', updated_at=NOW() WHERE id=$1
    `, row.EscrowID)
	if err != nil {
		return err
	}

	s.logPayment(ctx, tradeID, row.Provider, row.ProviderTxnID, amount, row.Currency, "refunded", nil) //nolint:errcheck
	return nil
}

// HandleWebhook PSP 웹훅을 처리한다.
// providerName: "toss" 또는 "stripe" (요청 경로 또는 헤더로 구분)
// payload: 웹훅 원본 바디
// signature: PSP 서명 헤더 값
func (s *PaymentServiceV2) HandleWebhook(ctx context.Context, providerName string, payload []byte, signature string) error {
	provider, ok := s.providers[providerName]
	if !ok {
		return fmt.Errorf("알 수 없는 PSP: %s", providerName)
	}

	event, err := provider.VerifyWebhook(payload, signature)
	if err != nil {
		return fmt.Errorf("웹훅 검증 실패: %w", err)
	}

	switch event.EventType {
	case "payment.done":
		return s.ConfirmPayment(ctx, event.OrderID, event.ProviderTxnID, providerName, event.Amount)
	case "payment.failed":
		// 실패 로그만 기록 (에스크로 상태는 PENDING 유지)
		s.logPayment(ctx, 0, providerName, event.ProviderTxnID, event.Amount, "", "failed", payload) //nolint:errcheck
		return nil
	case "payment.refunded":
		s.logPayment(ctx, 0, providerName, event.ProviderTxnID, event.Amount, "", "refunded", payload) //nolint:errcheck
		return nil
	default:
		// 알 수 없는 이벤트 — 무시
		return nil
	}
}

// logPayment payment_logs 테이블에 결제 이벤트를 기록한다.
// 오류가 발생해도 주 흐름에 영향을 주지 않도록 에러를 반환하지만 호출자가 무시해도 된다.
func (s *PaymentServiceV2) logPayment(
	ctx context.Context,
	tradeID int64,
	provider, providerTxnID string,
	amount int64,
	currency, status string,
	payload []byte,
) error {
	var tradeIDPtr *int64
	if tradeID > 0 {
		tradeIDPtr = &tradeID
	}

	var payloadJSON interface{}
	if payload != nil {
		payloadJSON = payload
	}

	_, err := s.db.ExecContext(ctx, `
        INSERT INTO payment_logs
            (trade_id, provider, provider_txn_id, amount, currency, status, payload)
        VALUES ($1, $2, $3, $4, $5, $6, $7)
    `, tradeIDPtr, provider, providerTxnID, amount, currency, status, payloadJSON)
	return err
}

// PaymentInitResultV2 결제 초기화 결과 (V2 확장 버전).
type PaymentInitResultV2 struct {
	TradeID       int64  `json:"trade_id"`
	OrderID       string `json:"order_id"`
	Provider      string `json:"provider"`       // "toss" or "stripe"
	ProviderTxnID string `json:"provider_txn_id,omitempty"`
	CheckoutURL   string `json:"checkout_url"`
	ClientKey     string `json:"client_key,omitempty"`
	Amount        int64  `json:"amount"`
	Currency      string `json:"currency"`
	IsMock        bool   `json:"is_mock"`
}
