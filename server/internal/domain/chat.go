package domain

import (
	"time"

	"github.com/google/uuid"
)

type MsgType string

const (
	MsgTypeText   MsgType = "TEXT"
	MsgTypeImage  MsgType = "IMAGE"
	MsgTypeSystem MsgType = "SYSTEM"
)

type ChatRoom struct {
	ID        int64     `db:"id"`
	TradeID   int64     `db:"trade_id"`
	BuyerID   uuid.UUID `db:"buyer_id"`
	SellerID  uuid.UUID `db:"seller_id"`
	IsActive  bool      `db:"is_active"`
	CreatedAt time.Time `db:"created_at"`
	UpdatedAt time.Time `db:"updated_at"`
}

type ChatMessage struct {
	ID        int64     `db:"id"`
	RoomID    int64     `db:"room_id"`
	SenderID  uuid.UUID `db:"sender_id"`
	Content   string    `db:"content"`
	MsgType   MsgType   `db:"msg_type"`
	IsDeleted bool      `db:"is_deleted"`
	CreatedAt time.Time `db:"created_at"`
}

// WsMessage WebSocket 전송용 DTO
type WsMessage struct {
	Type      string        `json:"type"`                 // "message" | "history" | "error" | "system"
	RoomID    int64         `json:"room_id"`
	SenderID  string        `json:"sender_id,omitempty"`
	Content   string        `json:"content,omitempty"`
	MsgType   MsgType       `json:"msg_type,omitempty"`
	MessageID int64         `json:"message_id,omitempty"`
	CreatedAt *time.Time    `json:"created_at,omitempty"`
	Messages  []ChatMessage `json:"messages,omitempty"`
	Error     string        `json:"error,omitempty"`
}
