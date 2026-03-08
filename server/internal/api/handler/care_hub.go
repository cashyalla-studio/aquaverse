package handler

import (
	"net/http"
	"strconv"

	"github.com/cashyalla/aquaverse/internal/api/middleware"
	"github.com/cashyalla/aquaverse/internal/domain"
	"github.com/cashyalla/aquaverse/internal/service"
	"github.com/labstack/echo/v4"
)

// CareHubHandler 케어 허브 핸들러
type CareHubHandler struct {
	svc *service.CareHubService
}

func NewCareHubHandler(svc *service.CareHubService) *CareHubHandler {
	return &CareHubHandler{svc: svc}
}

// CreateSchedule POST /api/v1/tanks/:id/schedules
func (h *CareHubHandler) CreateSchedule(c echo.Context) error {
	tankID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid tank id"})
	}

	userID := middleware.MustGetUserID(c)
	if userID.String() == "00000000-0000-0000-0000-000000000000" {
		return c.JSON(http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
	}

	var req domain.CreateScheduleRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": err.Error()})
	}
	if req.Title == "" || req.ScheduleType == "" || req.Frequency == "" || req.NextDueAt == "" {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "title, schedule_type, frequency, next_due_at 필수"})
	}

	schedule, err := h.svc.CreateSchedule(c.Request().Context(), tankID, userID, req)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}
	return c.JSON(http.StatusCreated, schedule)
}

// ListSchedules GET /api/v1/tanks/:id/schedules
func (h *CareHubHandler) ListSchedules(c echo.Context) error {
	tankID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid tank id"})
	}

	schedules, err := h.svc.ListSchedules(c.Request().Context(), tankID)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}
	return c.JSON(http.StatusOK, map[string]interface{}{
		"tank_id":   tankID,
		"schedules": schedules,
	})
}

// UpdateSchedule PUT /api/v1/schedules/:id
func (h *CareHubHandler) UpdateSchedule(c echo.Context) error {
	scheduleID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid schedule id"})
	}

	userID := middleware.MustGetUserID(c)
	if userID.String() == "00000000-0000-0000-0000-000000000000" {
		return c.JSON(http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
	}

	var req domain.UpdateScheduleRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": err.Error()})
	}

	schedule, err := h.svc.UpdateSchedule(c.Request().Context(), scheduleID, userID, req)
	if err != nil {
		if err.Error() == "수정 권한이 없습니다" {
			return c.JSON(http.StatusForbidden, map[string]string{"error": err.Error()})
		}
		if err.Error() == "일정을 찾을 수 없습니다" {
			return c.JSON(http.StatusNotFound, map[string]string{"error": err.Error()})
		}
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}
	return c.JSON(http.StatusOK, schedule)
}

// DeleteSchedule DELETE /api/v1/schedules/:id
func (h *CareHubHandler) DeleteSchedule(c echo.Context) error {
	scheduleID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid schedule id"})
	}

	userID := middleware.MustGetUserID(c)
	if userID.String() == "00000000-0000-0000-0000-000000000000" {
		return c.JSON(http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
	}

	if err := h.svc.DeleteSchedule(c.Request().Context(), scheduleID, userID); err != nil {
		return c.JSON(http.StatusNotFound, map[string]string{"error": err.Error()})
	}
	return c.JSON(http.StatusOK, map[string]string{"status": "deleted"})
}

// CompleteSchedule POST /api/v1/schedules/:id/complete
func (h *CareHubHandler) CompleteSchedule(c echo.Context) error {
	scheduleID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid schedule id"})
	}

	userID := middleware.MustGetUserID(c)
	if userID.String() == "00000000-0000-0000-0000-000000000000" {
		return c.JSON(http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
	}

	var req domain.CompleteScheduleRequest
	c.Bind(&req)

	log, err := h.svc.CompleteSchedule(c.Request().Context(), scheduleID, userID, req)
	if err != nil {
		if err.Error() == "완료 처리 권한이 없습니다" {
			return c.JSON(http.StatusForbidden, map[string]string{"error": err.Error()})
		}
		if err.Error() == "일정을 찾을 수 없습니다" {
			return c.JSON(http.StatusNotFound, map[string]string{"error": err.Error()})
		}
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}
	return c.JSON(http.StatusCreated, log)
}

// GetTodayTasks GET /api/v1/users/me/care-today
func (h *CareHubHandler) GetTodayTasks(c echo.Context) error {
	userID := middleware.MustGetUserID(c)
	if userID.String() == "00000000-0000-0000-0000-000000000000" {
		return c.JSON(http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
	}

	tasks, err := h.svc.GetTodayTasks(c.Request().Context(), userID)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}
	return c.JSON(http.StatusOK, map[string]interface{}{
		"date":  "",
		"tasks": tasks,
		"count": len(tasks),
	})
}

// GetStreak GET /api/v1/users/me/streak
func (h *CareHubHandler) GetStreak(c echo.Context) error {
	userID := middleware.MustGetUserID(c)
	if userID.String() == "00000000-0000-0000-0000-000000000000" {
		return c.JSON(http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
	}

	streak, err := h.svc.GetStreak(c.Request().Context(), userID)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}
	return c.JSON(http.StatusOK, streak)
}
