package service

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/cashyalla/aquaverse/internal/domain"
	"github.com/gorilla/websocket"
	"github.com/jmoiron/sqlx"
)

// ─── AuctionHub ─────────────────────────────────────────────────────────────

// AuctionClient WebSocket 클라이언트 (경매방 참가자)
type AuctionClient struct {
	ConnID string
	Send   chan []byte
}

// AuctionHub 경매별 WebSocket 브로드캐스트 허브
// rooms: map[auctionID] → map[connID] → *AuctionClient
type AuctionHub struct {
	rooms map[int64]map[string]*AuctionClient
	mu    sync.RWMutex
}

func newAuctionHub() *AuctionHub {
	return &AuctionHub{
		rooms: make(map[int64]map[string]*AuctionClient),
	}
}

// Join 클라이언트를 경매방에 등록
func (h *AuctionHub) Join(auctionID int64, client *AuctionClient) {
	h.mu.Lock()
	defer h.mu.Unlock()
	if h.rooms[auctionID] == nil {
		h.rooms[auctionID] = make(map[string]*AuctionClient)
	}
	h.rooms[auctionID][client.ConnID] = client
}

// Leave 클라이언트를 경매방에서 제거
func (h *AuctionHub) Leave(auctionID int64, client *AuctionClient) {
	h.mu.Lock()
	defer h.mu.Unlock()
	if conns, ok := h.rooms[auctionID]; ok {
		delete(conns, client.ConnID)
		if len(conns) == 0 {
			delete(h.rooms, auctionID)
		}
	}
	close(client.Send)
}

// Broadcast 경매방 전체 연결에 메시지 전송
func (h *AuctionHub) Broadcast(auctionID int64, data []byte) {
	h.mu.RLock()
	defer h.mu.RUnlock()
	for _, client := range h.rooms[auctionID] {
		select {
		case client.Send <- data:
		default:
			// 슬로우 클라이언트 메시지 드롭
		}
	}
}

// ─── AuctionService ──────────────────────────────────────────────────────────

// AuctionService 실시간 경매 서비스
type AuctionService struct {
	db  *sqlx.DB
	hub *AuctionHub
	// 경매 종료 타이머 추적 (auctionID → timer)
	timers map[int64]*time.Timer
	timerMu sync.Mutex
}

// NewAuctionService 생성자. Hub를 내부적으로 초기화한다.
func NewAuctionService(db *sqlx.DB) *AuctionService {
	return &AuctionService{
		db:     db,
		hub:    newAuctionHub(),
		timers: make(map[int64]*time.Timer),
	}
}

// Hub AuctionHub 반환 (핸들러에서 WebSocket 등록에 사용)
func (s *AuctionService) Hub() *AuctionHub {
	return s.hub
}

// ─── CRUD ────────────────────────────────────────────────────────────────────

// ListAuctions 상태별 경매 목록 조회
// status: "live" | "upcoming" | "ended" | "" (전체)
func (s *AuctionService) ListAuctions(ctx context.Context, status string) ([]domain.Auction, error) {
	query := `
		SELECT
			a.id, a.listing_id, a.seller_id::text, a.title, a.description,
			COALESCE(a.image_url, '') AS image_url,
			a.start_price, a.current_price, a.reserve_price, a.bid_increment,
			a.starts_at, a.ends_at, a.status,
			a.winner_id::text, a.final_price, a.created_at,
			COUNT(b.id)::int AS bid_count
		FROM auctions a
		LEFT JOIN auction_bids b ON b.auction_id = a.id`

	args := []interface{}{}
	if status != "" {
		// "upcoming" 은 scheduled + 아직 starts_at 미도래
		switch status {
		case "upcoming":
			query += ` WHERE a.status = 'scheduled'`
		case "live":
			query += ` WHERE a.status = 'live'`
		case "ended":
			query += ` WHERE a.status IN ('ended', 'cancelled')`
		default:
			query += ` WHERE a.status = $1`
			args = append(args, status)
		}
	}
	query += `
		GROUP BY a.id
		ORDER BY a.ends_at ASC`

	rows, err := s.db.QueryxContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var auctions []domain.Auction
	for rows.Next() {
		var a domain.Auction
		var sellerID, winnerID *string
		var imageURL *string
		if err := rows.Scan(
			&a.ID, &a.ListingID, &sellerID, &a.Title, &a.Description,
			&imageURL,
			&a.StartPrice, &a.CurrentPrice, &a.ReservePrice, &a.BidIncrement,
			&a.StartsAt, &a.EndsAt, &a.Status,
			&winnerID, &a.FinalPrice, &a.CreatedAt,
			&a.BidCount,
		); err != nil {
			return nil, err
		}
		if sellerID != nil {
			a.SellerID = *sellerID
		}
		if winnerID != nil {
			a.WinnerID = winnerID
		}
		if imageURL != nil {
			a.ImageURL = *imageURL
		}
		auctions = append(auctions, a)
	}
	return auctions, rows.Err()
}

