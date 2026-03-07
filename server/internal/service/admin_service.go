package service

import (
	"context"
	"encoding/json"
	"time"

	"github.com/cashyalla/aquaverse/internal/domain"
	"github.com/jmoiron/sqlx"
)

type AdminService struct {
	db *sqlx.DB
}

func NewAdminService(db *sqlx.DB) *AdminService {
	return &AdminService{db: db}
}

func (s *AdminService) GetKPI(ctx context.Context) (*domain.AdminKPI, error) {
	kpi := &domain.AdminKPI{}

	// MAU (30일 내 로그인)
	s.db.QueryRowContext(ctx, `
		SELECT COUNT(DISTINCT id) FROM users
		WHERE updated_at >= NOW() - INTERVAL '30 days'
	`).Scan(&kpi.MAU)

	// DAU (오늘 로그인)
	s.db.QueryRowContext(ctx, `
		SELECT COUNT(DISTINCT id) FROM users
		WHERE updated_at >= NOW() - INTERVAL '1 day'
	`).Scan(&kpi.DAU)

	// Total Users
	s.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM users`).Scan(&kpi.TotalUsers)

	// PRO Subscribers
	s.db.QueryRowContext(ctx, `
		SELECT COUNT(*) FROM user_subscriptions
		WHERE plan = 'PRO' AND status = 'ACTIVE'
	`).Scan(&kpi.ProSubscribers)

	// Total Listings
	s.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM listings`).Scan(&kpi.TotalListings)

	// Active Trades
	s.db.QueryRowContext(ctx, `
		SELECT COUNT(*) FROM trades
		WHERE status NOT IN ('COMPLETED', 'CANCELLED')
	`).Scan(&kpi.ActiveTrades)

	// Escrow Success Rate
	var released, total int
	s.db.QueryRowContext(ctx, `
		SELECT
			COUNT(CASE WHEN status = 'RELEASED' THEN 1 END),
			COUNT(*)
		FROM escrow_transactions
	`).Scan(&released, &total)
	if total > 0 {
		kpi.EscrowSuccessRate = float64(released) / float64(total) * 100
	}

	// Escrow Disputes
	s.db.QueryRowContext(ctx, `
		SELECT COUNT(*) FROM escrow_transactions WHERE status = 'DISPUTED'
	`).Scan(&kpi.EscrowDisputes)

	// CITES Filter Hits today
	s.db.QueryRowContext(ctx, `
		SELECT COUNT(*) FROM admin_audit_logs
		WHERE action = 'CITES_FILTER_HIT'
		AND created_at >= NOW() - INTERVAL '1 day'
	`).Scan(&kpi.CitesFilterHits)

	return kpi, nil
}

func (s *AdminService) ListUsers(ctx context.Context, limit, offset int, query string) ([]domain.AdminUserInfo, error) {
	var users []domain.AdminUserInfo
	sql := `
		SELECT u.id, u.email, u.username, u.role, u.trust_score,
		       u.is_banned, u.phone_verified, u.created_at,
		       COALESCE(lc.cnt, 0) AS listing_count,
		       COALESCE(tc.cnt, 0) AS trade_count,
		       COALESCE(fr.cnt, 0) AS fraud_reports
		FROM users u
		LEFT JOIN (SELECT seller_id, COUNT(*) cnt FROM listings GROUP BY seller_id) lc ON lc.seller_id = u.id
		LEFT JOIN (SELECT buyer_id, COUNT(*) cnt FROM trades GROUP BY buyer_id) tc ON tc.buyer_id = u.id
		LEFT JOIN (SELECT reported_user_id, COUNT(*) cnt FROM fraud_reports GROUP BY reported_user_id) fr ON fr.reported_user_id = u.id
	`
	if query != "" {
		sql += ` WHERE u.email ILIKE $3 OR u.username ILIKE $3`
		sql += ` ORDER BY u.created_at DESC LIMIT $1 OFFSET $2`
		return users, s.db.SelectContext(ctx, &users, sql, limit, offset, "%"+query+"%")
	}
	sql += ` ORDER BY u.created_at DESC LIMIT $1 OFFSET $2`
	return users, s.db.SelectContext(ctx, &users, sql, limit, offset)
}

func (s *AdminService) BanUser(ctx context.Context, adminID, userID, reason, ip string) error {
	_, err := s.db.ExecContext(ctx, `UPDATE users SET is_banned = true WHERE id = $1`, userID)
	if err != nil {
		return err
	}
	detail, _ := json.Marshal(map[string]string{"reason": reason})
	_, err = s.db.ExecContext(ctx, `
		INSERT INTO admin_audit_logs (admin_id, action, target_type, target_id, detail, ip_address)
		VALUES ($1, 'BAN_USER', 'USER', $2, $3, $4)
	`, adminID, userID, detail, ip)
	return err
}

func (s *AdminService) UnbanUser(ctx context.Context, adminID, userID, ip string) error {
	_, err := s.db.ExecContext(ctx, `UPDATE users SET is_banned = false WHERE id = $1`, userID)
	if err != nil {
		return err
	}
	_, err = s.db.ExecContext(ctx, `
		INSERT INTO admin_audit_logs (admin_id, action, target_type, target_id, ip_address)
		VALUES ($1, 'UNBAN_USER', 'USER', $2, $3)
	`, adminID, userID, ip)
	return err
}

func (s *AdminService) GetAuditLogs(ctx context.Context, limit, offset int) ([]domain.AdminAuditLog, error) {
	var logs []domain.AdminAuditLog
	err := s.db.SelectContext(ctx, &logs, `
		SELECT al.id, al.admin_id::text, u.email AS admin_email,
		       al.action, COALESCE(al.target_type, '') AS target_type,
		       COALESCE(al.target_id, '') AS target_id,
		       al.detail, COALESCE(al.ip_address::text, '') AS ip_address,
		       al.created_at
		FROM admin_audit_logs al
		JOIN users u ON u.id = al.admin_id
		ORDER BY al.created_at DESC
		LIMIT $1 OFFSET $2
	`, limit, offset)
	return logs, err
}

func (s *AdminService) GetCitesAuditStats(ctx context.Context) ([]map[string]interface{}, error) {
	// 최근 30일 CITES 필터 히트 통계 (실제로는 별도 cites_audit 테이블이 있어야 하지만 현재 없으므로 mock)
	_ = ctx
	return []map[string]interface{}{
		{"date": time.Now().Format("2006-01-02"), "hits": 0, "blocked": 0, "warned": 0},
	}, nil
}
