package handler

import (
	"net/http"
	"strconv"

	"github.com/cashyalla/aquaverse/internal/api/middleware"
	"github.com/cashyalla/aquaverse/internal/domain"
	"github.com/cashyalla/aquaverse/internal/service"
	"github.com/labstack/echo/v4"
)

type PaymentHandler struct {
	svc *service.PaymentService
}

func NewPaymentHandler(svc *service.PaymentService) *PaymentHandler {
	return &PaymentHandler{svc: svc}
}

// InitiatePayment POST /api/v1/trades/:id/payment/initiate
func (h *PaymentHandler) InitiatePayment(c echo.Context) error {
	tradeID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid trade id"})
	}

	buyerID := middleware.MustGetUserID(c).String()

	result, err := h.svc.InitiatePayment(c.Request().Context(), tradeID, buyerID)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": err.Error()})
	}

	return c.JSON(http.StatusOK, result)
}

// Webhook POST /api/v1/webhooks/payment
// 토스페이먼츠 웹훅 수신 (인증 불필요 - 웹훅 시크릿 헤더로 검증)
func (h *PaymentHandler) Webhook(c echo.Context) error {
	var payload domain.TossWebhookPayload
	if err := c.Bind(&payload); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid payload"})
	}

	// 결제 완료 이벤트만 처리
	if payload.EventType != "PAYMENT_STATUS_CHANGED" || payload.Data.Status != "DONE" {
		return c.JSON(http.StatusOK, map[string]string{"status": "ignored"})
	}

	err := h.svc.ConfirmPayment(
		c.Request().Context(),
		payload.Data.OrderID,
		payload.Data.PaymentKey,
		strconv.FormatInt(payload.Data.Amount, 10),
	)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}

	return c.JSON(http.StatusOK, map[string]string{"status": "ok"})
}

// MockConfirm POST /api/v1/trades/:id/payment/mock-confirm (개발/테스트용)
// 실제 PG 없이 FUNDED 상태로 전환
func (h *PaymentHandler) MockConfirm(c echo.Context) error {
	tradeID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid trade id"})
	}

	// pg_order_id 조회
	var orderID string
	if err := h.svc.DB().QueryRowContext(c.Request().Context(),
		`SELECT COALESCE(pg_order_id,'') FROM escrow_transactions WHERE trade_id=$1`, tradeID,
	).Scan(&orderID); err != nil || orderID == "" {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "결제 초기화를 먼저 하세요"})
	}

	if err := h.svc.ConfirmPayment(c.Request().Context(), orderID, "mock-payment-key", "0"); err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}

	return c.JSON(http.StatusOK, map[string]string{"status": "funded", "note": "모의 결제 확인 완료"})
}
