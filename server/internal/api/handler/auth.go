package handler

import (
	"net/http"

	"github.com/cashyalla/aquaverse/internal/api/middleware"
	"github.com/cashyalla/aquaverse/internal/service"
	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
)

type AuthHandler struct {
	authSvc *service.AuthService
}

func NewAuthHandler(authSvc *service.AuthService) *AuthHandler {
	return &AuthHandler{authSvc: authSvc}
}

// POST /api/v1/auth/register
func (h *AuthHandler) Register(c echo.Context) error {
	var req service.RegisterRequest
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid request body")
	}

	user, err := h.authSvc.Register(c.Request().Context(), req)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}

	return c.JSON(http.StatusCreated, map[string]interface{}{
		"id":       user.ID,
		"email":    user.Email,
		"username": user.Username,
		"message":  "registration successful",
	})
}

// POST /api/v1/auth/login
func (h *AuthHandler) Login(c echo.Context) error {
	var req service.LoginRequest
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid request body")
	}
	req.DeviceInfo = c.Request().UserAgent()

	pair, err := h.authSvc.Login(c.Request().Context(), req)
	if err != nil {
		return echo.NewHTTPError(http.StatusUnauthorized, err.Error())
	}

	return c.JSON(http.StatusOK, pair)
}

// POST /api/v1/auth/refresh
func (h *AuthHandler) Refresh(c echo.Context) error {
	var req struct {
		RefreshToken string `json:"refresh_token"`
	}
	if err := c.Bind(&req); err != nil || req.RefreshToken == "" {
		return echo.NewHTTPError(http.StatusBadRequest, "refresh_token required")
	}

	pair, err := h.authSvc.Refresh(c.Request().Context(), req.RefreshToken)
	if err != nil {
		return echo.NewHTTPError(http.StatusUnauthorized, err.Error())
	}

	return c.JSON(http.StatusOK, pair)
}

// POST /api/v1/auth/logout
func (h *AuthHandler) Logout(c echo.Context) error {
	var req struct {
		RefreshToken string `json:"refresh_token"`
	}
	c.Bind(&req)

	if req.RefreshToken != "" {
		_ = h.authSvc.Logout(c.Request().Context(), req.RefreshToken)
	}
	return c.NoContent(http.StatusNoContent)
}

// POST /api/v1/auth/logout-all
func (h *AuthHandler) LogoutAll(c echo.Context) error {
	userIDStr := c.Get(middleware.ContextKeyUserID).(string)
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid user id")
	}
	_ = h.authSvc.LogoutAll(c.Request().Context(), userID)
	return c.NoContent(http.StatusNoContent)
}
