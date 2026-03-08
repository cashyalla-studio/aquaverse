package handler

import (
	"net/http"
	"strconv"

	"github.com/cashyalla/aquaverse/internal/api/middleware"
	"github.com/cashyalla/aquaverse/internal/service"
	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
)

// BadgeHandler exposes badge definition, user-badge, and challenge endpoints.
type BadgeHandler struct {
	svc *service.BadgeService
}

func NewBadgeHandler(svc *service.BadgeService) *BadgeHandler {
	return &BadgeHandler{svc: svc}
}

// ListBadges godoc
// GET /api/v1/badges
// Returns every active badge definition.
func (h *BadgeHandler) ListBadges(c echo.Context) error {
	badges, err := h.svc.ListBadgeDefinitions(c.Request().Context())
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}
	return c.JSON(http.StatusOK, map[string]interface{}{"badges": badges})
}

// GetMyBadges godoc
// GET /api/v1/users/me/badges  (requires JWT)
// Returns all badges earned by the authenticated user.
func (h *BadgeHandler) GetMyBadges(c echo.Context) error {
	userID := middleware.MustGetUserID(c)
	badges, err := h.svc.GetUserBadges(c.Request().Context(), userID)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}
	return c.JSON(http.StatusOK, map[string]interface{}{"badges": badges})
}

// GetUserBadges godoc
// GET /api/v1/users/:id/badges
// Returns all badges earned by a specific user (public).
func (h *BadgeHandler) GetUserBadges(c echo.Context) error {
	targetID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid user id"})
	}
	badges, err := h.svc.GetUserBadges(c.Request().Context(), targetID)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}
	return c.JSON(http.StatusOK, map[string]interface{}{"badges": badges})
}

// ListChallenges godoc
// GET /api/v1/challenges
// Returns all active (non-expired) challenges.
func (h *BadgeHandler) ListChallenges(c echo.Context) error {
	challenges, err := h.svc.ListActiveChallenges(c.Request().Context())
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}
	return c.JSON(http.StatusOK, map[string]interface{}{"challenges": challenges})
}

// JoinChallenge godoc
// POST /api/v1/challenges/:id/join  (requires JWT)
// Registers the authenticated user as a participant in the challenge.
func (h *BadgeHandler) JoinChallenge(c echo.Context) error {
	challengeID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid challenge id"})
	}
	userID := middleware.MustGetUserID(c)
	if err := h.svc.JoinChallenge(c.Request().Context(), challengeID, userID); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": err.Error()})
	}
	return c.JSON(http.StatusOK, map[string]string{"status": "joined"})
}

// GetChallengeProgress godoc
// GET /api/v1/challenges/:id/progress  (requires JWT)
// Returns the authenticated user's progress for the given challenge.
func (h *BadgeHandler) GetChallengeProgress(c echo.Context) error {
	challengeID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid challenge id"})
	}
	userID := middleware.MustGetUserID(c)
	progress, err := h.svc.GetProgress(c.Request().Context(), challengeID, userID)
	if err != nil {
		if err.Error() == "not joined" {
			return c.JSON(http.StatusNotFound, map[string]string{"error": "not joined"})
		}
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}
	return c.JSON(http.StatusOK, progress)
}
