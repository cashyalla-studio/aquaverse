package service

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"

	"github.com/cashyalla/aquaverse/internal/domain"
	"github.com/cashyalla/aquaverse/internal/repository"
	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
)

const chatChannelPrefix = "chat:room:"

type ChatClient struct {
	UserID uuid.UUID
	Send   chan []byte
}

type ChatHub struct {
	rooms    map[int64]map[*ChatClient]bool
	mu       sync.RWMutex
	rdb      *redis.Client
	chatRepo *repository.ChatRepository
}

func NewChatHub(rdb *redis.Client, chatRepo *repository.ChatRepository) *ChatHub {
	return &ChatHub{
		rooms:    make(map[int64]map[*ChatClient]bool),
		rdb:      rdb,
		chatRepo: chatRepo,
	}
}

func (h *ChatHub) Join(roomID int64, client *ChatClient) {
	h.mu.Lock()
	defer h.mu.Unlock()
	if h.rooms[roomID] == nil {
		h.rooms[roomID] = make(map[*ChatClient]bool)
	}
	h.rooms[roomID][client] = true
}

func (h *ChatHub) Leave(roomID int64, client *ChatClient) {
	h.mu.Lock()
	defer h.mu.Unlock()
	if clients, ok := h.rooms[roomID]; ok {
		delete(clients, client)
		if len(clients) == 0 {
			delete(h.rooms, roomID)
		}
	}
	close(client.Send)
}

func (h *ChatHub) Broadcast(ctx context.Context, roomID int64, data []byte) {
	// Redis Pub/Sub으로 멀티 인스턴스 지원
	h.rdb.Publish(ctx, fmt.Sprintf("%s%d", chatChannelPrefix, roomID), data)
}

func (h *ChatHub) deliverLocal(roomID int64, data []byte) {
	h.mu.RLock()
	defer h.mu.RUnlock()
	for client := range h.rooms[roomID] {
		select {
		case client.Send <- data:
		default:
			// 슬로우 클라이언트 메시지 드롭 (채널 버퍼 초과)
		}
	}
}

// StartSubscriber Redis Pub/Sub 구독 goroutine
func (h *ChatHub) StartSubscriber(ctx context.Context) {
	// 패턴 구독: chat:room:*
	pubsub := h.rdb.PSubscribe(ctx, fmt.Sprintf("%s*", chatChannelPrefix))
	defer pubsub.Close()

	ch := pubsub.Channel()
	for {
		select {
		case msg, ok := <-ch:
			if !ok {
				return
			}
			// 채널명에서 roomID 추출
			var roomID int64
			fmt.Sscanf(msg.Channel, chatChannelPrefix+"%d", &roomID)
			h.deliverLocal(roomID, []byte(msg.Payload))
		case <-ctx.Done():
			return
		}
	}
}

// SendMessage 메시지 저장 + 브로드캐스트
func (h *ChatHub) SendMessage(ctx context.Context, roomID int64, senderID uuid.UUID, content string) (*domain.ChatMessage, error) {
	msg, err := h.chatRepo.SaveMessage(ctx, roomID, senderID, content, domain.MsgTypeText)
	if err != nil {
		return nil, err
	}

	wsMsg := domain.WsMessage{
		Type:      "message",
		RoomID:    roomID,
		SenderID:  senderID.String(),
		Content:   msg.Content,
		MsgType:   msg.MsgType,
		MessageID: msg.ID,
		CreatedAt: &msg.CreatedAt,
	}
	data, _ := json.Marshal(wsMsg)
	h.Broadcast(ctx, roomID, data)
	return msg, nil
}
