package handler

import (
	"net/http"

	"github.com/cashyalla/aquaverse/internal/api/middleware"
	"github.com/cashyalla/aquaverse/internal/notification"
	"github.com/labstack/echo/v4"
)

type NotificationHandler struct {
	fcm *notification.FCMService
}

func NewNotificationHandler(fcm *notification.FCMService) *NotificationHandler {
	return &NotificationHandler{fcm: fcm}
}

// RegisterToken POST /api/v1/notifications/fcm/register
func (h *NotificationHandler) RegisterToken(c echo.Context) error {
	userID := middleware.MustGetUserID(c).String()
	var req struct {
		Token    string `json:"token"`
		Platform string `json:"platform"`
	}
	if err := c.Bind(&req); err != nil || req.Token == "" {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "token 필수"})
	}
	if req.Platform == "" {
		req.Platform = "android"
	}
	if err := h.fcm.RegisterToken(c.Request().Context(), userID, req.Token, req.Platform); err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}
	return c.JSON(http.StatusOK, map[string]string{"status": "registered"})
}

// UnregisterToken DELETE /api/v1/notifications/fcm/unregister
func (h *NotificationHandler) UnregisterToken(c echo.Context) error {
	userID := middleware.MustGetUserID(c).String()
	var req struct {
		Token string `json:"token"`
	}
	if err := c.Bind(&req); err != nil || req.Token == "" {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "token 필수"})
	}
	h.fcm.UnregisterToken(c.Request().Context(), userID, req.Token)
	return c.JSON(http.StatusOK, map[string]string{"status": "unregistered"})
}
