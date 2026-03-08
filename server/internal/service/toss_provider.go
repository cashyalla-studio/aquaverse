package service

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

// TossPaymentsProvider 토스페이먼츠 PSP 구현체.
// PaymentProvider 인터페이스를 만족한다.
type TossPaymentsProvider struct {
	secretKey  string
	baseURL    string
	isSandbox  bool
	httpClient *http.Client
}

// NewTossProvider TossPaymentsProvider를 생성한다.
// secretKey가 비어 있거나 "test"이면 sandbox(mock) 모드로 동작한다.
func NewTossProvider(secretKey string) *TossPaymentsProvider {
	isSandbox := secretKey == "" || secretKey == "test"
	return &TossPaymentsProvider{
		secretKey:  secretKey,
		baseURL:    "https://api.tosspayments.com/v1",
		isSandbox:  isSandbox,
		httpClient: &http.Client{Timeout: 30 * time.Second},
	}
}

// Name PSP 이름을 반환한다.
func (p *TossPaymentsProvider) Name() string { return "toss" }

// Initiate 토스페이먼츠 결제를 초기화한다.
// sandbox 모드면 실제 API 호출 없이 mock 결제창 URL을 반환한다.
func (p *TossPaymentsProvider) Initiate(ctx context.Context, req PaymentRequest) (*PaymentInitResponse, error) {
	if p.isSandbox {
		checkoutURL := fmt.Sprintf(
			"https://sandbox.tosspayments.com/widget/sdk-v2/checkout?orderId=%s&amount=%d&orderName=%s",
			req.OrderID, req.Amount, urlEncode(req.Description),
		)
		return &PaymentInitResponse{
			Provider:      "toss",
			ProviderTxnID: "mock_toss_" + req.OrderID,
			CheckoutURL:   checkoutURL,
			ClientKey:     "test_ck_mock",
			IsMock:        true,
		}, nil
	}

	// 실제 토스페이먼츠 Payment Intent 생성
	reqBody := map[string]interface{}{
		"amount":      req.Amount,
		"orderId":     req.OrderID,
		"orderName":   req.Description,
		"currency":    req.Currency,
		"method":      "카드",
		"successUrl":  req.SuccessURL,
		"failUrl":     req.FailURL,
	}
	bodyBytes, _ := json.Marshal(reqBody)

	httpReq, err := http.NewRequestWithContext(ctx, "POST",
		p.baseURL+"/payments/key-in", bytes.NewReader(bodyBytes))
	if err != nil {
		return nil, fmt.Errorf("toss: request 생성 실패: %w", err)
	}
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Authorization", "Basic "+p.basicAuth())

	resp, err := p.httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("toss: API 호출 실패: %w", err)
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)

	var result map[string]interface{}
	json.Unmarshal(body, &result) //nolint:errcheck

	checkoutURL := fmt.Sprintf(
		"https://pay.toss.im/widget/v2?orderId=%s&amount=%d",
		req.OrderID, req.Amount,
	)
	if url, ok := result["checkoutUrl"].(string); ok {
		checkoutURL = url
	}

	txnID := ""
	if key, ok := result["paymentKey"].(string); ok {
		txnID = key
	}

	return &PaymentInitResponse{
		Provider:      "toss",
		ProviderTxnID: txnID,
		CheckoutURL:   checkoutURL,
		IsMock:        false,
	}, nil
}

// Confirm 토스페이먼츠 결제를 확인한다.
// sandbox 모드에서는 실제 API 호출을 건너뛰고 성공 응답을 반환한다.
func (p *TossPaymentsProvider) Confirm(ctx context.Context, providerTxnID string, amount int64) (*PaymentConfirmResponse, error) {
	if p.isSandbox {
		return &PaymentConfirmResponse{
			Success:       true,
			ProviderTxnID: providerTxnID,
			PaidAmount:    amount,
		}, nil
	}

	reqBody := map[string]interface{}{
		"paymentKey": providerTxnID,
		"amount":     amount,
	}
	bodyBytes, _ := json.Marshal(reqBody)

	httpReq, err := http.NewRequestWithContext(ctx, "POST",
		p.baseURL+"/payments/confirm", bytes.NewReader(bodyBytes))
	if err != nil {
		return nil, fmt.Errorf("toss: confirm request 생성 실패: %w", err)
	}
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Authorization", "Basic "+p.basicAuth())

	resp, err := p.httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("toss: confirm API 호출 실패: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("toss: 결제 확인 실패 (status %d): %s", resp.StatusCode, string(body))
	}

	return &PaymentConfirmResponse{
		Success:       true,
		ProviderTxnID: providerTxnID,
		PaidAmount:    amount,
	}, nil
}

