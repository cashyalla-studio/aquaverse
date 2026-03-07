package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"

	"github.com/cashyalla/aquaverse/internal/domain"
	"github.com/cashyalla/aquaverse/internal/service"
	"github.com/jmoiron/sqlx"
)

type MarketplaceRepository struct {
	db *sqlx.DB
}

func NewMarketplaceRepository(db *sqlx.DB) *MarketplaceRepository {
	return &MarketplaceRepository{db: db}
}

func (r *MarketplaceRepository) ListListings(ctx context.Context, filter service.ListingFilter) ([]domain.Listing, int, error) {
	args := []interface{}{}
	where := []string{fmt.Sprintf("l.status = $%d", 1)}
	args = append(args, filter.Status)
	idx := 2

	if filter.FishDataID != nil {
		where = append(where, fmt.Sprintf("l.fish_data_id = $%d", idx))
		args = append(args, *filter.FishDataID)
		idx++
	}
	if filter.TradeType != "" {
		where = append(where, fmt.Sprintf("(l.trade_type = $%d OR l.trade_type = 'ALL')", idx))
		args = append(args, filter.TradeType)
		idx++
	}
	if filter.Search != "" {
		where = append(where, fmt.Sprintf("(l.common_name ILIKE $%d OR l.scientific_name ILIKE $%d)", idx, idx+1))
		like := "%" + filter.Search + "%"
		args = append(args, like, like)
		idx += 2
	}

	// 위치 기반 필터 (PostGIS)
	distanceSelect := "NULL::float8 AS distance_km"
	if filter.Lat != nil && filter.Lng != nil && filter.RadiusKm != nil {
		distanceSelect = fmt.Sprintf(
			"ST_Distance(l.location, ST_MakePoint($%d, $%d)::geography) / 1000 AS distance_km",
			idx, idx+1,
		)
		where = append(where, fmt.Sprintf(
			"(l.location IS NULL OR ST_DWithin(l.location, ST_MakePoint($%d, $%d)::geography, $%d * 1000))",
			idx, idx+1, idx+2,
		))
		args = append(args, *filter.Lng, *filter.Lat, *filter.RadiusKm)
		idx += 3
	}

	whereClause := "WHERE " + strings.Join(where, " AND ")

	var total int
	cq := fmt.Sprintf("SELECT COUNT(*) FROM listings l %s", whereClause)
	if err := r.db.GetContext(ctx, &total, cq, args...); err != nil {
		return nil, 0, err
	}

	offset := (filter.Page - 1) * filter.Limit
	orderBy := "l.created_at DESC"
	if filter.Lat != nil {
		orderBy = "distance_km ASC NULLS LAST, l.created_at DESC"
	}

	q := fmt.Sprintf(`
		SELECT
			l.id, l.seller_id, l.fish_data_id, l.scientific_name, l.common_name,
			l.quantity, l.age_months, l.size_cm, l.sex, l.health_status,
			l.price, l.currency, l.price_negotiable, l.trade_type,
			l.allow_international, l.location_text, l.country_code,
			l.title, l.description, l.image_urls, l.status,
			l.view_count, l.favorite_count, l.bred_by_seller,
			l.created_at, l.updated_at, l.expires_at,
			%s
		FROM listings l
		%s
		ORDER BY %s
		LIMIT $%d OFFSET $%d
	`, distanceSelect, whereClause, orderBy, idx, idx+1)

	args = append(args, filter.Limit, offset)

	var listings []domain.Listing
	rows, err := r.db.QueryxContext(ctx, q, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	for rows.Next() {
		var l domain.Listing
		var distKm *float64
		if err := rows.Scan(
			&l.ID, &l.SellerID, &l.FishDataID, &l.ScientificName, &l.CommonName,
			&l.Quantity, &l.AgeMonths, &l.SizeCm, &l.Sex, &l.HealthStatus,
			&l.Price, &l.Currency, &l.PriceNegotiable, &l.TradeType,
			&l.AllowInternational, &l.LocationText, &l.CountryCode,
			&l.Title, &l.Description, nil, &l.Status,
			&l.ViewCount, &l.FavoriteCount, &l.BredBySeller,
			&l.CreatedAt, &l.UpdatedAt, &l.ExpiresAt,
			&distKm,
		); err != nil {
			continue
		}
		l.DistanceKm = distKm
		listings = append(listings, l)
	}

	return listings, total, nil
}

func (r *MarketplaceRepository) GetListing(ctx context.Context, id int64) (*domain.Listing, error) {
	var l domain.Listing
	q := `
		SELECT
			l.id, l.seller_id, l.fish_data_id, l.scientific_name, l.common_name,
			l.quantity, l.age_months, l.size_cm, l.sex, l.health_status,
			l.disease_history, l.bred_by_seller,
			l.tank_size_liters, l.feeding_type, l.water_ph, l.water_temp_c,
			l.price, l.price_usd, l.currency, l.price_negotiable, l.trade_type,
			l.allow_international, l.location_text, l.country_code,
			l.title, l.description, l.status,
			l.view_count, l.favorite_count,
			l.created_at, l.updated_at, l.expires_at, l.sold_at
		FROM listings l
		WHERE l.id = $1 AND l.status NOT IN ('DELETED')
	`
	if err := r.db.GetContext(ctx, &l, q, id); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, errors.New("listing not found")
		}
		return nil, err
	}

	// 조회수 비동기 증가
	go r.db.ExecContext(context.Background(), `UPDATE listings SET view_count = view_count + 1 WHERE id = $1`, id)

	return &l, nil
}

