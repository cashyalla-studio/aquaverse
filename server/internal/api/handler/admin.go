package handler

import (
	"net/http"
	"strconv"

	"github.com/cashyalla/aquaverse/internal/api/middleware"
	"github.com/cashyalla/aquaverse/internal/service"
	"github.com/labstack/echo/v4"
)

type AdminHandler struct {
	svc *service.AdminService
}

func NewAdminHandler(svc *service.AdminService) *AdminHandler {
	return &AdminHandler{svc: svc}
}

func (h *AdminHandler) GetKPI(c echo.Context) error {
	kpi, err := h.svc.GetKPI(c.Request().Context())
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}
	return c.JSON(http.StatusOK, kpi)
}

func (h *AdminHandler) ListUsers(c echo.Context) error {
	limit, _ := strconv.Atoi(c.QueryParam("limit"))
	if limit <= 0 {
		limit = 20
	}
	offset, _ := strconv.Atoi(c.QueryParam("offset"))
	q := c.QueryParam("q")

	users, err := h.svc.ListUsers(c.Request().Context(), limit, offset, q)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}
	return c.JSON(http.StatusOK, map[string]interface{}{"users": users, "limit": limit, "offset": offset})
}

func (h *AdminHandler) BanUser(c echo.Context) error {
	userID := c.Param("id")
	var req struct {
		Reason string `json:"reason"`
	}
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid request"})
	}
	adminID := middleware.MustGetUserID(c).String()
	if err := h.svc.BanUser(c.Request().Context(), adminID, userID, req.Reason, c.RealIP()); err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}
	return c.JSON(http.StatusOK, map[string]string{"status": "banned"})
}

func (h *AdminHandler) UnbanUser(c echo.Context) error {
	userID := c.Param("id")
	adminID := middleware.MustGetUserID(c).String()
	if err := h.svc.UnbanUser(c.Request().Context(), adminID, userID, c.RealIP()); err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}
	return c.JSON(http.StatusOK, map[string]string{"status": "unbanned"})
}

func (h *AdminHandler) GetAuditLogs(c echo.Context) error {
	limit, _ := strconv.Atoi(c.QueryParam("limit"))
	if limit <= 0 {
		limit = 50
	}
	offset, _ := strconv.Atoi(c.QueryParam("offset"))

	logs, err := h.svc.GetAuditLogs(c.Request().Context(), limit, offset)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}
	return c.JSON(http.StatusOK, map[string]interface{}{"logs": logs})
}

func (h *AdminHandler) GetCitesStats(c echo.Context) error {
	stats, err := h.svc.GetCitesAuditStats(c.Request().Context())
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}
	return c.JSON(http.StatusOK, map[string]interface{}{"stats": stats})
}
