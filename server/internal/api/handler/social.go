package handler

import (
	"net/http"
	"strconv"

	"github.com/cashyalla/aquaverse/internal/api/middleware"
	"github.com/cashyalla/aquaverse/internal/service"
	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
)

type SocialHandler struct {
	svc *service.SocialService
}

func NewSocialHandler(svc *service.SocialService) *SocialHandler {
	return &SocialHandler{svc: svc}
}

func (h *SocialHandler) Follow(c echo.Context) error {
	followerID := middleware.MustGetUserID(c)
	followingIDStr := c.Param("id")
	followingID, err := uuid.Parse(followingIDStr)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid user id"})
	}
	if err := h.svc.Follow(c.Request().Context(), followerID, followingID); err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}
	return c.JSON(http.StatusOK, map[string]string{"status": "following"})
}

func (h *SocialHandler) Unfollow(c echo.Context) error {
	followerID := middleware.MustGetUserID(c)
	followingIDStr := c.Param("id")
	followingID, err := uuid.Parse(followingIDStr)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid user id"})
	}
	if err := h.svc.Unfollow(c.Request().Context(), followerID, followingID); err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}
	return c.JSON(http.StatusOK, map[string]string{"status": "unfollowed"})
}

func (h *SocialHandler) GetFeed(c echo.Context) error {
	userID := middleware.MustGetUserID(c)
	limit, _ := strconv.Atoi(c.QueryParam("limit"))
	if limit <= 0 {
		limit = 20
	}
	offset, _ := strconv.Atoi(c.QueryParam("offset"))

	items, err := h.svc.GetFeed(c.Request().Context(), userID, limit, offset)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}
	return c.JSON(http.StatusOK, map[string]interface{}{"feed": items, "limit": limit, "offset": offset})
}

func (h *SocialHandler) GetSuggestions(c echo.Context) error {
	userID := middleware.MustGetUserID(c)
	suggestions, err := h.svc.GetSuggestions(c.Request().Context(), userID)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}
	return c.JSON(http.StatusOK, map[string]interface{}{"suggestions": suggestions})
}

func (h *SocialHandler) GetFollowing(c echo.Context) error {
	userID := middleware.MustGetUserID(c)
	limit, _ := strconv.Atoi(c.QueryParam("limit"))
	if limit <= 0 {
		limit = 20
	}
	offset, _ := strconv.Atoi(c.QueryParam("offset"))

	follows, err := h.svc.GetFollowing(c.Request().Context(), userID, limit, offset)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}
	return c.JSON(http.StatusOK, map[string]interface{}{"following": follows})
}
