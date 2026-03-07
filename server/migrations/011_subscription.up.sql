-- 구독 플랜
CREATE TABLE IF NOT EXISTS user_subscriptions (
    id              BIGSERIAL PRIMARY KEY,
    user_id         UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE UNIQUE,
    plan            VARCHAR(20) NOT NULL DEFAULT 'FREE', -- FREE, PRO
    status          VARCHAR(20) NOT NULL DEFAULT 'ACTIVE', -- ACTIVE, CANCELLED, EXPIRED
    started_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    expires_at      TIMESTAMPTZ,              -- NULL이면 영구 (관리자 부여)
    billing_key     VARCHAR(200),             -- 토스페이먼츠 빌링키 (정기결제)
    billing_amount  INT NOT NULL DEFAULT 0,  -- 월 결제액 (원)
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_subscriptions_user ON user_subscriptions(user_id);
CREATE INDEX IF NOT EXISTS idx_subscriptions_expires ON user_subscriptions(expires_at) WHERE expires_at IS NOT NULL;

-- 구독 이력
CREATE TABLE IF NOT EXISTS subscription_history (
    id          BIGSERIAL PRIMARY KEY,
    user_id     UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    plan        VARCHAR(20) NOT NULL,
    event       VARCHAR(50) NOT NULL, -- SUBSCRIBED, RENEWED, CANCELLED, EXPIRED
    amount      INT,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
