package handler

import (
	"net/http"
	"strconv"

	"github.com/cashyalla/aquaverse/internal/api/middleware"
	"github.com/cashyalla/aquaverse/internal/domain"
	"github.com/cashyalla/aquaverse/internal/service"
	"github.com/labstack/echo/v4"
)

type BusinessHandler struct {
	svc *service.BusinessService
}

func NewBusinessHandler(svc *service.BusinessService) *BusinessHandler {
	return &BusinessHandler{svc: svc}
}

// ListBusinesses GET /api/v1/businesses?city=서울&limit=20&offset=0
func (h *BusinessHandler) ListBusinesses(c echo.Context) error {
	city := c.QueryParam("city")
	limit, _ := strconv.Atoi(c.QueryParam("limit"))
	offset, _ := strconv.Atoi(c.QueryParam("offset"))
	profiles, err := h.svc.ListBusinesses(c.Request().Context(), city, limit, offset)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}
	return c.JSON(http.StatusOK, map[string]interface{}{"businesses": profiles, "total": len(profiles)})
}

// GetBusiness GET /api/v1/businesses/:id
func (h *BusinessHandler) GetBusiness(c echo.Context) error {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid id"})
	}
	profile, err := h.svc.GetProfile(c.Request().Context(), id)
	if err != nil {
		return c.JSON(http.StatusNotFound, map[string]string{"error": "업체를 찾을 수 없습니다"})
	}
	return c.JSON(http.StatusOK, profile)
}

// CreateBusiness POST /api/v1/businesses
func (h *BusinessHandler) CreateBusiness(c echo.Context) error {
	userID := middleware.MustGetUserID(c).String()
	var profile domain.BusinessProfile
	if err := c.Bind(&profile); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": err.Error()})
	}
	if profile.StoreName == "" {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "store_name 필수"})
	}
	result, err := h.svc.CreateProfile(c.Request().Context(), userID, profile)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}
	return c.JSON(http.StatusCreated, result)
}

// UpdateBusiness PUT /api/v1/businesses/:id
func (h *BusinessHandler) UpdateBusiness(c echo.Context) error {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid id"})
	}
	userID := middleware.MustGetUserID(c).String()
	var patch domain.BusinessProfile
	if err := c.Bind(&patch); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": err.Error()})
	}
	result, err := h.svc.UpdateProfile(c.Request().Context(), id, userID, patch)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}
	return c.JSON(http.StatusOK, result)
}

// AddReview POST /api/v1/businesses/:id/reviews
func (h *BusinessHandler) AddReview(c echo.Context) error {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid id"})
	}
	userID := middleware.MustGetUserID(c).String()
	var req struct {
		Rating  int    `json:"rating"`
		Content string `json:"content"`
	}
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": err.Error()})
	}
	if req.Rating < 1 || req.Rating > 5 {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "rating은 1-5"})
	}
	review, err := h.svc.AddReview(c.Request().Context(), id, userID, req.Rating, req.Content)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}
	return c.JSON(http.StatusCreated, review)
}

// NearbyBusinesses GET /api/v1/businesses/nearby?lat=37.5&lng=127.0&radius=5
func (h *BusinessHandler) NearbyBusinesses(c echo.Context) error {
	lat, err := strconv.ParseFloat(c.QueryParam("lat"), 64)
	if err != nil || lat == 0 {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "lat 파라미터 필요"})
	}
	lng, err := strconv.ParseFloat(c.QueryParam("lng"), 64)
	if err != nil || lng == 0 {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "lng 파라미터 필요"})
	}
	radius, _ := strconv.ParseFloat(c.QueryParam("radius"), 64)
	if radius <= 0 || radius > 50 {
		radius = 5.0 // 기본 5km
	}
	limit, _ := strconv.Atoi(c.QueryParam("limit"))

	businesses, err := h.svc.NearbyBusinesses(c.Request().Context(), lat, lng, radius, limit)
	if err != nil {
		// PostGIS 없는 환경 fallback: 일반 목록 반환
		businesses, err = h.svc.ListBusinesses(c.Request().Context(), "", limit, 0)
		if err != nil {
			return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
		}
	}
	return c.JSON(http.StatusOK, map[string]interface{}{
		"businesses": businesses,
		"center":     map[string]float64{"lat": lat, "lng": lng},
		"radius_km":  radius,
	})
}

// GetReviews GET /api/v1/businesses/:id/reviews
func (h *BusinessHandler) GetReviews(c echo.Context) error {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid id"})
	}
	reviews, err := h.svc.GetReviews(c.Request().Context(), id)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}
	return c.JSON(http.StatusOK, map[string]interface{}{"reviews": reviews, "total": len(reviews)})
}
