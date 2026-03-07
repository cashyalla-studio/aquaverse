package service

import (
	"context"
	"errors"

	"github.com/cashyalla/aquaverse/internal/domain"
	"github.com/shopspring/decimal"
)

type ListingFilter struct {
	Lat        *float64
	Lng        *float64
	RadiusKm   *float64
	FishDataID *int64
	MinPrice   *decimal.Decimal
	MaxPrice   *decimal.Decimal
	TradeType  string
	Search     string
	Status     string
	Page       int
	Limit      int
}

type ListingListResult struct {
	Items      []domain.Listing `json:"items"`
	TotalCount int              `json:"total_count"`
	Page       int              `json:"page"`
	Limit      int              `json:"limit"`
}

type CreateListingRequest struct {
	SellerID        string
	FishDataID      *int64           `json:"fish_data_id,omitempty"`
	ScientificName  *string          `json:"scientific_name,omitempty"`
	CommonName      string           `json:"common_name"`
	Quantity        int              `json:"quantity"`
	AgeMonths       *int             `json:"age_months,omitempty"`
	SizeCm          *float64         `json:"size_cm,omitempty"`
	Sex             domain.Sex       `json:"sex"`
	HealthStatus    domain.HealthStatus `json:"health_status"`
	DiseaseHistory  *string          `json:"disease_history,omitempty"`
	BredBySeller    bool             `json:"bred_by_seller"`
	TankSizeLiters  *int             `json:"tank_size_liters,omitempty"`
	TankMates       []string         `json:"tank_mates,omitempty"`
	FeedingType     *string          `json:"feeding_type,omitempty"`
	WaterPH         *float64         `json:"water_ph,omitempty"`
	WaterTempC      *float64         `json:"water_temp_c,omitempty"`
	Price           decimal.Decimal  `json:"price"`
	Currency        string           `json:"currency"`
	PriceNegotiable bool             `json:"price_negotiable"`
	TradeType       domain.TradeType `json:"trade_type"`
	AllowInternational bool          `json:"allow_international"`
	AllowedCountries   []string      `json:"allowed_countries,omitempty"`
	Latitude        *float64         `json:"latitude,omitempty"`
	Longitude       *float64         `json:"longitude,omitempty"`
	LocationText    string           `json:"location_text"`
	CountryCode     string           `json:"country_code"`
	Title           string           `json:"title"`
	Description     *string          `json:"description,omitempty"`
	ImageURLs       []string         `json:"image_urls"`
}

type InitiateTradeRequest struct {
	ListingID     int64
	BuyerID       string
	TradeType     domain.TradeType `json:"trade_type"`
	EscrowEnabled bool             `json:"escrow_enabled"`
	MeetupLat     *float64         `json:"meetup_lat,omitempty"`
	MeetupLng     *float64         `json:"meetup_lng,omitempty"`
}

type UpdateTradeStatusRequest struct {
	TradeID        int64
	UserID         string
	Status         domain.TradeStatus `json:"status"`
	TrackingNumber *string            `json:"tracking_number,omitempty"`
	CourierName    *string            `json:"courier_name,omitempty"`
	ArrivalPhotos  []string           `json:"arrival_photos,omitempty"`
	HealthConfirmed *bool             `json:"health_confirmed,omitempty"`
	DisputeReason  *string            `json:"dispute_reason,omitempty"`
}

type SubmitReviewRequest struct {
	TradeID             int64
	ReviewerID          string
	Rating              float64 `json:"rating"`
	RatingCommunication *int    `json:"rating_communication,omitempty"`
	RatingAccuracy      *int    `json:"rating_accuracy,omitempty"`
	RatingPackaging     *int    `json:"rating_packaging,omitempty"`
	RatingHealth        *int    `json:"rating_health,omitempty"`
	Comment             *string `json:"comment,omitempty"`
	Tags                []string `json:"tags,omitempty"`
}

type WatchFishRequest struct {
	UserID               string
	FishDataID           *int64   `json:"fish_data_id,omitempty"`
	CustomSpecies        *string  `json:"custom_species,omitempty"`
	MaxPrice             *float64 `json:"max_price,omitempty"`
	Latitude             *float64 `json:"latitude,omitempty"`
	Longitude            *float64 `json:"longitude,omitempty"`
	RadiusKm             float64  `json:"radius_km"`
	IncludeInternational bool     `json:"include_international"`
	NotifyPush           bool     `json:"notify_push"`
	NotifyEmail          bool     `json:"notify_email"`
}

