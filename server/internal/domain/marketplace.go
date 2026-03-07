package domain

import (
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

// HealthStatus 건강 상태
type HealthStatus string

const (
	HealthExcellent      HealthStatus = "EXCELLENT"
	HealthGood           HealthStatus = "GOOD"
	HealthDiseaseHistory HealthStatus = "DISEASE_HISTORY"
	HealthUnderTreatment HealthStatus = "UNDER_TREATMENT"
)

// TradeType 거래 방식
type TradeType string

const (
	TradeTypeDirect     TradeType = "DIRECT"      // 직거래
	TradeTypeCourier    TradeType = "COURIER"      // 일반 택배
	TradeTypeAquaCourier TradeType = "AQUA_COURIER" // 생물 전문 배송
	TradeTypeAll        TradeType = "ALL"
)

// ListingStatus 분양글 상태
type ListingStatus string

const (
	ListingStatusDraft    ListingStatus = "DRAFT"
	ListingStatusActive   ListingStatus = "ACTIVE"
	ListingStatusReserved ListingStatus = "RESERVED"
	ListingStatusSold     ListingStatus = "SOLD"
	ListingStatusExpired  ListingStatus = "EXPIRED"
	ListingStatusHidden   ListingStatus = "HIDDEN"
	ListingStatusDeleted  ListingStatus = "DELETED"
)

// TradeStatus 거래 진행 상태
type TradeStatus string

const (
	TradeStatusNegotiating TradeStatus = "NEGOTIATING"
	TradeStatusConfirmed   TradeStatus = "CONFIRMED"
	TradeStatusInDelivery  TradeStatus = "IN_DELIVERY"
	TradeStatusDelivered   TradeStatus = "DELIVERED"
	TradeStatusCompleted   TradeStatus = "COMPLETED"
	TradeStatusCancelled   TradeStatus = "CANCELLED"
	TradeStatusDisputed    TradeStatus = "DISPUTED"
)

// Sex 성별
type Sex string

const (
	SexMale    Sex = "MALE"
	SexFemale  Sex = "FEMALE"
	SexUnknown Sex = "UNKNOWN"
	SexMixed   Sex = "MIXED"
)

// Listing 분양 매물
type Listing struct {
	ID       int64     `db:"id"`
	SellerID uuid.UUID `db:"seller_id"`

	// 어종 정보
	FishDataID     *int64  `db:"fish_data_id"`
	ScientificName *string `db:"scientific_name"`
	CommonName     string  `db:"common_name"`

	// 개체 정보
	Quantity       int          `db:"quantity"`
	AgeMonths      *int         `db:"age_months"`
	SizeCm         *float64     `db:"size_cm"`
	Sex            Sex          `db:"sex"`
	HealthStatus   HealthStatus `db:"health_status"`
	DiseaseHistory *string      `db:"disease_history"`
	BredBySeller   bool         `db:"bred_by_seller"`

	// 수조 환경
	TankSizeLiters *int     `db:"tank_size_liters"`
	TankMates      []string `db:"tank_mates"` // JSONB
	FeedingType    *string  `db:"feeding_type"`
	WaterPH        *float64 `db:"water_ph"`
	WaterTempC     *float64 `db:"water_temp_c"`

	// 거래 조건
	Price           decimal.Decimal `db:"price"`
	PriceUSD        *decimal.Decimal `db:"price_usd"`
	Currency        string           `db:"currency"`
	PriceNegotiable bool             `db:"price_negotiable"`
	TradeType       TradeType        `db:"trade_type"`

	// 국제 거래
	AllowInternational bool     `db:"allow_international"`
	AllowedCountries   []string `db:"allowed_countries"` // JSONB

	// 위치 (PostGIS)
	Latitude     *float64 `db:"latitude"`
	Longitude    *float64 `db:"longitude"`
	LocationText string   `db:"location_text"` // 표시용: "서울 강남구"
	CountryCode  string   `db:"country_code"`

	// 콘텐츠
	Title       string   `db:"title"`
	Description *string  `db:"description"`
	ImageURLs   []string `db:"image_urls"` // JSONB

	// 상태
	Status        ListingStatus `db:"status"`
	ViewCount     int           `db:"view_count"`
	FavoriteCount int           `db:"favorite_count"`

	// 사기 방지
	AutoHold   bool    `db:"auto_hold"`
	HoldReason *string `db:"hold_reason"`

	CreatedAt time.Time  `db:"created_at"`
	UpdatedAt time.Time  `db:"updated_at"`
	ExpiresAt *time.Time `db:"expires_at"`
	SoldAt    *time.Time `db:"sold_at"`

	// 조인
	Seller      *UserProfile `db:"-"`
	DistanceKm  *float64     `db:"-"` // 거리 계산 결과
}

// Trade 거래
type Trade struct {
	ID        int64     `db:"id"`
	ListingID int64     `db:"listing_id"`
	SellerID  uuid.UUID `db:"seller_id"`
	BuyerID   uuid.UUID `db:"buyer_id"`

	TradeType    TradeType       `db:"trade_type"`
	AgreedPrice  decimal.Decimal `db:"agreed_price"`
	Currency     string          `db:"currency"`

	// 안전결제
	EscrowEnabled bool    `db:"escrow_enabled"`
	EscrowStatus  *string `db:"escrow_status"`
	PaymentMethod *string `db:"payment_method"`
	PaymentRef    *string `db:"payment_ref"`

	// 배송
	TrackingNumber *string `db:"tracking_number"`
	CourierName    *string `db:"courier_name"`
	DeliveryNotes  *string `db:"delivery_notes"`

	Status TradeStatus `db:"status"`

	// 직거래 위치 확인
	MeetupLatitude          *float64 `db:"meetup_latitude"`
	MeetupLongitude         *float64 `db:"meetup_longitude"`
	MeetupConfirmedSeller   bool     `db:"meetup_confirmed_seller"`
	MeetupConfirmedBuyer    bool     `db:"meetup_confirmed_buyer"`

	// 생물 도착 확인
	ArrivalConfirmedAt *time.Time `db:"arrival_confirmed_at"`
	ArrivalPhotoURLs   []string   `db:"arrival_photo_urls"` // JSONB
	HealthConfirmed    *bool      `db:"health_confirmed"`

	// 분쟁
	DisputedAt        *time.Time `db:"disputed_at"`
	DisputeReason     *string    `db:"dispute_reason"`
	DisputeResolvedAt *time.Time `db:"dispute_resolved_at"`
	AdminNote         *string    `db:"admin_note"`

	CreatedAt   time.Time  `db:"created_at"`
	UpdatedAt   time.Time  `db:"updated_at"`
	CompletedAt *time.Time `db:"completed_at"`
}

// TradeReview 거래 리뷰
type TradeReview struct {
	ID          int64     `db:"id"`
	TradeID     int64     `db:"trade_id"`
	ReviewerID  uuid.UUID `db:"reviewer_id"`
	RevieweeID  uuid.UUID `db:"reviewee_id"`

	Rating              float64 `db:"rating"`
	RatingCommunication *int    `db:"rating_communication"`
	RatingAccuracy      *int    `db:"rating_accuracy"`
	RatingPackaging     *int    `db:"rating_packaging"`
	RatingHealth        *int    `db:"rating_health"`

	Comment *string  `db:"comment"`
	Tags    []string `db:"tags"` // JSONB

	CreatedAt time.Time `db:"created_at"`
}

// UserTrustScore 신뢰도 집계
type UserTrustScore struct {
	UserID          uuid.UUID `db:"user_id"`
	TrustScore      float64   `db:"trust_score"`
	TotalTrades     int       `db:"total_trades"`
	CompletedTrades int       `db:"completed_trades"`
	AvgRating       *float64  `db:"avg_rating"`
	ResponseRate    *float64  `db:"response_rate"`
	Badges          []string  `db:"badges"` // JSONB
	FraudReports    int       `db:"fraud_report_count"`
	ConfirmedFrauds int       `db:"confirmed_fraud_count"`
	UpdatedAt       time.Time `db:"updated_at"`
}

// FishWatchSubscription 어종 알림 구독
type FishWatchSubscription struct {
	ID                  int64     `db:"id"`
	UserID              uuid.UUID `db:"user_id"`
	FishDataID          *int64    `db:"fish_data_id"`
	CustomSpecies       *string   `db:"custom_species"`
	MaxPrice            *float64  `db:"max_price"`
	Latitude            *float64  `db:"latitude"`
	Longitude           *float64  `db:"longitude"`
	RadiusKm            float64   `db:"radius_km"`
	IncludeInternational bool     `db:"include_international"`
	NotifyPush          bool      `db:"notify_push"`
	NotifyEmail         bool      `db:"notify_email"`
	NotifyInApp         bool      `db:"notify_in_app"`
	Active              bool      `db:"active"`
	LastNotifiedAt      *time.Time `db:"last_notified_at"`
	MatchCount          int       `db:"match_count"`
	CreatedAt           time.Time  `db:"created_at"`
}

// FraudReport 사기 신고
type FraudReport struct {
	ID             int64      `db:"id"`
	ReporterID     uuid.UUID  `db:"reporter_id"`
	ReportedUserID uuid.UUID  `db:"reported_user_id"`
	ListingID      *int64     `db:"listing_id"`
	TradeID        *int64     `db:"trade_id"`

	ReportType   string   `db:"report_type"` // FRAUD, NO_SHOW, DECEASED_FISH, WRONG_SPECIES
	Description  *string  `db:"description"`
	EvidenceURLs []string `db:"evidence_urls"` // JSONB

	Status    string     `db:"status"` // PENDING, UNDER_REVIEW, CONFIRMED, REJECTED
	AdminNote *string    `db:"admin_note"`

	CreatedAt   time.Time  `db:"created_at"`
	ResolvedAt  *time.Time `db:"resolved_at"`
}

// ListingFilter 분양글 목록 필터
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

// ListingListResult 분양글 목록 결과
type ListingListResult struct {
	Items      []Listing `json:"items"`
	TotalCount int       `json:"total_count"`
	Page       int       `json:"page"`
	Limit      int       `json:"limit"`
}
