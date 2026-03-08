package domain

import "time"

// AuctionStatus 경매 상태
type AuctionStatus string

const (
	AuctionStatusScheduled AuctionStatus = "scheduled"
	AuctionStatusLive      AuctionStatus = "live"
	AuctionStatusEnded     AuctionStatus = "ended"
	AuctionStatusCancelled AuctionStatus = "cancelled"
)

// Auction 실시간 경매
type Auction struct {
	ID           int64     `db:"id"            json:"id"`
	ListingID    *int64    `db:"listing_id"    json:"listing_id,omitempty"`
	SellerID     string    `db:"seller_id"     json:"seller_id"`
	Title        string    `db:"title"         json:"title"`
	Description  string    `db:"description"   json:"description"`
	ImageURL     string    `db:"image_url"     json:"image_url"`
	StartPrice   int64     `db:"start_price"   json:"start_price"`
	CurrentPrice *int64    `db:"current_price" json:"current_price"`
	ReservePrice *int64    `db:"reserve_price" json:"reserve_price,omitempty"`
	BidIncrement int64     `db:"bid_increment" json:"bid_increment"`
	StartsAt     time.Time `db:"starts_at"     json:"starts_at"`
	EndsAt       time.Time `db:"ends_at"       json:"ends_at"`
	Status       string    `db:"status"        json:"status"`
	WinnerID     *string   `db:"winner_id"     json:"winner_id,omitempty"`
	FinalPrice   *int64    `db:"final_price"   json:"final_price,omitempty"`
	CreatedAt    time.Time `db:"created_at"    json:"created_at"`
	BidCount     int       `db:"bid_count"     json:"bid_count"`
}

// AuctionBid 경매 입찰
type AuctionBid struct {
	ID        int64     `db:"id"         json:"id"`
	AuctionID int64     `db:"auction_id" json:"auction_id"`
	BidderID  string    `db:"bidder_id"  json:"bidder_id"`
	Amount    int64     `db:"amount"     json:"amount"`
	IsWinning bool      `db:"is_winning" json:"is_winning"`
	BidAt     time.Time `db:"bid_at"     json:"bid_at"`
}

// AuctionDetail 경매 상세 (최근 입찰 포함)
type AuctionDetail struct {
	Auction
	RecentBids []AuctionBid `json:"recent_bids"`
}

// CreateAuctionRequest 경매 생성 요청
type CreateAuctionRequest struct {
	ListingID    *int64  `json:"listing_id,omitempty"`
	Title        string  `json:"title"`
	Description  string  `json:"description"`
	ImageURL     string  `json:"image_url"`
	StartPrice   int64   `json:"start_price"`
	ReservePrice *int64  `json:"reserve_price,omitempty"`
	BidIncrement int64   `json:"bid_increment"`
	StartsAt     string  `json:"starts_at"` // RFC3339
	EndsAt       string  `json:"ends_at"`   // RFC3339
}

// PlaceBidRequest 입찰 요청
type PlaceBidRequest struct {
	Amount int64 `json:"amount"`
}

// AuctionWSMessage WebSocket 브로드캐스트 메시지
type AuctionWSMessage struct {
	Type         string `json:"type"`                    // "bid", "end", "status"
	CurrentPrice int64  `json:"current_price"`
	BidderID     string `json:"bidder_id,omitempty"`
	BidCount     int    `json:"bid_count"`
	EndsAt       string `json:"ends_at,omitempty"`
	WinnerID     string `json:"winner_id,omitempty"`
	FinalPrice   *int64 `json:"final_price,omitempty"`
}
