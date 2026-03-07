-- FCM 토큰 저장
CREATE TABLE IF NOT EXISTS users_fcm_tokens (
    id         BIGSERIAL PRIMARY KEY,
    user_id    UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    token      TEXT NOT NULL UNIQUE,
    platform   VARCHAR(10) NOT NULL DEFAULT 'android', -- 'android' | 'ios' | 'web'
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE INDEX IF NOT EXISTS idx_fcm_tokens_user ON users_fcm_tokens(user_id);

-- 알림 이력
CREATE TABLE IF NOT EXISTS notification_logs (
    id          BIGSERIAL PRIMARY KEY,
    user_id     UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    title       VARCHAR(200) NOT NULL,
    body        TEXT,
    data        JSONB,
    sent_at     TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    success     BOOLEAN NOT NULL DEFAULT FALSE
);
CREATE INDEX IF NOT EXISTS idx_notification_logs_user ON notification_logs(user_id, sent_at DESC);
