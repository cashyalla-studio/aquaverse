package handler

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/cashyalla/aquaverse/internal/api/middleware"
	"github.com/cashyalla/aquaverse/internal/domain"
	"github.com/cashyalla/aquaverse/internal/service"
	"github.com/google/uuid"
	"github.com/gorilla/websocket"
	"github.com/labstack/echo/v4"
)

// auctionUpgrader WebSocket 업그레이더 (채팅과 동일한 설정)
var auctionUpgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true // TODO: 프로덕션에서 origin 검증
	},
}

// AuctionHandler REST + WebSocket 핸들러
type AuctionHandler struct {
	svc *service.AuctionService
}

// NewAuctionHandler 생성자
func NewAuctionHandler(svc *service.AuctionService) *AuctionHandler {
	return &AuctionHandler{svc: svc}
}

// ─── REST 엔드포인트 ──────────────────────────────────────────────────────────

// ListAuctions GET /api/v1/auctions?status=live|upcoming|ended
func (h *AuctionHandler) ListAuctions(c echo.Context) error {
	status := c.QueryParam("status")
	auctions, err := h.svc.ListAuctions(c.Request().Context(), status)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}
	return c.JSON(http.StatusOK, map[string]interface{}{
		"items": auctions,
	})
}

// GetAuction GET /api/v1/auctions/:id
func (h *AuctionHandler) GetAuction(c echo.Context) error {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid auction id")
	}

	detail, err := h.svc.GetAuction(c.Request().Context(), id)
	if err != nil {
		return echo.NewHTTPError(http.StatusNotFound, "auction not found")
	}
	return c.JSON(http.StatusOK, detail)
}

// CreateAuction POST /api/v1/auctions (인증 필요)
func (h *AuctionHandler) CreateAuction(c echo.Context) error {
	sellerID := middleware.MustGetUserID(c)
	if sellerID == uuid.Nil {
		return echo.NewHTTPError(http.StatusUnauthorized, "authentication required")
	}

	var req domain.CreateAuctionRequest
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}

	auction, err := h.svc.CreateAuction(c.Request().Context(), sellerID.String(), req)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	return c.JSON(http.StatusCreated, auction)
}

// PlaceBid POST /api/v1/auctions/:id/bid (인증 필요)
func (h *AuctionHandler) PlaceBid(c echo.Context) error {
	auctionID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid auction id")
	}

	bidderID := middleware.MustGetUserID(c)
	if bidderID == uuid.Nil {
		return echo.NewHTTPError(http.StatusUnauthorized, "authentication required")
	}

	var req domain.PlaceBidRequest
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	if req.Amount <= 0 {
		return echo.NewHTTPError(http.StatusBadRequest, "amount must be positive")
	}

	bid, err := h.svc.PlaceBid(c.Request().Context(), auctionID, bidderID.String(), req.Amount)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	return c.JSON(http.StatusCreated, bid)
}

// EndAuction POST /api/v1/auctions/:id/end (판매자 또는 관리자)
func (h *AuctionHandler) EndAuction(c echo.Context) error {
	auctionID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid auction id")
	}

	callerID := middleware.MustGetUserID(c)
	if callerID == uuid.Nil {
		return echo.NewHTTPError(http.StatusUnauthorized, "authentication required")
	}
	role, _ := c.Get(middleware.ContextKeyUserRole).(string)

	// 판매자 본인이거나 관리자만 허용
	detail, err := h.svc.GetAuction(c.Request().Context(), auctionID)
	if err != nil {
		return echo.NewHTTPError(http.StatusNotFound, "auction not found")
	}
	if detail.SellerID != callerID.String() && role != "admin" {
		return echo.NewHTTPError(http.StatusForbidden, "only the seller or admin can end this auction")
	}

	if err := h.svc.EndAuction(c.Request().Context(), auctionID); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}
	return c.JSON(http.StatusOK, map[string]string{"status": "ended"})
}

// ─── WebSocket 엔드포인트 ─────────────────────────────────────────────────────

// Connect GET /api/v1/auctions/:id/ws — 실시간 WebSocket 연결
//
// 클라이언트 → 서버 메시지:
//
//	{ "amount": 15000 }   → PlaceBid 호출
//
// 서버 → 클라이언트 브로드캐스트:
//
//	AuctionWSMessage (type: "bid" | "end" | "status")
func (h *AuctionHandler) Connect(c echo.Context) error {
	auctionID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid auction id")
	}

	// 경매 존재 여부 확인
	detail, err := h.svc.GetAuction(c.Request().Context(), auctionID)
	if err != nil {
		return echo.NewHTTPError(http.StatusNotFound, "auction not found")
	}

	// 선택적 인증 (관람은 미인증 허용, 입찰은 인증 필요)
	callerID := middleware.MustGetUserID(c)

	ws, err := auctionUpgrader.Upgrade(c.Response(), c.Request(), nil)
	if err != nil {
		return err
	}
	defer ws.Close()

	connID := uuid.New().String()
	client := service.NewAuctionWSClient(connID)
	hub := h.svc.Hub()
	hub.Join(auctionID, client)
	defer hub.Leave(auctionID, client)

	// 연결 즉시 현재 경매 상태를 해당 클라이언트에게만 전송
	initialMsg := domain.AuctionWSMessage{
		Type:     "status",
		BidCount: detail.BidCount,
		EndsAt:   detail.EndsAt.Format("2006-01-02T15:04:05Z07:00"),
	}
	if detail.CurrentPrice != nil {
		initialMsg.CurrentPrice = *detail.CurrentPrice
	} else {
		initialMsg.CurrentPrice = detail.StartPrice
	}
	if initData, merr := json.Marshal(initialMsg); merr == nil {
		select {
		case client.Send <- initData:
		default:
		}
	}

	// 쓰기 goroutine
	done := make(chan struct{})
	go func() {
		defer close(done)
		service.WriteLoop(ws, client)
	}()

	// 읽기 루프 (메인 goroutine)
	ws.SetReadLimit(512)
	ws.SetPongHandler(func(string) error {
		return nil
	})

	for {
		_, raw, err := ws.ReadMessage()
		if err != nil {
			break
		}

		// 인증된 사용자만 입찰 허용
		if callerID == uuid.Nil {
			continue
		}

		var bidReq domain.PlaceBidRequest
		if jsonErr := json.Unmarshal(raw, &bidReq); jsonErr != nil || bidReq.Amount <= 0 {
			continue
		}

		// 입찰 처리 — 오류는 개별 클라이언트에게만 전송
		if _, bidErr := h.svc.PlaceBid(c.Request().Context(), auctionID, callerID.String(), bidReq.Amount); bidErr != nil {
			errData, _ := json.Marshal(map[string]string{
				"type":  "error",
				"error": bidErr.Error(),
			})
			select {
			case client.Send <- errData:
			default:
			}
		}
	}

	<-done
	return nil
}
