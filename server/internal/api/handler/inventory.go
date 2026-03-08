package handler

import (
	"net/http"
	"strconv"

	"github.com/cashyalla/aquaverse/internal/api/middleware"
	"github.com/cashyalla/aquaverse/internal/domain"
	"github.com/cashyalla/aquaverse/internal/service"
	"github.com/labstack/echo/v4"
)

// InventoryHandler Business Hub 재고 관리 API 핸들러
type InventoryHandler struct {
	svc *service.InventoryService
}

// NewInventoryHandler InventoryHandler를 생성한다.
func NewInventoryHandler(svc *service.InventoryService) *InventoryHandler {
	return &InventoryHandler{svc: svc}
}

// ListInventory GET /api/v1/businesses/:id/inventory
func (h *InventoryHandler) ListInventory(c echo.Context) error {
	businessID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid business id"})
	}

	items, err := h.svc.ListInventory(c.Request().Context(), businessID)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}
	return c.JSON(http.StatusOK, map[string]interface{}{
		"inventory": items,
		"total":     len(items),
	})
}

// CreateInventory POST /api/v1/businesses/:id/inventory  (JWT 필요)
func (h *InventoryHandler) CreateInventory(c echo.Context) error {
	businessID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid business id"})
	}
	userID := middleware.MustGetUserID(c).String()
	if userID == "" {
		return c.JSON(http.StatusUnauthorized, map[string]string{"error": "인증이 필요합니다"})
	}

	var req domain.ShopInventoryRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": err.Error()})
	}

	item, err := h.svc.UpsertInventory(c.Request().Context(), businessID, userID, req)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": err.Error()})
	}
	return c.JSON(http.StatusCreated, item)
}

// UpdateInventory PUT /api/v1/businesses/:id/inventory/:itemId  (JWT 필요)
func (h *InventoryHandler) UpdateInventory(c echo.Context) error {
	businessID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid business id"})
	}
	itemID, err := strconv.ParseInt(c.Param("itemId"), 10, 64)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid item id"})
	}
	userID := middleware.MustGetUserID(c).String()
	if userID == "" {
		return c.JSON(http.StatusUnauthorized, map[string]string{"error": "인증이 필요합니다"})
	}

	var req domain.ShopInventoryRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": err.Error()})
	}

	item, err := h.svc.UpdateInventory(c.Request().Context(), businessID, itemID, userID, req)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": err.Error()})
	}
	return c.JSON(http.StatusOK, item)
}

// DeleteInventory DELETE /api/v1/businesses/:id/inventory/:itemId  (JWT 필요)
func (h *InventoryHandler) DeleteInventory(c echo.Context) error {
	businessID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid business id"})
	}
	itemID, err := strconv.ParseInt(c.Param("itemId"), 10, 64)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid item id"})
	}
	userID := middleware.MustGetUserID(c).String()
	if userID == "" {
		return c.JSON(http.StatusUnauthorized, map[string]string{"error": "인증이 필요합니다"})
	}

	if err := h.svc.DeleteInventory(c.Request().Context(), businessID, itemID, userID); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": err.Error()})
	}
	return c.JSON(http.StatusOK, map[string]string{"message": "재고 항목이 삭제되었습니다"})
}

// GetBusinessStats GET /api/v1/businesses/:id/stats
func (h *InventoryHandler) GetBusinessStats(c echo.Context) error {
	businessID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid business id"})
	}

	stats, err := h.svc.GetBusinessStats(c.Request().Context(), businessID)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}
	return c.JSON(http.StatusOK, stats)
}
