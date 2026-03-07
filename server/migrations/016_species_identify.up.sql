-- 016_species_identify.up.sql
-- AI 어종 식별 로그

CREATE TABLE IF NOT EXISTS species_identification_log (
    id              BIGSERIAL PRIMARY KEY,
    user_id         UUID REFERENCES users(id) ON DELETE SET NULL,
    image_key       TEXT,                   -- MinIO 이미지 키 (선택)
    candidates      JSONB NOT NULL,         -- [{name, scientific_name, confidence, description}]
    top_fish_id     BIGINT REFERENCES fish_data(id) ON DELETE SET NULL,  -- 최상위 매칭 어종
    model_used      VARCHAR(50) NOT NULL DEFAULT 'claude-haiku-4-5-20251001',
    processing_ms   INT,                    -- 처리 시간 (ms)
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE INDEX idx_species_id_log_user ON species_identification_log(user_id, created_at DESC);
