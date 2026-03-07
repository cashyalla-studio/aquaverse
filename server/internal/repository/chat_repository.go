package repository

import (
	"context"

	"github.com/cashyalla/aquaverse/internal/domain"
	"github.com/jmoiron/sqlx"
)

type ChatRepository struct {
	db *sqlx.DB
}

func NewChatRepository(db *sqlx.DB) *ChatRepository {
	return &ChatRepository{db: db}
}

// DB returns the underlying sqlx.DB instance.
func (r *ChatRepository) DB() *sqlx.DB {
	return r.db
}

func (r *ChatRepository) GetOrCreateRoom(ctx context.Context, tradeID, buyerID, sellerID interface{}) (*domain.ChatRoom, error) {
	room := &domain.ChatRoom{}
	err := r.db.GetContext(ctx, room, `SELECT * FROM chat_rooms WHERE trade_id=$1`, tradeID)
	if err == nil {
		return room, nil
	}
	err = r.db.QueryRowxContext(ctx,
		`INSERT INTO chat_rooms (trade_id, buyer_id, seller_id) VALUES ($1,$2,$3)
		 ON CONFLICT (trade_id) DO UPDATE SET updated_at=NOW()
		 RETURNING *`,
		tradeID, buyerID, sellerID,
	).StructScan(room)
	return room, err
}

func (r *ChatRepository) GetRoom(ctx context.Context, roomID int64) (*domain.ChatRoom, error) {
	room := &domain.ChatRoom{}
	err := r.db.GetContext(ctx, room, `SELECT * FROM chat_rooms WHERE id=$1`, roomID)
	return room, err
}

func (r *ChatRepository) SaveMessage(ctx context.Context, roomID int64, senderID interface{}, content string, msgType domain.MsgType) (*domain.ChatMessage, error) {
	msg := &domain.ChatMessage{}
	err := r.db.QueryRowxContext(ctx,
		`INSERT INTO chat_messages (room_id, sender_id, content, msg_type) VALUES ($1,$2,$3,$4) RETURNING *`,
		roomID, senderID, content, msgType,
	).StructScan(msg)
	return msg, err
}

func (r *ChatRepository) GetHistory(ctx context.Context, roomID int64, limit int) ([]domain.ChatMessage, error) {
	var msgs []domain.ChatMessage
	err := r.db.SelectContext(ctx, &msgs,
		`SELECT * FROM chat_messages WHERE room_id=$1 AND is_deleted=FALSE ORDER BY created_at DESC LIMIT $2`,
		roomID, limit,
	)
	// reverse to chronological order
	for i, j := 0, len(msgs)-1; i < j; i, j = i+1, j-1 {
		msgs[i], msgs[j] = msgs[j], msgs[i]
	}
	return msgs, err
}

func (r *ChatRepository) GetRoomByTradeID(ctx context.Context, tradeID int64) (*domain.ChatRoom, error) {
	room := &domain.ChatRoom{}
	err := r.db.GetContext(ctx, room, `SELECT * FROM chat_rooms WHERE trade_id=$1`, tradeID)
	return room, err
}