// GetAuction 경매 상세 + 최근 입찰 5개
func (s *AuctionService) GetAuction(ctx context.Context, id int64) (*domain.AuctionDetail, error) {
	var a domain.Auction
	err := s.db.QueryRowContext(ctx, `
		SELECT
			a.id, a.listing_id, a.seller_id::text, a.title, a.description,
			COALESCE(a.image_url, '') AS image_url,
			a.start_price, a.current_price, a.reserve_price, a.bid_increment,
			a.starts_at, a.ends_at, a.status,
			a.winner_id::text, a.final_price, a.created_at,
			COUNT(b.id)::int AS bid_count
		FROM auctions a
		LEFT JOIN auction_bids b ON b.auction_id = a.id
		WHERE a.id = $1
		GROUP BY a.id
	`, id).Scan(
		&a.ID, &a.ListingID, &a.SellerID, &a.Title, &a.Description,
		&a.ImageURL,
		&a.StartPrice, &a.CurrentPrice, &a.ReservePrice, &a.BidIncrement,
		&a.StartsAt, &a.EndsAt, &a.Status,
		&a.WinnerID, &a.FinalPrice, &a.CreatedAt,
		&a.BidCount,
	)
	if err != nil {
		return nil, fmt.Errorf("auction not found: %w", err)
	}

	// 최근 입찰 5개
	rows, err := s.db.QueryxContext(ctx, `
		SELECT id, auction_id, bidder_id::text, amount, is_winning, bid_at
		FROM auction_bids
		WHERE auction_id = $1
		ORDER BY bid_at DESC
		LIMIT 5
	`, id)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var bids []domain.AuctionBid
	for rows.Next() {
		var b domain.AuctionBid
		if err := rows.Scan(&b.ID, &b.AuctionID, &b.BidderID, &b.Amount, &b.IsWinning, &b.BidAt); err != nil {
			return nil, err
		}
		bids = append(bids, b)
	}

	return &domain.AuctionDetail{Auction: a, RecentBids: bids}, rows.Err()
}

// CreateAuction 새 경매 생성
func (s *AuctionService) CreateAuction(ctx context.Context, sellerID string, req domain.CreateAuctionRequest) (*domain.Auction, error) {
	if req.Title == "" {
		return nil, errors.New("title is required")
	}
	if req.StartPrice <= 0 {
		return nil, errors.New("start_price must be positive")
	}
	if req.BidIncrement <= 0 {
		req.BidIncrement = 1000
	}

	startsAt, err := time.Parse(time.RFC3339, req.StartsAt)
	if err != nil {
		return nil, errors.New("invalid starts_at: must be RFC3339")
	}
	endsAt, err := time.Parse(time.RFC3339, req.EndsAt)
	if err != nil {
		return nil, errors.New("invalid ends_at: must be RFC3339")
	}
	if !endsAt.After(startsAt) {
		return nil, errors.New("ends_at must be after starts_at")
	}

	var auctionID int64
	err = s.db.QueryRowContext(ctx, `
		INSERT INTO auctions
			(listing_id, seller_id, title, description, image_url,
			 start_price, reserve_price, bid_increment, starts_at, ends_at, status)
		VALUES
			($1, $2::uuid, $3, $4, $5, $6, $7, $8, $9, $10, 'scheduled')
		RETURNING id
	`,
		req.ListingID, sellerID, req.Title, req.Description, req.ImageURL,
		req.StartPrice, req.ReservePrice, req.BidIncrement, startsAt, endsAt,
	).Scan(&auctionID)
	if err != nil {
		return nil, err
	}

	// 경매 시작 / 종료 자동 예약
	s.scheduleAuction(auctionID, startsAt, endsAt)

	auction := &domain.Auction{
		ID:           auctionID,
		ListingID:    req.ListingID,
		SellerID:     sellerID,
		Title:        req.Title,
		Description:  req.Description,
		ImageURL:     req.ImageURL,
		StartPrice:   req.StartPrice,
		ReservePrice: req.ReservePrice,
		BidIncrement: req.BidIncrement,
		StartsAt:     startsAt,
		EndsAt:       endsAt,
		Status:       string(domain.AuctionStatusScheduled),
		BidCount:     0,
	}
	return auction, nil
}

