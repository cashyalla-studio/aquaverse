package domain

// Subscription 사용자 구독 정보
type Subscription struct {
	ID            int64   `db:"id" json:"id"`
	UserID        string  `db:"user_id" json:"user_id"`
	Plan          string  `db:"plan" json:"plan"`           // FREE, PRO
	Status        string  `db:"status" json:"status"`       // ACTIVE, CANCELLED, EXPIRED
	StartedAt     string  `db:"started_at" json:"started_at"`
	ExpiresAt     *string `db:"expires_at" json:"expires_at,omitempty"`
	BillingAmount int     `db:"billing_amount" json:"billing_amount"`
	CreatedAt     string  `db:"created_at" json:"created_at"`
}

// SubscriptionPlan 구독 플랜 정보
type SubscriptionPlan struct {
	ID       string   `json:"id"` // FREE, PRO
	Name     string   `json:"name"`
	PriceKRW int      `json:"price_krw"`
	Features []string `json:"features"`
}

// PredefinedPlans 사전 정의된 플랜 목록
var PredefinedPlans = []SubscriptionPlan{
	{
		ID: "FREE", Name: "무료",
		PriceKRW: 0,
		Features: []string{"기본 어종 정보 조회", "커뮤니티 참여", "하루 3회 AI 진단"},
	},
	{
		ID: "PRO", Name: "PRO",
		PriceKRW: 9900,
		Features: []string{
			"AI 진단 무제한",
			"광고 없음",
			"분양글 상단 노출 (월 1회)",
			"합사 AI 고급 추천",
			"수질 이력 무제한 저장",
		},
	},
}