type FraudReportRequest struct {
	ReporterID     string
	ReportedUserID string   `json:"reported_user_id"`
	ListingID      *int64   `json:"listing_id,omitempty"`
	TradeID        *int64   `json:"trade_id,omitempty"`
	ReportType     string   `json:"report_type"`
	Description    *string  `json:"description,omitempty"`
	EvidenceURLs   []string `json:"evidence_urls,omitempty"`
}

type MarketplaceRepository interface {
	ListListings(ctx context.Context, filter ListingFilter) ([]domain.Listing, int, error)
	GetListing(ctx context.Context, id int64) (*domain.Listing, error)
	CreateListing(ctx context.Context, listing *domain.Listing) error
	UpdateListingStatus(ctx context.Context, id int64, status domain.ListingStatus) error
	GetListingBySeller(ctx context.Context, id int64, sellerID string) (*domain.Listing, error)
	CreateTrade(ctx context.Context, trade *domain.Trade) error
	GetTrade(ctx context.Context, id int64) (*domain.Trade, error)
	UpdateTrade(ctx context.Context, trade *domain.Trade) error
	CreateReview(ctx context.Context, review *domain.TradeReview) error
	UpdateTrustScore(ctx context.Context, userID string) error
	CreateWatchSubscription(ctx context.Context, sub *domain.FishWatchSubscription) error
	CreateFraudReport(ctx context.Context, report *domain.FraudReport) error
	GetFraudCountByUser(ctx context.Context, userID string) (int, error)
}

type MarketplaceNotifier interface {
	NotifyNewListing(ctx context.Context, listing *domain.Listing)
	NotifyTradeUpdate(ctx context.Context, trade *domain.Trade, targetUserID string)
}

type FraudDetector interface {
	CheckListing(ctx context.Context, req CreateListingRequest) (bool, string)
}

type MarketplaceService struct {
	repo     MarketplaceRepository
	notifier MarketplaceNotifier
	fraud    FraudDetector
}

func NewMarketplaceService(
	repo MarketplaceRepository,
	notifier MarketplaceNotifier,
	fraud FraudDetector,
) *MarketplaceService {
	return &MarketplaceService{repo: repo, notifier: notifier, fraud: fraud}
}

func (s *MarketplaceService) ListListings(ctx context.Context, filter ListingFilter) (*ListingListResult, error) {
	if filter.Page < 1 {
		filter.Page = 1
	}
	if filter.Limit < 1 || filter.Limit > 100 {
		filter.Limit = 20
	}
	if filter.Status == "" {
		filter.Status = string(domain.ListingStatusActive)
	}

	items, total, err := s.repo.ListListings(ctx, filter)
	if err != nil {
		return nil, err
	}
	return &ListingListResult{Items: items, TotalCount: total, Page: filter.Page, Limit: filter.Limit}, nil
}

func (s *MarketplaceService) GetListing(ctx context.Context, id int64) (*domain.Listing, error) {
	return s.repo.GetListing(ctx, id)
}

func (s *MarketplaceService) CreateListing(ctx context.Context, req CreateListingRequest) (*domain.Listing, error) {
	if req.CommonName == "" {
		return nil, errors.New("common_name is required")
	}
	if req.Quantity < 1 {
		return nil, errors.New("quantity must be at least 1")
	}
	if len(req.ImageURLs) == 0 {
		return nil, errors.New("at least 1 image is required")
	}

	// 사기 방지 검사
	isHold, holdReason := s.fraud.CheckListing(ctx, req)

	listing := &domain.Listing{
		CommonName:         req.CommonName,
		FishDataID:         req.FishDataID,
		ScientificName:     req.ScientificName,
		Quantity:           req.Quantity,
		AgeMonths:          req.AgeMonths,
		SizeCm:             req.SizeCm,
		Sex:                req.Sex,
		HealthStatus:       req.HealthStatus,
		DiseaseHistory:     req.DiseaseHistory,
		BredBySeller:       req.BredBySeller,
		TankSizeLiters:     req.TankSizeLiters,
		TankMates:          req.TankMates,
		FeedingType:        req.FeedingType,
		WaterPH:            req.WaterPH,
		WaterTempC:         req.WaterTempC,
		Price:              req.Price,
		Currency:           req.Currency,
		PriceNegotiable:    req.PriceNegotiable,
		TradeType:          req.TradeType,
		AllowInternational: req.AllowInternational,
		AllowedCountries:   req.AllowedCountries,
		Latitude:           req.Latitude,
		Longitude:          req.Longitude,
		LocationText:       req.LocationText,
		CountryCode:        req.CountryCode,
		Title:              req.Title,
		Description:        req.Description,
		ImageURLs:          req.ImageURLs,
		Status:             domain.ListingStatusActive,
		AutoHold:           isHold,
	}
	if isHold {
		listing.Status = domain.ListingStatusHidden
		listing.HoldReason = &holdReason
	}

	if err := s.repo.CreateListing(ctx, listing); err != nil {
		return nil, err
	}

	// 알림 구독자에게 비동기 알림
	go s.notifier.NotifyNewListing(context.Background(), listing)

	return listing, nil
}