func (r *MarketplaceRepository) GetListingBySeller(ctx context.Context, id int64, sellerID string) (*domain.Listing, error) {
	var l domain.Listing
	q := `SELECT id, seller_id, status FROM listings WHERE id = $1 AND seller_id = $2`
	if err := r.db.GetContext(ctx, &l, q, id, sellerID); err != nil {
		return nil, errors.New("listing not found or unauthorized")
	}
	return &l, nil
}

func (r *MarketplaceRepository) CreateListing(ctx context.Context, l *domain.Listing) error {
	q := `
		INSERT INTO listings (
			seller_id, fish_data_id, scientific_name, common_name,
			quantity, age_months, size_cm, sex, health_status, disease_history, bred_by_seller,
			tank_size_liters, feeding_type, water_ph, water_temp_c,
			price, currency, price_negotiable, trade_type,
			allow_international, location_text, country_code,
			title, description, image_urls,
			status, auto_hold, hold_reason
		) VALUES (
			$1, $2, $3, $4,
			$5, $6, $7, $8, $9, $10, $11,
			$12, $13, $14, $15,
			$16, $17, $18, $19,
			$20, $21, $22,
			$23, $24, '[]'::jsonb,
			$25, $26, $27
		)
		RETURNING id, created_at, updated_at
	`
	return r.db.QueryRowContext(ctx, q,
		l.SellerID, l.FishDataID, l.ScientificName, l.CommonName,
		l.Quantity, l.AgeMonths, l.SizeCm, l.Sex, l.HealthStatus, l.DiseaseHistory, l.BredBySeller,
		l.TankSizeLiters, l.FeedingType, l.WaterPH, l.WaterTempC,
		l.Price, l.Currency, l.PriceNegotiable, l.TradeType,
		l.AllowInternational, l.LocationText, l.CountryCode,
		l.Title, l.Description,
		l.Status, l.AutoHold, l.HoldReason,
	).Scan(&l.ID, &l.CreatedAt, &l.UpdatedAt)
}

func (r *MarketplaceRepository) UpdateListingStatus(ctx context.Context, id int64, status domain.ListingStatus) error {
	q := `UPDATE listings SET status = $1, updated_at = NOW() WHERE id = $2`
	_, err := r.db.ExecContext(ctx, q, string(status), id)
	return err
}

func (r *MarketplaceRepository) CreateTrade(ctx context.Context, t *domain.Trade) error {
	q := `
		INSERT INTO trades (listing_id, seller_id, buyer_id, trade_type, agreed_price, currency, escrow_enabled, status)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		RETURNING id, created_at, updated_at
	`
	return r.db.QueryRowContext(ctx, q,
		t.ListingID, t.SellerID, t.BuyerID, t.TradeType,
		t.AgreedPrice, t.Currency, t.EscrowEnabled, t.Status,
	).Scan(&t.ID, &t.CreatedAt, &t.UpdatedAt)
}

func (r *MarketplaceRepository) GetTrade(ctx context.Context, id int64) (*domain.Trade, error) {
	var t domain.Trade
	q := `SELECT * FROM trades WHERE id = $1`
	if err := r.db.GetContext(ctx, &t, q, id); err != nil {
		return nil, errors.New("trade not found")
	}
	return &t, nil
}