// PlaceBid 입찰 처리 (트랜잭션)
func (s *AuctionService) PlaceBid(ctx context.Context, auctionID int64, bidderID string, amount int64) (*domain.AuctionBid, error) {
	tx, err := s.db.BeginTxx(ctx, nil)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	// 현재 경매 상태 및 최고가 잠금 조회 (FOR UPDATE)
	var status string
	var currentPrice *int64
	var bidIncrement int64
	var endsAt time.Time
	err = tx.QueryRowContext(ctx, `
		SELECT status, current_price, bid_increment, ends_at
		FROM auctions
		WHERE id = $1
		FOR UPDATE
	`, auctionID).Scan(&status, &currentPrice, &bidIncrement, &endsAt)
	if err != nil {
		return nil, fmt.Errorf("auction not found: %w", err)
	}

	if status != string(domain.AuctionStatusLive) {
		return nil, errors.New("auction is not live")
	}
	if time.Now().After(endsAt) {
		return nil, errors.New("auction has already ended")
	}

	// 최소 입찰가 계산
	var minBid int64
	if currentPrice != nil {
		minBid = *currentPrice + bidIncrement
	} else {
		// 아직 아무도 입찰하지 않은 경우 시작가 로드
		var startPrice int64
		tx.QueryRowContext(ctx, `SELECT start_price FROM auctions WHERE id = $1`, auctionID).Scan(&startPrice)
		minBid = startPrice
	}

	if amount < minBid {
		return nil, fmt.Errorf("bid amount must be at least %d", minBid)
	}

	// 이전 최고 입찰자 낙찰 취소
	if _, err := tx.ExecContext(ctx, `
		UPDATE auction_bids SET is_winning = FALSE
		WHERE auction_id = $1 AND is_winning = TRUE
	`, auctionID); err != nil {
		return nil, err
	}

	// 새 입찰 삽입
	var bidID int64
	var bidAt time.Time
	err = tx.QueryRowContext(ctx, `
		INSERT INTO auction_bids (auction_id, bidder_id, amount, is_winning)
		VALUES ($1, $2::uuid, $3, TRUE)
		RETURNING id, bid_at
	`, auctionID, bidderID, amount).Scan(&bidID, &bidAt)
	if err != nil {
		return nil, err
	}

	// auctions.current_price 갱신
	if _, err := tx.ExecContext(ctx, `
		UPDATE auctions SET current_price = $1 WHERE id = $2
	`, amount, auctionID); err != nil {
		return nil, err
	}

	if err := tx.Commit(); err != nil {
		return nil, err
	}

	bid := &domain.AuctionBid{
		ID:        bidID,
		AuctionID: auctionID,
		BidderID:  bidderID,
		Amount:    amount,
		IsWinning: true,
		BidAt:     bidAt,
	}

	// 현재 입찰 수 조회 (브로드캐스트용)
	var bidCount int
	s.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM auction_bids WHERE auction_id = $1`, auctionID).Scan(&bidCount)

	// WebSocket 브로드캐스트
	s.BroadcastBid(auctionID, domain.AuctionWSMessage{
		Type:         "bid",
		CurrentPrice: amount,
		BidderID:     bidderID,
		BidCount:     bidCount,
	})

	return bid, nil
}

// EndAuction 경매 종료 처리 (winner 결정, status=ended)
func (s *AuctionService) EndAuction(ctx context.Context, auctionID int64) error {
	// 기존 타이머 정리
	s.timerMu.Lock()
	if t, ok := s.timers[auctionID]; ok {
		t.Stop()
		delete(s.timers, auctionID)
	}
	s.timerMu.Unlock()

	// 최고 낙찰자 조회
	var winnerID *string
	var finalPrice *int64
	s.db.QueryRowContext(ctx, `
		SELECT bidder_id::text, amount
		FROM auction_bids
		WHERE auction_id = $1 AND is_winning = TRUE
		LIMIT 1
	`, auctionID).Scan(&winnerID, &finalPrice)

	_, err := s.db.ExecContext(ctx, `
		UPDATE auctions
		SET status = 'ended', winner_id = $2::uuid, final_price = $3
		WHERE id = $1 AND status IN ('live', 'scheduled')
	`, auctionID, winnerIDOrNull(winnerID), finalPrice)
	if err != nil {
		return err
	}

	// 최종 현재가 조회 (브로드캐스트용)
	var currentPrice int64
	var bidCount int
	s.db.QueryRowContext(ctx, `
		SELECT COALESCE(current_price, start_price), (SELECT COUNT(*) FROM auction_bids WHERE auction_id = $1)
		FROM auctions WHERE id = $1
	`, auctionID).Scan(&currentPrice, &bidCount)

	// WebSocket 브로드캐스트: 경매 종료
	msg := domain.AuctionWSMessage{
		Type:         "end",
		CurrentPrice: currentPrice,
		BidCount:     bidCount,
		FinalPrice:   finalPrice,
	}
	if winnerID != nil {
		msg.WinnerID = *winnerID
	}
	s.BroadcastBid(auctionID, msg)

	return nil
}

// BroadcastBid 경매방 전체에 메시지 브로드캐스트
func (s *AuctionService) BroadcastBid(auctionID int64, msg domain.AuctionWSMessage) {
	data, err := json.Marshal(msg)
	if err != nil {
		return
	}
	s.hub.Broadcast(auctionID, data)
}

// ─── 내부 헬퍼 ───────────────────────────────────────────────────────────────

// scheduleAuction starts_at에 상태를 live로, ends_at에 EndAuction을 예약
func (s *AuctionService) scheduleAuction(auctionID int64, startsAt, endsAt time.Time) {
	now := time.Now()

	startDelay := startsAt.Sub(now)
	endDelay := endsAt.Sub(now)

	// 이미 starts_at이 지났으면 즉시 live로 전환
	if startDelay <= 0 {
		s.markLive(auctionID)
	} else {
		time.AfterFunc(startDelay, func() {
			s.markLive(auctionID)
		})
	}

	if endDelay <= 0 {
		// 이미 ends_at이 지났으면 즉시 종료
		go s.EndAuction(context.Background(), auctionID)
		return
	}

	s.timerMu.Lock()
	defer s.timerMu.Unlock()
	s.timers[auctionID] = time.AfterFunc(endDelay, func() {
		s.EndAuction(context.Background(), auctionID)
	})
}

// markLive 경매 상태를 scheduled → live 로 변경
func (s *AuctionService) markLive(auctionID int64) {
	ctx := context.Background()
	_, err := s.db.ExecContext(ctx, `
		UPDATE auctions SET status = 'live' WHERE id = $1 AND status = 'scheduled'
	`, auctionID)
	if err != nil {
		return
	}
	s.BroadcastBid(auctionID, domain.AuctionWSMessage{
		Type: "status",
	})
}

// winnerIDOrNull *string → interface{} (nil-safe, SQL NULL 처리)
func winnerIDOrNull(s *string) interface{} {
	if s == nil {
		return nil
	}
	return *s
}

// NewAuctionWSClient 새 WebSocket 클라이언트 생성 헬퍼 (핸들러에서 사용)
func NewAuctionWSClient(connID string) *AuctionClient {
	return &AuctionClient{
		ConnID: connID,
		Send:   make(chan []byte, 256),
	}
}

// WriteLoop WebSocket 쓰기 루프. conn이 닫히거나 Send 채널이 닫히면 반환.
func WriteLoop(conn *websocket.Conn, client *AuctionClient) {
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()
	for {
		select {
		case data, ok := <-client.Send:
			conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
			if !ok {
				conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}
			if err := conn.WriteMessage(websocket.TextMessage, data); err != nil {
				return
			}
		case <-ticker.C:
			conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
			if err := conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}
