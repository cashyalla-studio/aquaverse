-- pgvector RAG 챗봇 마이그레이션
-- pgvector 익스텐션 설치 시도 (없으면 무시)
DO $$
BEGIN
    CREATE EXTENSION IF NOT EXISTS vector;
EXCEPTION WHEN OTHERS THEN
    RAISE NOTICE 'pgvector extension not available, using JSONB fallback for embeddings';
END;
$$;

-- fish_embeddings: pgvector 없을 때 JSONB fallback 사용
CREATE TABLE IF NOT EXISTS fish_embeddings (
    fish_id     BIGINT PRIMARY KEY REFERENCES fish_data(id) ON DELETE CASCADE,
    embedding   JSONB,          -- float32 배열 (pgvector 없을 때 JSONB 저장)
    content     TEXT NOT NULL,  -- 임베딩된 원본 텍스트
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- RAG 전용 세션 (기존 WebSocket chat_sessions, chat_messages와 이름 충돌 방지)
CREATE TABLE IF NOT EXISTS rag_sessions (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id     UUID REFERENCES users(id) ON DELETE CASCADE,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    last_msg_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- RAG 전용 메시지
CREATE TABLE IF NOT EXISTS rag_messages (
    id          BIGSERIAL PRIMARY KEY,
    session_id  UUID NOT NULL REFERENCES rag_sessions(id) ON DELETE CASCADE,
    role        VARCHAR(10) NOT NULL CHECK (role IN ('user', 'assistant')),
    content     TEXT NOT NULL,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_fish_embeddings_updated ON fish_embeddings(updated_at);
CREATE INDEX IF NOT EXISTS idx_rag_sessions_user ON rag_sessions(user_id, last_msg_at DESC);
CREATE INDEX IF NOT EXISTS idx_rag_messages_session ON rag_messages(session_id, created_at);
