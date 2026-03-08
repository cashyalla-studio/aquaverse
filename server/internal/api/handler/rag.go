package handler

import (
	"net/http"
	"strconv"

	"github.com/cashyalla/aquaverse/internal/api/middleware"
	"github.com/cashyalla/aquaverse/internal/domain"
	"github.com/cashyalla/aquaverse/internal/service"
	"github.com/labstack/echo/v4"
)

// RAGHandler pgvector RAG 챗봇 핸들러
type RAGHandler struct {
	svc *service.RAGService
}

func NewRAGHandler(svc *service.RAGService) *RAGHandler {
	return &RAGHandler{svc: svc}
}

// Ask POST /api/v1/chat/ask
// 로그인하지 않아도 사용 가능하며, 로그인 시 user_id를 세션에 기록한다.
func (h *RAGHandler) Ask(c echo.Context) error {
	var req domain.AskRequest
	if err := c.Bind(&req); err != nil || req.Question == "" {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "question is required"})
	}

	// 선택적 인증
	var userID string
	if uid, err := middleware.GetUserID(c); err == nil {
		userID = uid.String()
	}

	resp, err := h.svc.Answer(c.Request().Context(), req.Question, req.SessionID, userID)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}
	return c.JSON(http.StatusOK, resp)
}

// IndexFish POST /api/v1/admin/fish/:id/index (관리자만)
// 지정한 어종의 임베딩을 생성하여 fish_embeddings에 저장한다.
func (h *RAGHandler) IndexFish(c echo.Context) error {
	fishIDStr := c.Param("id")
	fishID, err := strconv.ParseInt(fishIDStr, 10, 64)
	if err != nil || fishID <= 0 {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid fish id"})
	}

	if err := h.svc.IndexFish(c.Request().Context(), fishID); err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}
	return c.JSON(http.StatusOK, map[string]interface{}{
		"status":  "indexed",
		"fish_id": fishID,
	})
}

// GetSessions GET /api/v1/chat/sessions
// 로그인한 사용자의 채팅 세션 목록을 반환한다.
func (h *RAGHandler) GetSessions(c echo.Context) error {
	userID, err := middleware.GetUserID(c)
	if err != nil {
		return c.JSON(http.StatusUnauthorized, map[string]string{"error": "authentication required"})
	}

	sessions, err := h.svc.GetSessionsByUser(c.Request().Context(), userID.String())
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}
	return c.JSON(http.StatusOK, map[string]interface{}{"sessions": sessions})
}

// GetMessages GET /api/v1/chat/sessions/:id/messages
// 특정 세션의 메시지 목록을 반환한다.
func (h *RAGHandler) GetMessages(c echo.Context) error {
	sessionID := c.Param("id")
	if sessionID == "" {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "session id is required"})
	}

	messages, err := h.svc.GetMessages(c.Request().Context(), sessionID)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}
	return c.JSON(http.StatusOK, map[string]interface{}{"messages": messages})
}
