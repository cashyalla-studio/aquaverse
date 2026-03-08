package handler

import (
	"net/http"
	"strconv"

	"github.com/cashyalla/aquaverse/internal/api/middleware"
	"github.com/cashyalla/aquaverse/internal/domain"
	"github.com/cashyalla/aquaverse/internal/service"
	"github.com/labstack/echo/v4"
)

// ExpertHandler Expert Connect API 핸들러
type ExpertHandler struct {
	svc *service.ExpertService
}

// NewExpertHandler ExpertHandler를 생성한다.
func NewExpertHandler(svc *service.ExpertService) *ExpertHandler {
	return &ExpertHandler{svc: svc}
}

// ListExperts GET /api/v1/experts?type=vet&page=1&limit=20
func (h *ExpertHandler) ListExperts(c echo.Context) error {
	expertType := c.QueryParam("type")
	page, _ := strconv.Atoi(c.QueryParam("page"))
	limit, _ := strconv.Atoi(c.QueryParam("limit"))

	experts, total, err := h.svc.ListExperts(c.Request().Context(), expertType, page, limit)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}
	return c.JSON(http.StatusOK, map[string]interface{}{
		"experts": experts,
		"total":   total,
	})
}

// GetExpert GET /api/v1/experts/:id
func (h *ExpertHandler) GetExpert(c echo.Context) error {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid id"})
	}
	expert, err := h.svc.GetExpert(c.Request().Context(), id)
	if err != nil {
		return c.JSON(http.StatusNotFound, map[string]string{"error": err.Error()})
	}
	return c.JSON(http.StatusOK, expert)
}

// UpsertProfile PUT /api/v1/experts/profile  (JWT 필요)
func (h *ExpertHandler) UpsertProfile(c echo.Context) error {
	userID := middleware.MustGetUserID(c).String()
	if userID == "" {
		return c.JSON(http.StatusUnauthorized, map[string]string{"error": "인증이 필요합니다"})
	}

	var req domain.ExpertProfileRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": err.Error()})
	}
	if err := h.svc.UpsertProfile(c.Request().Context(), userID, req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": err.Error()})
	}
	return c.JSON(http.StatusOK, map[string]string{"message": "프로필이 저장되었습니다"})
}

// CreateConsultation POST /api/v1/experts/:id/consult  (JWT 필요)
func (h *ExpertHandler) CreateConsultation(c echo.Context) error {
	expertID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid expert id"})
	}
	userID := middleware.MustGetUserID(c).String()
	if userID == "" {
		return c.JSON(http.StatusUnauthorized, map[string]string{"error": "인증이 필요합니다"})
	}

	var req domain.ConsultationRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": err.Error()})
	}

	consultation, err := h.svc.CreateConsultation(c.Request().Context(), userID, expertID, req)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": err.Error()})
	}
	return c.JSON(http.StatusCreated, consultation)
}

// GetMyConsultations GET /api/v1/consultations/me  (JWT 필요)
func (h *ExpertHandler) GetMyConsultations(c echo.Context) error {
	userID := middleware.MustGetUserID(c).String()
	if userID == "" {
		return c.JSON(http.StatusUnauthorized, map[string]string{"error": "인증이 필요합니다"})
	}

	consultations, err := h.svc.GetMyConsultations(c.Request().Context(), userID)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}
	return c.JSON(http.StatusOK, map[string]interface{}{
		"consultations": consultations,
		"total":         len(consultations),
	})
}

// UpdateConsultationStatus PUT /api/v1/consultations/:id/status  (JWT 필요)
func (h *ExpertHandler) UpdateConsultationStatus(c echo.Context) error {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid id"})
	}
	userID := middleware.MustGetUserID(c).String()
	if userID == "" {
		return c.JSON(http.StatusUnauthorized, map[string]string{"error": "인증이 필요합니다"})
	}

	var req struct {
		Status string `json:"status"`
	}
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": err.Error()})
	}
	if req.Status == "" {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "status 필수"})
	}

	if err := h.svc.UpdateConsultationStatus(c.Request().Context(), id, userID, req.Status); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": err.Error()})
	}
	return c.JSON(http.StatusOK, map[string]string{"message": "상태가 변경되었습니다"})
}

// CreateReview POST /api/v1/consultations/:id/review  (JWT 필요)
func (h *ExpertHandler) CreateReview(c echo.Context) error {
	consultationID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid id"})
	}
	reviewerID := middleware.MustGetUserID(c).String()
	if reviewerID == "" {
		return c.JSON(http.StatusUnauthorized, map[string]string{"error": "인증이 필요합니다"})
	}

	var req struct {
		Rating  int    `json:"rating"`
		Comment string `json:"comment"`
	}
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": err.Error()})
	}

	if err := h.svc.CreateReview(c.Request().Context(), consultationID, reviewerID, req.Rating, req.Comment); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": err.Error()})
	}
	return c.JSON(http.StatusCreated, map[string]string{"message": "리뷰가 등록되었습니다"})
}