func (s *MarketplaceService) UpdateListingStatus(ctx context.Context, id int64, userID, status string) error {
	listing, err := s.repo.GetListingBySeller(ctx, id, userID)
	if err != nil {
		return errors.New("listing not found or unauthorized")
	}
	_ = listing
	return s.repo.UpdateListingStatus(ctx, id, domain.ListingStatus(status))
}

func (s *MarketplaceService) InitiateTrade(ctx context.Context, req InitiateTradeRequest) (*domain.Trade, error) {
	listing, err := s.repo.GetListing(ctx, req.ListingID)
	if err != nil {
		return nil, errors.New("listing not found")
	}
	if listing.Status != domain.ListingStatusActive {
		return nil, errors.New("listing is not available")
	}

	trade := &domain.Trade{
		ListingID:     req.ListingID,
		BuyerID:       req.BuyerID,
		AgreedPrice:   listing.Price,
		Currency:      listing.Currency,
		TradeType:     req.TradeType,
		EscrowEnabled: req.EscrowEnabled,
		Status:        domain.TradeStatusNegotiating,
	}
	if err := s.repo.CreateTrade(ctx, trade); err != nil {
		return nil, err
	}
	return trade, nil
}

func (s *MarketplaceService) UpdateTradeStatus(ctx context.Context, req UpdateTradeStatusRequest) error {
	trade, err := s.repo.GetTrade(ctx, req.TradeID)
	if err != nil {
		return errors.New("trade not found")
	}
	trade.Status = req.Status
	if req.TrackingNumber != nil {
		trade.TrackingNumber = req.TrackingNumber
	}
	if req.CourierName != nil {
		trade.CourierName = req.CourierName
	}
	return s.repo.UpdateTrade(ctx, trade)
}

func (s *MarketplaceService) SubmitReview(ctx context.Context, req SubmitReviewRequest) error {
	if req.Rating < 1 || req.Rating > 5 {
		return errors.New("rating must be between 1 and 5")
	}
	review := &domain.TradeReview{
		TradeID:             req.TradeID,
		Rating:              req.Rating,
		RatingCommunication: req.RatingCommunication,
		RatingAccuracy:      req.RatingAccuracy,
		RatingPackaging:     req.RatingPackaging,
		RatingHealth:        req.RatingHealth,
		Comment:             req.Comment,
		Tags:                req.Tags,
	}
	if err := s.repo.CreateReview(ctx, review); err != nil {
		return err
	}
	// 신뢰도 점수 갱신
	go func() {
		_ = s.repo.UpdateTrustScore(context.Background(), req.ReviewerID)
	}()
	return nil
}

func (s *MarketplaceService) SubscribeWatch(ctx context.Context, req WatchFishRequest) (*domain.FishWatchSubscription, error) {
	if req.FishDataID == nil && req.CustomSpecies == nil {
		return nil, errors.New("fish_data_id or custom_species is required")
	}
	sub := &domain.FishWatchSubscription{
		FishDataID:           req.FishDataID,
		CustomSpecies:        req.CustomSpecies,
		MaxPrice:             req.MaxPrice,
		Latitude:             req.Latitude,
		Longitude:            req.Longitude,
		RadiusKm:             req.RadiusKm,
		IncludeInternational: req.IncludeInternational,
		NotifyPush:           req.NotifyPush,
		NotifyEmail:          req.NotifyEmail,
		NotifyInApp:          true,
		Active:               true,
	}
	if err := s.repo.CreateWatchSubscription(ctx, sub); err != nil {
		return nil, err
	}
	return sub, nil
}

func (s *MarketplaceService) ReportFraud(ctx context.Context, req FraudReportRequest) error {
	report := &domain.FraudReport{
		ReportType:   req.ReportType,
		Description:  req.Description,
		EvidenceURLs: req.EvidenceURLs,
		ListingID:    req.ListingID,
		TradeID:      req.TradeID,
		Status:       "PENDING",
	}
	return s.repo.CreateFraudReport(ctx, report)
}
