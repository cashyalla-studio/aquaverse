package service

import (
	"context"
	"errors"
	"fmt"

	"github.com/cashyalla/aquaverse/internal/domain"
	"github.com/jmoiron/sqlx"
	"github.com/shopspring/decimal"
)

// EscrowService: 에스크로 거래 관리
// 실제 PG 연동은 P2. P1에서는 상태 관리 + UI 레이어만 구현.
type EscrowService struct {
	db *sqlx.DB
}

func NewEscrowService(db *sqlx.DB) *EscrowService {
	return &EscrowService{db: db}
}

// CreateEscrow: 거래 시작 시 에스크로 레코드 생성
func (s *EscrowService) CreateEscrow(ctx context.Context, tradeID int64, amount decimal.Decimal, currency string) error {
	_, err := s.db.ExecContext(ctx,
		`INSERT INTO escrow_transactions (trade_id, amount, currency, status)
         VALUES ($1, $2, $3, 'PENDING')
         ON CONFLICT (trade_id) DO NOTHING`,
		tradeID, amount, currency,
	)
	return err
}

// FundEscrow: 구매자가 에스크로에 금액 입금 (PG 연동 전: 시뮬레이션)
func (s *EscrowService) FundEscrow(ctx context.Context, tradeID int64, userID string) error {
	tx, err := s.db.BeginTxx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	var escrow struct {
		ID     int64  `db:"id"`
		Status string `db:"status"`
	}
	if err := tx.GetContext(ctx, &escrow,
		`SELECT id, status FROM escrow_transactions WHERE trade_id = $1`, tradeID,
	); err != nil {
		return fmt.Errorf("escrow not found: %w", err)
	}
	if escrow.Status != "PENDING" {
		return errors.New("escrow is not in PENDING state")
	}

	_, err = tx.ExecContext(ctx,
		`UPDATE escrow_transactions SET status='FUNDED', funded_at=NOW(), updated_at=NOW() WHERE trade_id=$1`,
		tradeID,
	)
	if err != nil {
		return err
	}
	// 거래 상태도 CONFIRMED로
	_, err = tx.ExecContext(ctx,
		`UPDATE trades SET status=$1, updated_at=NOW() WHERE id=$2`,
		domain.TradeStatusConfirmed, tradeID,
	)
	if err != nil {
		return err
	}
	return tx.Commit()
}

// ReleaseEscrow: 거래 완료 후 판매자에게 출금 (P2: PG 실제 송금)
func (s *EscrowService) ReleaseEscrow(ctx context.Context, tradeID int64) error {
	_, err := s.db.ExecContext(ctx,
		`UPDATE escrow_transactions SET status='RELEASED', released_at=NOW(), updated_at=NOW() WHERE trade_id=$1 AND status='FUNDED'`,
		tradeID,
	)
	return err
}

// RefundEscrow: 분쟁 또는 취소 시 환불
func (s *EscrowService) RefundEscrow(ctx context.Context, tradeID int64, reason string) error {
	_, err := s.db.ExecContext(ctx,
		`UPDATE escrow_transactions SET status='REFUNDED', refunded_at=NOW(), dispute_reason=$2, updated_at=NOW() WHERE trade_id=$1`,
		tradeID, reason,
	)
	return err
}

// GetEscrowStatus: 에스크로 상태 조회
func (s *EscrowService) GetEscrowStatus(ctx context.Context, tradeID int64) (string, error) {
	var status string
	err := s.db.GetContext(ctx, &status,
		`SELECT status FROM escrow_transactions WHERE trade_id=$1`, tradeID,
	)
	return status, err
}