func (r *MarketplaceRepository) UpdateTrade(ctx context.Context, t *domain.Trade) error {
	q := `
		UPDATE trades SET
			status = $1, tracking_number = $2, courier_name = $3,
			arrival_confirmed_at = $4, health_confirmed = $5,
			disputed_at = $6, dispute_reason = $7,
			meetup_confirmed_seller = $8, meetup_confirmed_buyer = $9,
			updated_at = NOW()
		WHERE id = $10
	`
	_, err := r.db.ExecContext(ctx, q,
		t.Status, t.TrackingNumber, t.CourierName,
		t.ArrivalConfirmedAt, t.HealthConfirmed,
		t.DisputedAt, t.DisputeReason,
		t.MeetupConfirmedSeller, t.MeetupConfirmedBuyer,
		t.ID,
	)
	return err
}

func (r *MarketplaceRepository) CreateReview(ctx context.Context, review *domain.TradeReview) error {
	q := `
		INSERT INTO trade_reviews (trade_id, reviewer_id, reviewee_id, rating,
			rating_communication, rating_accuracy, rating_packaging, rating_health, comment)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
		RETURNING id, created_at
	`
	return r.db.QueryRowContext(ctx, q,
		review.TradeID, review.ReviewerID, review.RevieweeID, review.Rating,
		review.RatingCommunication, review.RatingAccuracy, review.RatingPackaging, review.RatingHealth,
		review.Comment,
	).Scan(&review.ID, &review.CreatedAt)
}

func (r *MarketplaceRepository) UpdateTrustScore(ctx context.Context, userID string) error {
	_, err := r.db.ExecContext(ctx, `SELECT update_trust_score($1)`, userID)
	return err
}

func (r *MarketplaceRepository) CreateWatchSubscription(ctx context.Context, sub *domain.FishWatchSubscription) error {
	q := `
		INSERT INTO fish_watch_subscriptions
			(user_id, fish_data_id, custom_species, max_price, radius_km, include_international, notify_push, notify_email, notify_in_app)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
		ON CONFLICT (user_id, fish_data_id, custom_species) DO UPDATE SET
			max_price = EXCLUDED.max_price,
			radius_km = EXCLUDED.radius_km,
			active = TRUE
		RETURNING id, created_at
	`
	return r.db.QueryRowContext(ctx, q,
		sub.UserID, sub.FishDataID, sub.CustomSpecies, sub.MaxPrice,
		sub.RadiusKm, sub.IncludeInternational,
		sub.NotifyPush, sub.NotifyEmail, sub.NotifyInApp,
	).Scan(&sub.ID, &sub.CreatedAt)
}

func (r *MarketplaceRepository) CreateFraudReport(ctx context.Context, report *domain.FraudReport) error {
	q := `
		INSERT INTO fraud_reports (reporter_id, reported_user_id, listing_id, trade_id, report_type, description, status)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		RETURNING id, created_at
	`
	return r.db.QueryRowContext(ctx, q,
		report.ReporterID, report.ReportedUserID, report.ListingID, report.TradeID,
		report.ReportType, report.Description, report.Status,
	).Scan(&report.ID, &report.CreatedAt)
}

func (r *MarketplaceRepository) GetFraudCountByUser(ctx context.Context, userID string) (int, error) {
	var count int
	q := `SELECT COUNT(*) FROM fraud_reports WHERE reported_user_id = $1 AND status = 'CONFIRMED'`
	return count, r.db.GetContext(ctx, &count, q, userID)
}

func (r *MarketplaceRepository) FindMatchingWatchSubscriptions(ctx context.Context, listing *domain.Listing) ([]domain.FishWatchSubscription, error) {
	q := `
		SELECT *
		FROM fish_watch_subscriptions
		WHERE active = TRUE
		  AND (fish_data_id = $1 OR custom_species ILIKE $2)
		  AND (max_price IS NULL OR max_price >= $3)
		  AND (location IS NULL OR radius_km IS NULL
		       OR ST_DWithin(location, ST_MakePoint($4, $5)::geography, radius_km * 1000))
		  AND (include_international = TRUE OR $6 = 'KR')
	`
	var subs []domain.FishWatchSubscription
	lat := 0.0
	lng := 0.0
	if listing.Latitude != nil {
		lat = *listing.Latitude
	}
	if listing.Longitude != nil {
		lng = *listing.Longitude
	}
	err := r.db.SelectContext(ctx, &subs, q,
		listing.FishDataID, listing.CommonName,
		listing.Price, lng, lat, listing.CountryCode,
	)
	return subs, err
}
