package handler

import (
	"net/http"
	"strconv"

	"github.com/cashyalla/aquaverse/internal/api/middleware"
	"github.com/cashyalla/aquaverse/internal/domain"
	"github.com/cashyalla/aquaverse/internal/service"
	"github.com/labstack/echo/v4"
)

type CommunityHandler struct {
	commSvc *service.CommunityService
}

func NewCommunityHandler(commSvc *service.CommunityService) *CommunityHandler {
	return &CommunityHandler{commSvc: commSvc}
}

// GET /api/v1/boards
// 로케일별 게시판 목록 (X-Locale 헤더 또는 ?locale= 파라미터)
func (h *CommunityHandler) ListBoards(c echo.Context) error {
	locale := h.resolveLocale(c)
	boards, err := h.commSvc.ListBoards(c.Request().Context(), locale)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}
	return c.JSON(http.StatusOK, boards)
}

// GET /api/v1/boards/:boardID/posts
func (h *CommunityHandler) ListPosts(c echo.Context) error {
	boardID, err := strconv.ParseInt(c.Param("boardID"), 10, 64)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid board id")
	}

	locale := h.resolveLocale(c)

	page, _ := strconv.Atoi(c.QueryParam("page"))
	if page < 1 {
		page = 1
	}
	limit, _ := strconv.Atoi(c.QueryParam("limit"))
	if limit < 1 || limit > 50 {
		limit = 20
	}

	// 게시판이 요청 로케일과 일치하는지 검증 (엄격 분리)
	result, err := h.commSvc.ListPosts(c.Request().Context(), boardID, locale, page, limit)
	if err != nil {
		if err == service.ErrLocaleMismatch {
			return echo.NewHTTPError(http.StatusForbidden, "board locale mismatch: this board is not available in your locale")
		}
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}
	return c.JSON(http.StatusOK, result)
}

// GET /api/v1/boards/:boardID/posts/:postID
func (h *CommunityHandler) GetPost(c echo.Context) error {
	boardID, err := strconv.ParseInt(c.Param("boardID"), 10, 64)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid board id")
	}
	postID, err := strconv.ParseInt(c.Param("postID"), 10, 64)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid post id")
	}
	locale := h.resolveLocale(c)

	post, err := h.commSvc.GetPost(c.Request().Context(), boardID, postID, locale)
	if err != nil {
		if err == service.ErrLocaleMismatch {
			return echo.NewHTTPError(http.StatusForbidden, "board locale mismatch")
		}
		return echo.NewHTTPError(http.StatusNotFound, "post not found")
	}
	return c.JSON(http.StatusOK, post)
}

// POST /api/v1/boards/:boardID/posts
func (h *CommunityHandler) CreatePost(c echo.Context) error {
	boardID, err := strconv.ParseInt(c.Param("boardID"), 10, 64)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid board id")
	}

	userID, ok := c.Get(middleware.ContextKeyUserID).(string)
	if !ok || userID == "" {
		return echo.NewHTTPError(http.StatusUnauthorized)
	}

	locale := h.resolveLocale(c)

	var req service.CreatePostRequest
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	req.BoardID = boardID
	req.AuthorID = userID
	req.Locale = locale

	post, err := h.commSvc.CreatePost(c.Request().Context(), req)
	if err != nil {
		if err == service.ErrLocaleMismatch {
			return echo.NewHTTPError(http.StatusForbidden, "board locale mismatch: post language must match board locale")
		}
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	return c.JSON(http.StatusCreated, post)
}

// POST /api/v1/boards/:boardID/posts/:postID/like
func (h *CommunityHandler) LikePost(c echo.Context) error {
	postID, err := strconv.ParseInt(c.Param("postID"), 10, 64)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid post id")
	}
	userID := c.Get(middleware.ContextKeyUserID).(string)

	liked, err := h.commSvc.ToggleLike(c.Request().Context(), postID, userID)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}
	return c.JSON(http.StatusOK, map[string]bool{"liked": liked})
}

// POST /api/v1/boards/:boardID/posts/:postID/comments
func (h *CommunityHandler) CreateComment(c echo.Context) error {
	postID, err := strconv.ParseInt(c.Param("postID"), 10, 64)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid post id")
	}
	userID := c.Get(middleware.ContextKeyUserID).(string)

	var req service.CreateCommentRequest
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	req.PostID = postID
	req.AuthorID = userID

	comment, err := h.commSvc.CreateComment(c.Request().Context(), req)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	return c.JSON(http.StatusCreated, comment)
}

func (h *CommunityHandler) resolveLocale(c echo.Context) domain.Locale {
	// 컨텍스트에서 먼저 (미들웨어 설정)
	if loc, ok := c.Get(middleware.ContextKeyLocale).(domain.Locale); ok && loc.IsValid() {
		return loc
	}
	// 쿼리 파라미터
	loc := domain.Locale(c.QueryParam("locale"))
	if loc.IsValid() {
		return loc
	}
	return domain.LocaleENUS
}
