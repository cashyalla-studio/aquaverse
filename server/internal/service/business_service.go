package service

import (
	"context"
	"github.com/cashyalla/aquaverse/internal/domain"
	"github.com/jmoiron/sqlx"
)

type BusinessService struct {
	db *sqlx.DB
}

func NewBusinessService(db *sqlx.DB) *BusinessService {
	return &BusinessService{db: db}
}

// CreateProfile 업체 프로필 등록
func (s *BusinessService) CreateProfile(ctx context.Context, userID string, profile domain.BusinessProfile) (*domain.BusinessProfile, error) {
	profile.UserID = userID
	var id int64
	err := s.db.QueryRowContext(ctx, `
        INSERT INTO business_profiles (user_id, store_name, description, address, city, phone, website, logo_url, lat, lng)
        VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10)
        RETURNING id`,
		userID, profile.StoreName, profile.Description, profile.Address,
		profile.City, profile.Phone, profile.Website, profile.LogoURL,
		profile.Lat, profile.Lng,
	).Scan(&id)
	if err != nil {
		return nil, err
	}
	// account_type 업데이트
	s.db.ExecContext(ctx, `UPDATE users SET account_type='BUSINESS' WHERE id=$1`, userID)
	profile.ID = id
	return &profile, nil
}

// GetProfile 업체 프로필 조회
func (s *BusinessService) GetProfile(ctx context.Context, businessID int64) (*domain.BusinessProfile, error) {
	var p domain.BusinessProfile
	err := s.db.GetContext(ctx, &p, `
        SELECT bp.*,
               COALESCE(AVG(br.rating),0) as avg_rating,
               COUNT(br.id) as review_count
        FROM business_profiles bp
        LEFT JOIN business_reviews br ON br.business_id = bp.id
        WHERE bp.id=$1
        GROUP BY bp.id
    `, businessID)
	return &p, err
}

// UpdateProfile 업체 프로필 수정
func (s *BusinessService) UpdateProfile(ctx context.Context, businessID int64, userID string, patch domain.BusinessProfile) (*domain.BusinessProfile, error) {
	_, err := s.db.ExecContext(ctx, `
        UPDATE business_profiles
        SET store_name=COALESCE(NULLIF($1,''),store_name),
            description=COALESCE(NULLIF($2,''),description),
            address=COALESCE(NULLIF($3,''),address),
            city=COALESCE(NULLIF($4,''),city),
            phone=COALESCE(NULLIF($5,''),phone),
            website=COALESCE(NULLIF($6,''),website),
            logo_url=COALESCE(NULLIF($7,''),logo_url),
            updated_at=NOW()
        WHERE id=$8 AND user_id=$9
    `, patch.StoreName, patch.Description, patch.Address, patch.City,
		patch.Phone, patch.Website, patch.LogoURL, businessID, userID)
	if err != nil {
		return nil, err
	}
	return s.GetProfile(ctx, businessID)
}

// ListBusinesses 업체 목록 (도시 필터 또는 전체)
func (s *BusinessService) ListBusinesses(ctx context.Context, city string, limit, offset int) ([]domain.BusinessProfile, error) {
	if limit <= 0 || limit > 50 {
		limit = 20
	}
	var profiles []domain.BusinessProfile
	var err error
	if city != "" {
		err = s.db.SelectContext(ctx, &profiles, `
            SELECT bp.*,
                   COALESCE(AVG(br.rating),0) as avg_rating,
                   COUNT(br.id) as review_count
            FROM business_profiles bp
            LEFT JOIN business_reviews br ON br.business_id = bp.id
            WHERE bp.city ILIKE $1
            GROUP BY bp.id
            ORDER BY bp.is_verified DESC, avg_rating DESC
            LIMIT $2 OFFSET $3
        `, "%"+city+"%", limit, offset)
	} else {
		err = s.db.SelectContext(ctx, &profiles, `
            SELECT bp.*,
                   COALESCE(AVG(br.rating),0) as avg_rating,
                   COUNT(br.id) as review_count
            FROM business_profiles bp
            LEFT JOIN business_reviews br ON br.business_id = bp.id
            GROUP BY bp.id
            ORDER BY bp.is_verified DESC, avg_rating DESC
            LIMIT $1 OFFSET $2
        `, limit, offset)
	}
	return profiles, err
}

// NearbyBusinesses PostGIS ST_DWithin으로 반경 내 업체 검색
func (s *BusinessService) NearbyBusinesses(ctx context.Context, lat, lng, radiusKm float64, limit int) ([]domain.BusinessProfile, error) {
	if limit <= 0 || limit > 50 {
		limit = 20
	}
	var profiles []domain.BusinessProfile
	err := s.db.SelectContext(ctx, &profiles, `
        SELECT bp.*,
               COALESCE(AVG(br.rating),0) as avg_rating,
               COUNT(br.id) as review_count,
               ST_Distance(
                   ST_MakePoint(bp.lng, bp.lat)::geography,
                   ST_MakePoint($2, $1)::geography
               ) / 1000.0 as distance_km
        FROM business_profiles bp
        LEFT JOIN business_reviews br ON br.business_id = bp.id
        WHERE bp.lat IS NOT NULL AND bp.lng IS NOT NULL
          AND ST_DWithin(
              ST_MakePoint(bp.lng, bp.lat)::geography,
              ST_MakePoint($2, $1)::geography,
              $3 * 1000
          )
        GROUP BY bp.id
        ORDER BY distance_km ASC
        LIMIT $4
    `, lat, lng, radiusKm, limit)
	return profiles, err
}

// AddReview 리뷰 작성
func (s *BusinessService) AddReview(ctx context.Context, businessID int64, reviewerID string, rating int, content string) (*domain.BusinessReview, error) {
	var id int64
	err := s.db.QueryRowContext(ctx, `
        INSERT INTO business_reviews (business_id, reviewer_id, rating, content)
        VALUES ($1,$2,$3,$4)
        ON CONFLICT (business_id, reviewer_id) DO UPDATE SET rating=$3, content=$4
        RETURNING id`,
		businessID, reviewerID, rating, content,
	).Scan(&id)
	if err != nil {
		return nil, err
	}
	return &domain.BusinessReview{
		ID: id, BusinessID: businessID, ReviewerID: reviewerID,
		Rating: rating, Content: content,
	}, nil
}

// GetReviews 업체 리뷰 목록
func (s *BusinessService) GetReviews(ctx context.Context, businessID int64) ([]domain.BusinessReview, error) {
	var reviews []domain.BusinessReview
	err := s.db.SelectContext(ctx, &reviews, `
        SELECT br.*, u.username as reviewer_name
        FROM business_reviews br
        JOIN users u ON u.id = br.reviewer_id
        WHERE br.business_id=$1
        ORDER BY br.created_at DESC
        LIMIT 50
    `, businessID)
	return reviews, err
}
