package service

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"
)

const stripeAPIBase = "https://api.stripe.com/v1"

// StripeProvider Stripe PSP 구현체.
// go-stripe 패키지 없이 Stripe REST API를 직접 HTTP 호출한다.
// PaymentProvider 인터페이스를 만족한다.
type StripeProvider struct {
	secretKey     string
	webhookSecret string
	httpClient    *http.Client
}

// NewStripeProvider StripeProvider를 생성한다.
// secretKey가 비어 있으면 mock 모드로 동작한다.
func NewStripeProvider(secretKey, webhookSecret string) *StripeProvider {
	return &StripeProvider{
		secretKey:     secretKey,
		webhookSecret: webhookSecret,
		httpClient:    &http.Client{Timeout: 30 * time.Second},
	}
}

// Name PSP 이름을 반환한다.
func (p *StripeProvider) Name() string { return "stripe" }

// Initiate Stripe Checkout Session을 생성하고 결제창 URL을 반환한다.
// STRIPE_SECRET_KEY 미설정 시 mock 응답을 반환한다.
func (p *StripeProvider) Initiate(ctx context.Context, req PaymentRequest) (*PaymentInitResponse, error) {
	if p.secretKey == "" {
		// mock 응답 — 환경변수 미설정 시 개발/테스트용
		return &PaymentInitResponse{
			Provider:      "stripe",
			ProviderTxnID: "cs_mock_" + req.OrderID,
			CheckoutURL:   "https://checkout.stripe.com/c/pay/mock_" + req.OrderID,
			IsMock:        true,
		}, nil
	}

	// Stripe Checkout Session POST /v1/checkout/sessions
	// 참고: https://stripe.com/docs/api/checkout/sessions/create
	form := url.Values{}
	form.Set("mode", "payment")
	form.Set("payment_method_types[]", "card")
	form.Set("line_items[0][price_data][currency]", strings.ToLower(req.Currency))
	form.Set("line_items[0][price_data][unit_amount]", strconv.FormatInt(req.Amount, 10))
	form.Set("line_items[0][price_data][product_data][name]", req.Description)
	form.Set("line_items[0][quantity]", "1")
	form.Set("client_reference_id", req.OrderID)
	form.Set("success_url", withOrderID(req.SuccessURL, req.OrderID))
	form.Set("cancel_url", withOrderID(req.FailURL, req.OrderID))
	// Stripe는 metadata로 주문 ID를 전달
	form.Set("metadata[order_id]", req.OrderID)

	httpReq, err := http.NewRequestWithContext(ctx, "POST",
		stripeAPIBase+"/checkout/sessions",
		strings.NewReader(form.Encode()),
	)
	if err != nil {
		return nil, fmt.Errorf("stripe: request 생성 실패: %w", err)
	}
	httpReq.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	httpReq.SetBasicAuth(p.secretKey, "")

	resp, err := p.httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("stripe: API 호출 실패: %w", err)
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("stripe: Checkout Session 생성 실패 (status %d): %s",
			resp.StatusCode, string(body))
	}

	var session struct {
		ID  string `json:"id"`
		URL string `json:"url"`
	}
	if err := json.Unmarshal(body, &session); err != nil {
		return nil, fmt.Errorf("stripe: 응답 파싱 실패: %w", err)
	}

	return &PaymentInitResponse{
		Provider:      "stripe",
		ProviderTxnID: session.ID,
		CheckoutURL:   session.URL,
		IsMock:        false,
	}, nil
}

// Confirm Stripe PaymentIntent를 조회하여 결제를 확인한다.
// mock 모드에서는 항상 성공을 반환한다.
// Stripe Checkout은 웹훅으로 확인하는 것이 정석이지만,
// 클라이언트 리다이렉트 후 폴링 시나리오를 위해 Checkout Session 조회를 지원한다.
func (p *StripeProvider) Confirm(ctx context.Context, providerTxnID string, amount int64) (*PaymentConfirmResponse, error) {
	if p.secretKey == "" {
		return &PaymentConfirmResponse{
			Success:       true,
			ProviderTxnID: providerTxnID,
			PaidAmount:    amount,
		}, nil
	}

	// GET /v1/checkout/sessions/{session_id}
	httpReq, err := http.NewRequestWithContext(ctx, "GET",
		stripeAPIBase+"/checkout/sessions/"+providerTxnID, nil)
	if err != nil {
		return nil, fmt.Errorf("stripe: confirm request 생성 실패: %w", err)
	}
	httpReq.SetBasicAuth(p.secretKey, "")

	resp, err := p.httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("stripe: confirm API 호출 실패: %w", err)
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("stripe: Session 조회 실패 (status %d): %s",
			resp.StatusCode, string(body))
	}

	var session struct {
		PaymentStatus string `json:"payment_status"` // "paid", "unpaid", "no_payment_required"
		AmountTotal   int64  `json:"amount_total"`
		PaymentIntent string `json:"payment_intent"`
	}
	if err := json.Unmarshal(body, &session); err != nil {
		return nil, fmt.Errorf("stripe: confirm 응답 파싱 실패: %w", err)
	}

	if session.PaymentStatus != "paid" {
		return &PaymentConfirmResponse{
			Success:       false,
			ProviderTxnID: providerTxnID,
			PaidAmount:    0,
		}, nil
	}

	return &PaymentConfirmResponse{
		Success:       true,
		ProviderTxnID: providerTxnID,
		PaidAmount:    session.AmountTotal,
	}, nil
}

