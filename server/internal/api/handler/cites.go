package handler

import (
	"net/http"

	"github.com/cashyalla/aquaverse/internal/repository"
	"github.com/labstack/echo/v4"
)

type CitesHandler struct {
	citesRepo *repository.CitesRepository
}

func NewCitesHandler(citesRepo *repository.CitesRepository) *CitesHandler {
	return &CitesHandler{citesRepo: citesRepo}
}

// GET /api/v1/cites/check?scientific_name=xxx
func (h *CitesHandler) Check(c echo.Context) error {
	name := c.QueryParam("scientific_name")
	if name == "" {
		return echo.NewHTTPError(http.StatusBadRequest, "scientific_name required")
	}
	result, err := h.citesRepo.CheckScientificName(c.Request().Context(), name)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "check failed")
	}
	return c.JSON(http.StatusOK, result)
}
