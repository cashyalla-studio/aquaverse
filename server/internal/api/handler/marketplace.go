package handler

import (
	"net/http"
	"strconv"

	"github.com/cashyalla/aquaverse/internal/api/middleware"
	"github.com/cashyalla/aquaverse/internal/service"
	"github.com/labstack/echo/v4"
)

type MarketplaceHandler struct {
	mktSvc *service.MarketplaceService
}

func NewMarketplaceHandler(mktSvc *service.MarketplaceService) *MarketplaceHandler {
	return &MarketplaceHandler{mktSvc: mktSvc}
}

// GET /api/v1/listings
// Query: lat, lng, radius_km, fish_id, min_price, max_price, trade_type, page, limit
func (h *MarketplaceHandler) ListListings(c echo.Context) error {
	filter := domain.ListingFilter{
		Page:  1,
		Limit: 20,
	}

	if lat := c.QueryParam("lat"); lat != "" {
		v, _ := strconv.ParseFloat(lat, 64)
		filter.Lat = &v
	}
	if lng := c.QueryParam("lng"); lng != "" {
		v, _ := strconv.ParseFloat(lng, 64)
		filter.Lng = &v
	}
	if r := c.QueryParam("radius_km"); r != "" {
		v, _ := strconv.ParseFloat(r, 64)
		filter.RadiusKm = &v
	}
	if fishID := c.QueryParam("fish_id"); fishID != "" {
		id, _ := strconv.ParseInt(fishID, 10, 64)
		filter.FishDataID = &id
	}
	if p := c.QueryParam("page"); p != "" {
		filter.Page, _ = strconv.Atoi(p)
	}
	if l := c.QueryParam("limit"); l != "" {
		filter.Limit, _ = strconv.Atoi(l)
	}
	filter.TradeType = c.QueryParam("trade_type")
	filter.Search = c.QueryParam("q")

	result, err := h.mktSvc.ListListings(c.Request().Context(), filter)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}
	return c.JSON(http.StatusOK, result)
}

// GET /api/v1/listings/:id
func (h *MarketplaceHandler) GetListing(c echo.Context) error {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid listing id")
	}
	listing, err := h.mktSvc.GetListing(c.Request().Context(), id)
	if err != nil {
		return echo.NewHTTPError(http.StatusNotFound, "listing not found")
	}
	return c.JSON(http.StatusOK, listing)
}

// POST /api/v1/listings
func (h *MarketplaceHandler) CreateListing(c echo.Context) error {
	userID := c.Get(middleware.ContextKeyUserID).(string)

	var req service.CreateListingRequest
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	req.SellerID = userID

	listing, err := h.mktSvc.CreateListing(c.Request().Context(), req)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	return c.JSON(http.StatusCreated, listing)
}

// PUT /api/v1/listings/:id/status
func (h *MarketplaceHandler) UpdateListingStatus(c echo.Context) error {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid listing id")
	}
	userID := c.Get(middleware.ContextKeyUserID).(string)

	var req struct {
		Status string `json:"status"`
	}
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}

	if err := h.mktSvc.UpdateListingStatus(c.Request().Context(), id, userID, req.Status); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	return c.JSON(http.StatusOK, map[string]string{"status": req.Status})
}

// POST /api/v1/listings/:id/trade
func (h *MarketplaceHandler) InitiateTrade(c echo.Context) error {
	listingID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid listing id")
	}
	buyerID := c.Get(middleware.ContextKeyUserID).(string)

	var req service.InitiateTradeRequest
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	req.ListingID = listingID
	req.BuyerID = buyerID

	trade, err := h.mktSvc.InitiateTrade(c.Request().Context(), req)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	return c.JSON(http.StatusCreated, trade)
}

// PUT /api/v1/trades/:id/status
func (h *MarketplaceHandler) UpdateTradeStatus(c echo.Context) error {
	tradeID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid trade id")
	}
	userID := c.Get(middleware.ContextKeyUserID).(string)

	var req service.UpdateTradeStatusRequest
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	req.TradeID = tradeID
	req.UserID = userID

	if err := h.mktSvc.UpdateTradeStatus(c.Request().Context(), req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	return c.NoContent(http.StatusOK)
}

// POST /api/v1/trades/:id/review
func (h *MarketplaceHandler) SubmitReview(c echo.Context) error {
	tradeID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid trade id")
	}
	reviewerID := c.Get(middleware.ContextKeyUserID).(string)

	var req service.SubmitReviewRequest
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	req.TradeID = tradeID
	req.ReviewerID = reviewerID

	if err := h.mktSvc.SubmitReview(c.Request().Context(), req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	return c.NoContent(http.StatusCreated)
}

// POST /api/v1/listings/:id/watch
func (h *MarketplaceHandler) WatchFish(c echo.Context) error {
	userID := c.Get(middleware.ContextKeyUserID).(string)

	var req service.WatchFishRequest
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	req.UserID = userID

	sub, err := h.mktSvc.SubscribeWatch(c.Request().Context(), req)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	return c.JSON(http.StatusCreated, sub)
}

// POST /api/v1/fraud-reports
func (h *MarketplaceHandler) ReportFraud(c echo.Context) error {
	reporterID := c.Get(middleware.ContextKeyUserID).(string)

	var req service.FraudReportRequest
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	req.ReporterID = reporterID

	if err := h.mktSvc.ReportFraud(c.Request().Context(), req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	return c.NoContent(http.StatusCreated)
}