// Refund Stripe Refund를 생성한다.
// providerTxnID는 PaymentIntent ID (pi_xxx) 또는 Charge ID (ch_xxx)를 받는다.
// mock 모드에서는 실제 API 호출을 건너뛴다.
func (p *StripeProvider) Refund(ctx context.Context, providerTxnID string, amount int64) error {
	if p.secretKey == "" {
		return nil
	}

	form := url.Values{}
	// payment_intent 또는 charge 중 하나를 전달
	if strings.HasPrefix(providerTxnID, "pi_") {
		form.Set("payment_intent", providerTxnID)
	} else {
		form.Set("charge", providerTxnID)
	}
	if amount > 0 {
		form.Set("amount", strconv.FormatInt(amount, 10))
	}

	httpReq, err := http.NewRequestWithContext(ctx, "POST",
		stripeAPIBase+"/refunds",
		strings.NewReader(form.Encode()),
	)
	if err != nil {
		return fmt.Errorf("stripe: refund request 생성 실패: %w", err)
	}
	httpReq.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	httpReq.SetBasicAuth(p.secretKey, "")

	resp, err := p.httpClient.Do(httpReq)
	if err != nil {
		return fmt.Errorf("stripe: refund API 호출 실패: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("stripe: 환불 실패 (status %d): %s", resp.StatusCode, string(body))
	}
	return nil
}

// VerifyWebhook Stripe-Signature 헤더를 검증하고 웹훅 이벤트를 파싱한다.
// signature 파라미터는 "Stripe-Signature" 헤더 값이어야 한다.
// webhookSecret이 비어 있으면 서명 검증을 건너뛰고 페이로드만 파싱한다.
func (p *StripeProvider) VerifyWebhook(payload []byte, signature string) (*WebhookEvent, error) {
	// 서명 검증 (webhookSecret이 설정된 경우)
	if p.webhookSecret != "" {
		if err := p.verifyStripeSignature(payload, signature); err != nil {
			return nil, fmt.Errorf("stripe: 웹훅 서명 검증 실패: %w", err)
		}
	}

	// 이벤트 파싱
	var event struct {
		Type string `json:"type"`
		Data struct {
			Object struct {
				ID            string `json:"id"`
				PaymentIntent string `json:"payment_intent"`
				AmountTotal   int64  `json:"amount_total"`
				// checkout.session 공통
				ClientRefID string `json:"client_reference_id"` // = order_id
				Metadata    struct {
					OrderID string `json:"order_id"`
				} `json:"metadata"`
				// payment_intent 공통
				Amount int64 `json:"amount"`
			} `json:"object"`
		} `json:"data"`
	}
	if err := json.Unmarshal(payload, &event); err != nil {
		return nil, fmt.Errorf("stripe: 웹훅 파싱 실패: %w", err)
	}

	obj := event.Data.Object
	// OrderID 추출: metadata.order_id → client_reference_id 순으로 폴백
	orderID := obj.Metadata.OrderID
	if orderID == "" {
		orderID = obj.ClientRefID
	}

	// 금액 추출: amount_total(checkout.session) → amount(payment_intent)
	amount := obj.AmountTotal
	if amount == 0 {
		amount = obj.Amount
	}

	// txnID: checkout session id → payment_intent id 순으로 폴백
	txnID := obj.ID
	if obj.PaymentIntent != "" {
		txnID = obj.PaymentIntent
	}

	// Stripe 이벤트 타입 → 공통 이벤트 타입 변환
	var eventType string
	switch event.Type {
	case "checkout.session.completed", "payment_intent.succeeded":
		eventType = "payment.done"
	case "checkout.session.expired", "payment_intent.payment_failed":
		eventType = "payment.failed"
	case "charge.refunded":
		eventType = "payment.refunded"
	default:
		return &WebhookEvent{EventType: "payment.unknown"}, nil
	}

	return &WebhookEvent{
		EventType:     eventType,
		ProviderTxnID: txnID,
		OrderID:       orderID,
		Amount:        amount,
	}, nil
}

// verifyStripeSignature Stripe 웹훅 서명을 HMAC-SHA256으로 검증한다.
// 참고: https://stripe.com/docs/webhooks/signatures
func (p *StripeProvider) verifyStripeSignature(payload []byte, sigHeader string) error {
	// sigHeader 형식: t=타임스탬프,v1=서명,v1=서명,...
	var timestamp, v1sig string
	for _, part := range strings.Split(sigHeader, ",") {
		if strings.HasPrefix(part, "t=") {
			timestamp = strings.TrimPrefix(part, "t=")
		} else if strings.HasPrefix(part, "v1=") {
			v1sig = strings.TrimPrefix(part, "v1=")
		}
	}
	if timestamp == "" || v1sig == "" {
		return fmt.Errorf("서명 헤더 형식 오류")
	}

	// signed_payload = timestamp + "." + payload
	signedPayload := timestamp + "." + string(payload)
	mac := hmac.New(sha256.New, []byte(p.webhookSecret))
	mac.Write([]byte(signedPayload))
	expected := hex.EncodeToString(mac.Sum(nil))

	if !hmac.Equal([]byte(expected), []byte(v1sig)) {
		return fmt.Errorf("서명 불일치")
	}

	// 재전송 공격 방지: 5분 이내 타임스탬프만 허용
	ts, err := strconv.ParseInt(timestamp, 10, 64)
	if err != nil {
		return fmt.Errorf("타임스탬프 파싱 오류")
	}
	if time.Since(time.Unix(ts, 0)) > 5*time.Minute {
		return fmt.Errorf("웹훅 타임스탬프 만료 (재전송 공격 방지)")
	}

	return nil
}

// withOrderID success/cancel URL에 order_id 쿼리 파라미터를 추가한다.
func withOrderID(baseURL, orderID string) string {
	if baseURL == "" {
		return ""
	}
	sep := "?"
	if strings.Contains(baseURL, "?") {
		sep = "&"
	}
	return baseURL + sep + "order_id=" + url.QueryEscape(orderID)
}
