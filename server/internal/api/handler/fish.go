package handler

import (
	"net/http"
	"strconv"

	"github.com/cashyalla/aquaverse/internal/domain"
	"github.com/cashyalla/aquaverse/internal/service"
	"github.com/labstack/echo/v4"
)

type FishHandler struct {
	fishSvc *service.FishService
}

func NewFishHandler(fishSvc *service.FishService) *FishHandler {
	return &FishHandler{fishSvc: fishSvc}
}

// GET /api/v1/fish
// Query: page, limit, family, care_level, locale
func (h *FishHandler) List(c echo.Context) error {
	locale := domain.Locale(c.QueryParam("locale"))
	if locale == "" || !locale.IsValid() {
		locale = domain.LocaleENUS
	}

	page, _ := strconv.Atoi(c.QueryParam("page"))
	if page < 1 {
		page = 1
	}
	limit, _ := strconv.Atoi(c.QueryParam("limit"))
	if limit < 1 || limit > 100 {
		limit = 20
	}

	filter := service.FishFilter{
		Family:    c.QueryParam("family"),
		CareLevel: c.QueryParam("care_level"),
		Search:    c.QueryParam("q"),
		Locale:    locale,
		Page:      page,
		Limit:     limit,
	}

	result, err := h.fishSvc.List(c.Request().Context(), filter)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}
	return c.JSON(http.StatusOK, result)
}

// GET /api/v1/fish/:id
func (h *FishHandler) Get(c echo.Context) error {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid fish id")
	}

	locale := domain.Locale(c.QueryParam("locale"))
	if locale == "" || !locale.IsValid() {
		locale = domain.LocaleENUS
	}

	fish, err := h.fishSvc.GetByID(c.Request().Context(), id, locale)
	if err != nil {
		return echo.NewHTTPError(http.StatusNotFound, "fish not found")
	}
	return c.JSON(http.StatusOK, fish)
}

// GET /api/v1/fish/search
func (h *FishHandler) Search(c echo.Context) error {
	q := c.QueryParam("q")
	if len(q) < 2 {
		return echo.NewHTTPError(http.StatusBadRequest, "query too short")
	}

	locale := domain.Locale(c.QueryParam("locale"))
	if locale == "" || !locale.IsValid() {
		locale = domain.LocaleENUS
	}

	results, err := h.fishSvc.Search(c.Request().Context(), q, locale)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}
	return c.JSON(http.StatusOK, results)
}

// GET /api/v1/fish/families
func (h *FishHandler) ListFamilies(c echo.Context) error {
	families, err := h.fishSvc.ListFamilies(c.Request().Context())
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}
	return c.JSON(http.StatusOK, families)
}
