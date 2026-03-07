package handler

import (
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"github.com/cashyalla/aquaverse/internal/api/middleware"
	"github.com/cashyalla/aquaverse/internal/domain"
	"github.com/cashyalla/aquaverse/internal/repository"
	"github.com/cashyalla/aquaverse/internal/service"
	"github.com/gorilla/websocket"
	"github.com/labstack/echo/v4"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true // TODO: 프로덕션에서 origin 검증
	},
}

type ChatHandler struct {
	hub      *service.ChatHub
	chatRepo *repository.ChatRepository
}

func NewChatHandler(hub *service.ChatHub, chatRepo *repository.ChatRepository) *ChatHandler {
	return &ChatHandler{hub: hub, chatRepo: chatRepo}
}

// Connect GET /api/v1/trades/:id/chat (WebSocket upgrade)
func (h *ChatHandler) Connect(c echo.Context) error {
	tradeID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid trade id")
	}

	userID := middleware.MustGetUserID(c)

	// 채팅방 조회
	room, err := h.chatRepo.GetRoomByTradeID(c.Request().Context(), tradeID)
	if err != nil {
		// 방이 없으면 에러 (InitiateTrade 시 생성)
		return echo.NewHTTPError(http.StatusNotFound, "chat room not found")
	}

	// 참여 권한: buyer 또는 seller만
	if room.BuyerID != userID && room.SellerID != userID {
		return echo.NewHTTPError(http.StatusForbidden, "not a participant")
	}

	ws, err := upgrader.Upgrade(c.Response(), c.Request(), nil)
	if err != nil {
		return err
	}
	defer ws.Close()

	client := &service.ChatClient{
		UserID: userID,
		Send:   make(chan []byte, 256),
	}
	h.hub.Join(room.ID, client)
	defer h.hub.Leave(room.ID, client)

	// 히스토리 전송
	history, _ := h.chatRepo.GetHistory(c.Request().Context(), room.ID, 50)
	histMsg := domain.WsMessage{Type: "history", RoomID: room.ID, Messages: history}
	if data, err := json.Marshal(histMsg); err == nil {
		ws.WriteMessage(websocket.TextMessage, data)
	}

	// 쓰기 goroutine
	done := make(chan struct{})
	go func() {
		defer close(done)
		ticker := time.NewTicker(30 * time.Second)
		defer ticker.Stop()
		for {
			select {
			case data, ok := <-client.Send:
				ws.SetWriteDeadline(time.Now().Add(10 * time.Second))
				if !ok {
					ws.WriteMessage(websocket.CloseMessage, []byte{})
					return
				}
				if err := ws.WriteMessage(websocket.TextMessage, data); err != nil {
					return
				}
			case <-ticker.C:
				ws.SetWriteDeadline(time.Now().Add(10 * time.Second))
				if err := ws.WriteMessage(websocket.PingMessage, nil); err != nil {
					return
				}
			}
		}
	}()

	// 읽기 루프 (메인 goroutine)
	ws.SetReadLimit(2000)
	ws.SetReadDeadline(time.Now().Add(60 * time.Second))
	ws.SetPongHandler(func(string) error {
		ws.SetReadDeadline(time.Now().Add(60 * time.Second))
		return nil
	})

	type incomingMsg struct {
		Content string `json:"content"`
	}

	for {
		_, raw, err := ws.ReadMessage()
		if err != nil {
			break
		}
		ws.SetReadDeadline(time.Now().Add(60 * time.Second))

		var incoming incomingMsg
		if err := json.Unmarshal(raw, &incoming); err != nil || len(incoming.Content) == 0 {
			continue
		}

		h.hub.SendMessage(c.Request().Context(), room.ID, userID, incoming.Content)
	}
	<-done
	return nil
}

// GetHistory GET /api/v1/trades/:id/chat/history
func (h *ChatHandler) GetHistory(c echo.Context) error {
	tradeID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid trade id")
	}

	room, err := h.chatRepo.GetRoomByTradeID(c.Request().Context(), tradeID)
	if err != nil {
		return echo.NewHTTPError(http.StatusNotFound, "chat room not found")
	}

	userID := middleware.MustGetUserID(c)
	if room.BuyerID != userID && room.SellerID != userID {
		return echo.NewHTTPError(http.StatusForbidden, "not a participant")
	}

	msgs, err := h.chatRepo.GetHistory(c.Request().Context(), room.ID, 100)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to load history")
	}
	return c.JSON(http.StatusOK, msgs)
}
