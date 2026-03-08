package service

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/cashyalla/aquaverse/internal/domain"
	"github.com/cashyalla/aquaverse/internal/repository"
	"github.com/jmoiron/sqlx"
)

// InventoryService Business Hub 재고 관리 서비스
type InventoryService struct {
	db        *sqlx.DB
	citesRepo *repository.CitesRepository
}

// NewInventoryService InventoryService를 생성한다.
func NewInventoryService(db *sqlx.DB, citesRepo *repository.CitesRepository) *InventoryService {
	return &InventoryService{db: db, citesRepo: citesRepo}
}

// ownsBusinessProfile 해당 businessID가 userID 소유인지 확인한다.
func (s *InventoryService) ownsBusinessProfile(ctx context.Context, businessID int64, userID string) (bool, error) {
	var ownerID string
	err := s.db.QueryRowContext(ctx,
		`SELECT user_id FROM business_profiles WHERE id = $1`, businessID,
	).Scan(&ownerID)
	if err == sql.ErrNoRows {
		return false, fmt.Errorf("업체를 찾을 수 없습니다")
	}
	if err != nil {
		return false, err
	}
	return ownerID == userID, nil
}

// ListInventory 업체 재고 목록 조회
func (s *InventoryService) ListInventory(ctx context.Context, businessID int64) ([]domain.ShopInventory, error) {
	var items []domain.ShopInventory
	err := s.db.SelectContext(ctx, &items, `
		SELECT
			si.id, si.business_id, si.fish_data_id, si.custom_name,
			si.quantity, si.price, si.cites_status, si.is_available, si.updated_at,
			COALESCE(fd.primary_common_name, '') AS fish_name
		FROM shop_inventory si
		LEFT JOIN fish_data fd ON fd.id = si.fish_data_id
		WHERE si.business_id = $1
		ORDER BY si.is_available DESC, si.updated_at DESC
	`, businessID)
	return items, err
}

// UpsertInventory 재고 추가 또는 수정 (CITES 자동 체크 포함)
func (s *InventoryService) UpsertInventory(ctx context.Context, businessID int64, userID string, req domain.ShopInventoryRequest) (*domain.ShopInventory, error) {
	ok, err := s.ownsBusinessProfile(ctx, businessID, userID)
	if err != nil {
		return nil, err
	}
	if !ok {
		return nil, fmt.Errorf("해당 업체를 수정할 권한이 없습니다")
	}

	if req.FishDataID == nil && req.CustomName == "" {
		return nil, fmt.Errorf("fish_data_id 또는 custom_name 중 하나는 필수입니다")
	}

	// CITES 자동 체크: fish_data_id가 있으면 학명을 조회하여 체크
	var citesStatus *string
	if req.FishDataID != nil {
		var scientificName string
		scanErr := s.db.QueryRowContext(ctx,
			`SELECT scientific_name FROM fish_data WHERE id = $1`, *req.FishDataID,
		).Scan(&scientificName)
		if scanErr == nil && scientificName != "" {
			result, checkErr := s.citesRepo.CheckScientificName(ctx, scientificName)
			if checkErr == nil {
				var statusStr string
				switch {
				case result.IsBlocked:
					statusStr = "BLOCKED"
				case result.HasWarning:
					statusStr = "WARNING"
				default:
					statusStr = "CLEAR"
				}
				citesStatus = &statusStr
			}
		}
	}

	isAvailable := true
	if req.IsAvailable != nil {
		isAvailable = *req.IsAvailable
	}

	var item domain.ShopInventory
	err = s.db.QueryRowContext(ctx, `
		INSERT INTO shop_inventory (business_id, fish_data_id, custom_name, quantity, price, cites_status, is_available, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, NOW())
		RETURNING id, business_id, fish_data_id, custom_name, quantity, price, cites_status, is_available, updated_at
	`, businessID, req.FishDataID, nullableStr(req.CustomName), req.Quantity, req.Price, citesStatus, isAvailable,
	).Scan(
		&item.ID, &item.BusinessID, &item.FishDataID, &item.CustomName,
		&item.Quantity, &item.Price, &item.CitesStatus, &item.IsAvailable, &item.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}

	// fish_name 채우기
	if req.FishDataID != nil {
		s.db.QueryRowContext(ctx,
			`SELECT COALESCE(primary_common_name, '') FROM fish_data WHERE id = $1`, *req.FishDataID,
		).Scan(&item.FishName)
	} else {
		item.FishName = req.CustomName
	}

	return &item, nil
}

