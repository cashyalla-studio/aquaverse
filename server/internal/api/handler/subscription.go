package handler

import (
	"net/http"

	"github.com/cashyalla/aquaverse/internal/api/middleware"
	"github.com/cashyalla/aquaverse/internal/service"
	"github.com/labstack/echo/v4"
)

type SubscriptionHandler struct {
	svc *service.SubscriptionService
}

func NewSubscriptionHandler(svc *service.SubscriptionService) *SubscriptionHandler {
	return &SubscriptionHandler{svc: svc}
}

// GetPlans GET /api/v1/subscriptions/plans
func (h *SubscriptionHandler) GetPlans(c echo.Context) error {
	return c.JSON(http.StatusOK, map[string]interface{}{"plans": h.svc.GetPlans()})
}

// GetMySubscription GET /api/v1/subscriptions/me
func (h *SubscriptionHandler) GetMySubscription(c echo.Context) error {
	userID := middleware.MustGetUserID(c).String()
	sub, err := h.svc.GetSubscription(c.Request().Context(), userID)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}
	return c.JSON(http.StatusOK, sub)
}

// Subscribe POST /api/v1/subscriptions/subscribe
func (h *SubscriptionHandler) Subscribe(c echo.Context) error {
	userID := middleware.MustGetUserID(c).String()
	var req struct {
		BillingKey string `json:"billing_key"`
		Trial      bool   `json:"trial"` // true이면 빌링키 없이 1개월 무료
	}
	c.Bind(&req)

	var sub interface{}
	var err error
	if req.Trial || req.BillingKey == "" {
		sub, err = h.svc.SubscribeFree(c.Request().Context(), userID)
	} else {
		sub, err = h.svc.Subscribe(c.Request().Context(), userID, req.BillingKey)
	}
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}
	return c.JSON(http.StatusOK, sub)
}

// Cancel POST /api/v1/subscriptions/cancel
func (h *SubscriptionHandler) Cancel(c echo.Context) error {
	userID := middleware.MustGetUserID(c).String()
	if err := h.svc.Cancel(c.Request().Context(), userID); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": err.Error()})
	}
	return c.JSON(http.StatusOK, map[string]string{"status": "cancelled"})
}
