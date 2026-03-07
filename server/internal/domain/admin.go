package domain

import "time"

type AdminKPI struct {
	MAU               int     `json:"mau"`                // 월간 활성 사용자
	DAU               int     `json:"dau"`                // 일간 활성 사용자
	TotalUsers        int     `json:"total_users"`
	ProSubscribers    int     `json:"pro_subscribers"`
	TotalListings     int     `json:"total_listings"`
	ActiveTrades      int     `json:"active_trades"`
	EscrowSuccessRate float64 `json:"escrow_success_rate"` // %
	EscrowDisputes    int     `json:"escrow_disputes"`
	TotalRevenue      int64   `json:"total_revenue_krw"`   // 원화
	CitesFilterHits   int     `json:"cites_filter_hits"`   // 오늘
	ClaudeAPICallsToday int   `json:"claude_api_calls_today"`
}

type AdminUserInfo struct {
	ID            string    `json:"id" db:"id"`
	Email         string    `json:"email" db:"email"`
	Username      string    `json:"username" db:"username"`
	Role          string    `json:"role" db:"role"`
	TrustScore    float64   `json:"trust_score" db:"trust_score"`
	IsBanned      bool      `json:"is_banned" db:"is_banned"`
	PhoneVerified bool      `json:"phone_verified" db:"phone_verified"`
	CreatedAt     time.Time `json:"created_at" db:"created_at"`
	ListingCount  int       `json:"listing_count" db:"listing_count"`
	TradeCount    int       `json:"trade_count" db:"trade_count"`
	FraudReports  int       `json:"fraud_reports" db:"fraud_reports"`
}

type AdminAuditLog struct {
	ID         int64     `json:"id" db:"id"`
	AdminID    string    `json:"admin_id" db:"admin_id"`
	AdminEmail string    `json:"admin_email" db:"admin_email"`
	Action     string    `json:"action" db:"action"`
	TargetType string    `json:"target_type" db:"target_type"`
	TargetID   string    `json:"target_id" db:"target_id"`
	Detail     []byte    `json:"detail" db:"detail"`
	IPAddress  string    `json:"ip_address" db:"ip_address"`
	CreatedAt  time.Time `json:"created_at" db:"created_at"`
}

type BanUserRequest struct {
	UserID string `json:"user_id"`
	Reason string `json:"reason"`
}
