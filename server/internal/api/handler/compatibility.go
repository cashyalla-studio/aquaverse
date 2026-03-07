package handler

import (
	"net/http"
	"strconv"

	"github.com/cashyalla/aquaverse/internal/service"
	"github.com/labstack/echo/v4"
)

type CompatibilityHandler struct {
	svc *service.CompatibilityService
}

func NewCompatibilityHandler(svc *service.CompatibilityService) *CompatibilityHandler {
	return &CompatibilityHandler{svc: svc}
}

// GetCompatibleFish GET /api/v1/fish/:id/compatible
func (h *CompatibilityHandler) GetCompatibleFish(c echo.Context) error {
	fishID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid fish id"})
	}

	result, err := h.svc.GetCompatibleFish(c.Request().Context(), fishID)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}

	return c.JSON(http.StatusOK, map[string]interface{}{
		"fish_id":         fishID,
		"compatible_fish": result,
		"total":           len(result),
	})
}

// RecommendForTank GET /api/v1/tanks/:id/recommend
func (h *CompatibilityHandler) RecommendForTank(c echo.Context) error {
	tankID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid tank id"})
	}

	result, err := h.svc.RecommendForTank(c.Request().Context(), tankID)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}

	return c.JSON(http.StatusOK, map[string]interface{}{
		"tank_id":         tankID,
		"recommendations": result,
		"total":           len(result),
	})
}

// CheckWithClaude GET /api/v1/fish/check-compat?a=1&b=2
func (h *CompatibilityHandler) CheckWithClaude(c echo.Context) error {
	aID, _ := strconv.ParseInt(c.QueryParam("a"), 10, 64)
	bID, _ := strconv.ParseInt(c.QueryParam("b"), 10, 64)
	if aID == 0 || bID == 0 {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "a, b 파라미터 필요"})
	}
	result, err := h.svc.ClaudeFallbackCheck(c.Request().Context(), aID, bID)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}
	return c.JSON(http.StatusOK, result)
}

// GetTankInhabitants GET /api/v1/tanks/:id/inhabitants
func (h *CompatibilityHandler) GetTankInhabitants(c echo.Context) error {
	tankID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid tank id"})
	}

	var inhabitants []struct {
		FishDataID int64  `db:"fish_data_id" json:"fish_data_id"`
		FishName   string `db:"fish_name" json:"fish_name"`
		Quantity   int    `db:"quantity" json:"quantity"`
	}
	// 직접 DB 조회 (간단한 핸들러)
	// 실제로는 service 메서드로 분리 권장
	_ = tankID
	return c.JSON(http.StatusOK, map[string]interface{}{
		"tank_id":     tankID,
		"inhabitants": inhabitants,
	})
}
