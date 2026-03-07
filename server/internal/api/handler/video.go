package handler

import (
	"net/http"
	"strconv"

	"github.com/cashyalla/aquaverse/internal/api/middleware"
	"github.com/cashyalla/aquaverse/internal/service"
	"github.com/labstack/echo/v4"
)

type VideoHandler struct {
	svc *service.VideoService
}

func NewVideoHandler(svc *service.VideoService) *VideoHandler {
	return &VideoHandler{svc: svc}
}

// GetFeed GET /api/v1/videos?limit=10&offset=0
func (h *VideoHandler) GetFeed(c echo.Context) error {
	limit, _ := strconv.Atoi(c.QueryParam("limit"))
	offset, _ := strconv.Atoi(c.QueryParam("offset"))

	// 인증 선택적
	userID := ""
	if uid, err := middleware.GetUserID(c); err == nil {
		userID = uid.String()
	}

	posts, err := h.svc.GetFeed(c.Request().Context(), userID, limit, offset)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}
	return c.JSON(http.StatusOK, map[string]interface{}{"videos": posts, "total": len(posts)})
}

// CreatePost POST /api/v1/videos
func (h *VideoHandler) CreatePost(c echo.Context) error {
	userID := middleware.MustGetUserID(c).String()
	var req struct {
		Title        string `json:"title"`
		Description  string `json:"description"`
		VideoKey     string `json:"video_key"`
		ThumbnailKey string `json:"thumbnail_key"`
		DurationSec  int    `json:"duration_sec"`
	}
	if err := c.Bind(&req); err != nil || req.Title == "" || req.VideoKey == "" {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "title과 video_key 필수"})
	}
	post, err := h.svc.CreatePost(c.Request().Context(), userID, req.Title, req.Description, req.VideoKey, req.ThumbnailKey, req.DurationSec)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}
	return c.JSON(http.StatusCreated, post)
}

// LikePost POST /api/v1/videos/:id/like
func (h *VideoHandler) LikePost(c echo.Context) error {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid id"})
	}
	userID := middleware.MustGetUserID(c).String()
	liked, err := h.svc.LikePost(c.Request().Context(), id, userID)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}
	return c.JSON(http.StatusOK, map[string]bool{"liked": liked})
}

// DeletePost DELETE /api/v1/videos/:id
func (h *VideoHandler) DeletePost(c echo.Context) error {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid id"})
	}
	userID := middleware.MustGetUserID(c).String()
	if err := h.svc.DeletePost(c.Request().Context(), id, userID); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": err.Error()})
	}
	return c.JSON(http.StatusOK, map[string]string{"status": "deleted"})
}

// IncrementView POST /api/v1/videos/:id/view
func (h *VideoHandler) IncrementView(c echo.Context) error {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid id"})
	}
	go h.svc.IncrementView(c.Request().Context(), id)
	return c.JSON(http.StatusOK, map[string]string{"status": "ok"})
}
