package handler

import (
	"net/http"
	"strconv"

	"github.com/cashyalla/aquaverse/internal/api/middleware"
	"github.com/cashyalla/aquaverse/internal/service"
	"github.com/labstack/echo/v4"
)

type EscrowHandler struct {
	escrowSvc *service.EscrowService
}

func NewEscrowHandler(escrowSvc *service.EscrowService) *EscrowHandler {
	return &EscrowHandler{escrowSvc: escrowSvc}
}

// POST /api/v1/trades/:id/escrow/fund — 구매자 에스크로 입금
func (h *EscrowHandler) Fund(c echo.Context) error {
	tradeID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid trade id")
	}
	userID := middleware.MustGetUserID(c)
	if err := h.escrowSvc.FundEscrow(c.Request().Context(), tradeID, userID.String()); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	return c.JSON(http.StatusOK, map[string]string{"status": "funded"})
}

// POST /api/v1/trades/:id/escrow/release — 구매자 수령 확인 → 판매자 출금
func (h *EscrowHandler) Release(c echo.Context) error {
	tradeID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid trade id")
	}
	if err := h.escrowSvc.ReleaseEscrow(c.Request().Context(), tradeID); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}
	return c.JSON(http.StatusOK, map[string]string{"status": "released"})
}

// POST /api/v1/trades/:id/escrow/refund — 분쟁/취소 환불
func (h *EscrowHandler) Refund(c echo.Context) error {
	tradeID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid trade id")
	}
	var req struct {
		Reason string `json:"reason"`
	}
	c.Bind(&req)
	if err := h.escrowSvc.RefundEscrow(c.Request().Context(), tradeID, req.Reason); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}
	return c.JSON(http.StatusOK, map[string]string{"status": "refunded"})
}

// GET /api/v1/trades/:id/escrow — 에스크로 상태 조회
func (h *EscrowHandler) GetStatus(c echo.Context) error {
	tradeID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid trade id")
	}
	status, err := h.escrowSvc.GetEscrowStatus(c.Request().Context(), tradeID)
	if err != nil {
		return echo.NewHTTPError(http.StatusNotFound, "escrow not found")
	}
	return c.JSON(http.StatusOK, map[string]string{"escrow_status": status})
}
