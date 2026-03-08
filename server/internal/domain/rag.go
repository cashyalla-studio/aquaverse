package domain

import "time"

// FishEmbedding pgvector에 저장된 어종 임베딩
type FishEmbedding struct {
	FishID    int64     `db:"fish_id"`
	Content   string    `db:"content"`
	UpdatedAt time.Time `db:"updated_at"`
}

// RAGSession RAG 챗봇 전용 세션 (WebSocket chat_sessions와 분리)
type RAGSession struct {
	ID        string    `db:"id"`
	UserID    *string   `db:"user_id"`
	CreatedAt time.Time `db:"created_at"`
	LastMsgAt time.Time `db:"last_msg_at"`
}

// RAGMessage RAG 채팅 메시지 (role: 'user' | 'assistant')
type RAGMessage struct {
	ID        int64     `db:"id"`
	SessionID string    `db:"session_id"`
	Role      string    `db:"role"`
	Content   string    `db:"content"`
	CreatedAt time.Time `db:"created_at"`
}

// AskRequest POST /api/v1/chat/ask 요청 본문
type AskRequest struct {
	Question  string `json:"question"`
	SessionID string `json:"session_id,omitempty"`
}

// AskResponse POST /api/v1/chat/ask 응답 본문
type AskResponse struct {
	Answer    string       `json:"answer"`
	SessionID string       `json:"session_id"`
	Sources   []FishSource `json:"sources,omitempty"`
}

// FishSource RAG 검색으로 참조된 어종 출처
type FishSource struct {
	ID   int64  `json:"id"`
	Name string `json:"name"`
}
