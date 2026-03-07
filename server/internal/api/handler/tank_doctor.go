package handler

import (
	"net/http"
	"strconv"

	"github.com/cashyalla/aquaverse/internal/domain"
	"github.com/cashyalla/aquaverse/internal/service"
	"github.com/labstack/echo/v4"
)

type TankDoctorHandler struct {
	svc *service.TankDoctorService
}

func NewTankDoctorHandler(svc *service.TankDoctorService) *TankDoctorHandler {
	return &TankDoctorHandler{svc: svc}
}

// RecordWaterParams POST /api/v1/tanks/:id/water-params
func (h *TankDoctorHandler) RecordWaterParams(c echo.Context) error {
	tankID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid tank id"})
	}

	var params domain.WaterParams
	if err := c.Bind(&params); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": err.Error()})
	}
	params.TankID = tankID

	result, err := h.svc.RecordWaterParams(c.Request().Context(), params)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}
	return c.JSON(http.StatusCreated, result)
}

// GetWaterHistory GET /api/v1/tanks/:id/water-params
func (h *TankDoctorHandler) GetWaterHistory(c echo.Context) error {
	tankID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid tank id"})
	}

	history, err := h.svc.GetWaterHistory(c.Request().Context(), tankID, 20)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}
	return c.JSON(http.StatusOK, map[string]interface{}{
		"tank_id": tankID,
		"history": history,
	})
}

// GetDiagnosis GET /api/v1/tanks/:id/diagnosis
func (h *TankDoctorHandler) GetDiagnosis(c echo.Context) error {
	tankID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid tank id"})
	}

	diag, err := h.svc.DiagnoseTank(c.Request().Context(), tankID)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}
	return c.JSON(http.StatusOK, diag)
}
