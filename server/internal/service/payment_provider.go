package service

import "context"

// PaymentProvider 다중 PSP(Payment Service Provider)를 위한 공통 인터페이스.
// Toss, Stripe 등 각 PSP 구현체는 이 인터페이스를 만족해야 한다.
type PaymentProvider interface {
	// Initiate 결제를 초기화하고 결제창 URL 또는 클라이언트 키를 반환한다.
	Initiate(ctx context.Context, req PaymentRequest) (*PaymentInitResponse, error)
	// Confirm PSP 측 결제를 검증하고 확인 처리한다.
	Confirm(ctx context.Context, providerTxnID string, amount int64) (*PaymentConfirmResponse, error)
	// Refund 결제를 환불한다.
	Refund(ctx context.Context, providerTxnID string, amount int64) error
	// VerifyWebhook PSP가 전송한 웹훅 페이로드와 서명을 검증한다.
	VerifyWebhook(payload []byte, signature string) (*WebhookEvent, error)
	// Name PSP 식별자 문자열 ("toss", "stripe" 등)을 반환한다.
	Name() string
}

// PaymentRequest 결제 초기화 요청 공통 구조체.
type PaymentRequest struct {
	OrderID     string // 주문 고유 식별자 (플랫폼 내부 생성)
	Amount      int64  // 결제 금액 (최소 단위: KRW=원, USD=센트)
	Currency    string // ISO 4217 통화 코드 (KRW, USD 등)
	Description string // 결제창에 표시할 주문명
	SuccessURL  string // 결제 완료 후 리다이렉트 URL
	FailURL     string // 결제 실패/취소 후 리다이렉트 URL
}

// PaymentInitResponse 결제 초기화 결과 공통 구조체.
type PaymentInitResponse struct {
	Provider      string // PSP 이름 ("toss", "stripe")
	ProviderTxnID string // PSP 내부 거래 ID (Toss: paymentKey, Stripe: cs_xxx)
	CheckoutURL   string // 결제창 URL (Stripe Checkout Session URL 등)
	ClientKey     string // 클라이언트용 키 (Toss: clientKey, Stripe: publishableKey)
	IsMock        bool   // 환경변수 미설정 시 mock 모드 여부
}

// PaymentConfirmResponse 결제 확인 결과 공통 구조체.
type PaymentConfirmResponse struct {
	Success       bool
	ProviderTxnID string
	PaidAmount    int64
}

// WebhookEvent PSP 웹훅에서 파싱된 공통 이벤트 구조체.
type WebhookEvent struct {
	EventType     string // 'payment.done', 'payment.failed', 'payment.refunded'
	ProviderTxnID string // PSP 내부 거래 ID
	OrderID       string // 플랫폼 주문 ID
	Amount        int64  // 금액
}
