package marketplace

import (
	"context"
	"time"

	"github.com/cashyalla/aquaverse/internal/service"
)

// FraudDetector 사기 방지 시스템
type FraudDetector struct {
	cache ListingCountCache
}

type ListingCountCache interface {
	GetRecentListingCount(ctx context.Context, sellerID string) (int64, error)
	IncrRecentListingCount(ctx context.Context, sellerID string) error
}

func NewFraudDetector(cache ListingCountCache) *FraudDetector {
	return &FraudDetector{cache: cache}
}

// CheckListing 분양글 등록 전 사기 가능성 검사
// returns: (auto_hold, hold_reason)
func (d *FraudDetector) CheckListing(ctx context.Context, req service.CreateListingRequest) (bool, string) {
	// 1. 동일 판매자 24시간 내 3건 이상 등록
	count, _ := d.cache.GetRecentListingCount(ctx, req.SellerID)
	if count >= 3 {
		return true, "too_many_listings_24h"
	}

	// 2. 비정상적으로 낮은 가격 감지 (가격이 0 초과인 경우만)
	// 실제 구현에서는 fish_data 테이블의 평균 시세와 비교
	priceFloat, _ := req.Price.Float64()
	if priceFloat > 0 && priceFloat < 100 && req.CommonName != "" {
		// 유료 분양인데 가격이 100원 미만 → 의심
		// 실제로는 어종별 평균 시세 DB와 비교
		return true, "suspiciously_low_price"
	}

	// 3. 치료 중인 생물 판매 경고 (홀드는 아님, 경고만)
	if req.HealthStatus == "UNDER_TREATMENT" {
		// 강제 홀드 아님, 경고 표시만 (실제 구현에서 분양글에 경고 뱃지 표시)
		_ = true
	}

	// 등록 횟수 증가
	_ = d.cache.IncrRecentListingCount(ctx, req.SellerID)

	return false, ""
}

// SeasonalShippingWarning 계절별 배송 경고
func SeasonalShippingWarning() string {
	month := time.Now().Month()
	switch {
	case month == 12 || month <= 2:
		return "winter_shipping_risk"
	case month >= 7 && month <= 8:
		return "summer_shipping_risk"
	}
	return ""
}