// UpdateInventory 재고 수정
func (s *InventoryService) UpdateInventory(ctx context.Context, businessID, itemID int64, userID string, req domain.ShopInventoryRequest) (*domain.ShopInventory, error) {
	ok, err := s.ownsBusinessProfile(ctx, businessID, userID)
	if err != nil {
		return nil, err
	}
	if !ok {
		return nil, fmt.Errorf("해당 업체를 수정할 권한이 없습니다")
	}

	// 아이템이 해당 업체 소속인지 확인
	var exists bool
	s.db.QueryRowContext(ctx,
		`SELECT EXISTS(SELECT 1 FROM shop_inventory WHERE id = $1 AND business_id = $2)`,
		itemID, businessID,
	).Scan(&exists)
	if !exists {
		return nil, fmt.Errorf("재고 항목을 찾을 수 없습니다")
	}

	// CITES 재체크 (fish_data_id가 있는 경우)
	var citesStatus *string
	if req.FishDataID != nil {
		var scientificName string
		scanErr := s.db.QueryRowContext(ctx,
			`SELECT scientific_name FROM fish_data WHERE id = $1`, *req.FishDataID,
		).Scan(&scientificName)
		if scanErr == nil && scientificName != "" {
			result, checkErr := s.citesRepo.CheckScientificName(ctx, scientificName)
			if checkErr == nil {
				var statusStr string
				switch {
				case result.IsBlocked:
					statusStr = "BLOCKED"
				case result.HasWarning:
					statusStr = "WARNING"
				default:
					statusStr = "CLEAR"
				}
				citesStatus = &statusStr
			}
		}
	}

	isAvailable := true
	if req.IsAvailable != nil {
		isAvailable = *req.IsAvailable
	}

	var item domain.ShopInventory
	err = s.db.QueryRowContext(ctx, `
		UPDATE shop_inventory
		SET fish_data_id = COALESCE($1, fish_data_id),
		    custom_name  = COALESCE($2, custom_name),
		    quantity     = $3,
		    price        = COALESCE($4, price),
		    cites_status = COALESCE($5, cites_status),
		    is_available = $6,
		    updated_at   = NOW()
		WHERE id = $7 AND business_id = $8
		RETURNING id, business_id, fish_data_id, custom_name, quantity, price, cites_status, is_available, updated_at
	`, req.FishDataID, nullableStr(req.CustomName), req.Quantity, req.Price, citesStatus, isAvailable, itemID, businessID,
	).Scan(
		&item.ID, &item.BusinessID, &item.FishDataID, &item.CustomName,
		&item.Quantity, &item.Price, &item.CitesStatus, &item.IsAvailable, &item.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}

	// fish_name 채우기
	if item.FishDataID != nil {
		s.db.QueryRowContext(ctx,
			`SELECT COALESCE(primary_common_name, '') FROM fish_data WHERE id = $1`, *item.FishDataID,
		).Scan(&item.FishName)
	} else if item.CustomName != nil {
		item.FishName = *item.CustomName
	}

	return &item, nil
}

// DeleteInventory 재고 항목 삭제
func (s *InventoryService) DeleteInventory(ctx context.Context, businessID, itemID int64, userID string) error {
	ok, err := s.ownsBusinessProfile(ctx, businessID, userID)
	if err != nil {
		return err
	}
	if !ok {
		return fmt.Errorf("해당 업체를 수정할 권한이 없습니다")
	}

	result, err := s.db.ExecContext(ctx,
		`DELETE FROM shop_inventory WHERE id = $1 AND business_id = $2`,
		itemID, businessID,
	)
	if err != nil {
		return err
	}
	n, _ := result.RowsAffected()
	if n == 0 {
		return fmt.Errorf("재고 항목을 찾을 수 없습니다")
	}
	return nil
}

// GetBusinessStats 업체 재고/리뷰 통계 조회
func (s *InventoryService) GetBusinessStats(ctx context.Context, businessID int64) (*domain.BusinessStats, error) {
	stats := &domain.BusinessStats{BusinessID: businessID}

	// 재고 통계
	s.db.QueryRowContext(ctx, `
		SELECT
			COUNT(*) AS total_inventory,
			COUNT(*) FILTER (WHERE is_available = TRUE) AS available_items,
			COALESCE(SUM(price * quantity) FILTER (WHERE is_available = TRUE), 0) AS total_value
		FROM shop_inventory
		WHERE business_id = $1
	`, businessID).Scan(&stats.TotalInventory, &stats.AvailableItems, &stats.TotalValue)

	// 리뷰 통계
	s.db.QueryRowContext(ctx, `
		SELECT COUNT(*), COALESCE(AVG(rating), 0)
		FROM business_reviews
		WHERE business_id = $1
	`, businessID).Scan(&stats.ReviewCount, &stats.AvgRating)

	return stats, nil
}

// nullableStr 빈 문자열이면 nil을 반환한다 (DB nullable 컬럼 대응).
func nullableStr(s string) *string {
	if s == "" {
		return nil
	}
	return &s
}
