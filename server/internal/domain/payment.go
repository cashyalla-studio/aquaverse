package domain

// PaymentInitResult 결제 초기화 결과
type PaymentInitResult struct {
	TradeID     int64  `json:"trade_id"`
	OrderID     string `json:"order_id"`
	CheckoutURL string `json:"checkout_url"`
	Amount      int64  `json:"amount"`
	Currency    string `json:"currency"`
	IsSandbox   bool   `json:"is_sandbox"`
}

// TossWebhookPayload 토스페이먼츠 웹훅 페이로드
type TossWebhookPayload struct {
	EventType string `json:"eventType"`
	Data      struct {
		PaymentKey string `json:"paymentKey"`
		OrderID    string `json:"orderId"`
		Amount     int64  `json:"amount"`
		Status     string `json:"status"` // DONE, CANCELED, etc.
	} `json:"data"`
}