// Refund 토스페이먼츠 결제를 환불한다.
// sandbox 모드에서는 실제 API 호출을 건너뛴다.
func (p *TossPaymentsProvider) Refund(ctx context.Context, providerTxnID string, amount int64) error {
	if p.isSandbox {
		return nil
	}

	reqBody := map[string]interface{}{
		"cancelReason": "고객 요청 환불",
		"cancelAmount": amount,
	}
	bodyBytes, _ := json.Marshal(reqBody)

	url := fmt.Sprintf("%s/payments/%s/cancel", p.baseURL, providerTxnID)
	httpReq, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewReader(bodyBytes))
	if err != nil {
		return fmt.Errorf("toss: refund request 생성 실패: %w", err)
	}
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Authorization", "Basic "+p.basicAuth())

	resp, err := p.httpClient.Do(httpReq)
	if err != nil {
		return fmt.Errorf("toss: refund API 호출 실패: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("toss: 환불 실패 (status %d): %s", resp.StatusCode, string(body))
	}
	return nil
}

// VerifyWebhook 토스페이먼츠 웹훅을 파싱하고 검증한다.
// 토스는 별도 서명 헤더 없이 웹훅 시크릿 헤더(Toss-Signature)를 사용한다.
// 현재 구현에서는 페이로드 파싱에 집중하며, 시그니처 검증은 호출자가 수행한다.
func (p *TossPaymentsProvider) VerifyWebhook(payload []byte, signature string) (*WebhookEvent, error) {
	var raw struct {
		EventType string `json:"eventType"`
		Data      struct {
			PaymentKey string `json:"paymentKey"`
			OrderID    string `json:"orderId"`
			Amount     int64  `json:"amount"`
			Status     string `json:"status"`
		} `json:"data"`
	}
	if err := json.Unmarshal(payload, &raw); err != nil {
		return nil, fmt.Errorf("toss: 웹훅 파싱 실패: %w", err)
	}

	// PAYMENT_STATUS_CHANGED + DONE → payment.done
	// 그 외 이벤트는 무시
	eventType := ""
	switch {
	case raw.EventType == "PAYMENT_STATUS_CHANGED" && raw.Data.Status == "DONE":
		eventType = "payment.done"
	case raw.EventType == "PAYMENT_STATUS_CHANGED" && raw.Data.Status == "CANCELED":
		eventType = "payment.failed"
	default:
		// 알 수 없는 이벤트 — 빈 이벤트 반환 (호출자가 무시)
		return &WebhookEvent{EventType: "payment.unknown"}, nil
	}

	return &WebhookEvent{
		EventType:     eventType,
		ProviderTxnID: raw.Data.PaymentKey,
		OrderID:       raw.Data.OrderID,
		Amount:        raw.Data.Amount,
	}, nil
}

// basicAuth HTTP Basic Auth 인코딩 문자열을 반환한다 (secretKey:).
func (p *TossPaymentsProvider) basicAuth() string {
	return base64.StdEncoding.EncodeToString([]byte(p.secretKey + ":"))
}

// urlEncode URL에서 사용할 수 있도록 공백을 %20으로 치환한다 (간단 버전).
func urlEncode(s string) string {
	result := ""
	for _, c := range s {
		switch c {
		case ' ':
			result += "%20"
		case '&':
			result += "%26"
		case '=':
			result += "%3D"
		default:
			result += string(c)
		}
	}
	return result
}
