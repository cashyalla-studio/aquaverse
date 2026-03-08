-- Chat rooms: 1 room per trade
CREATE TABLE IF NOT EXISTS chat_rooms (
    id          BIGSERIAL PRIMARY KEY,
    trade_id    BIGINT NOT NULL UNIQUE REFERENCES trades(id) ON DELETE CASCADE,
    buyer_id    UUID NOT NULL REFERENCES users(id),
    seller_id   UUID NOT NULL REFERENCES users(id),
    is_active   BOOLEAN NOT NULL DEFAULT TRUE,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE INDEX idx_chat_rooms_buyer  ON chat_rooms(buyer_id);
CREATE INDEX idx_chat_rooms_seller ON chat_rooms(seller_id);

-- Chat messages (persistent)
CREATE TABLE IF NOT EXISTS chat_messages (
    id          BIGSERIAL PRIMARY KEY,
    room_id     BIGINT NOT NULL REFERENCES chat_rooms(id) ON DELETE CASCADE,
    sender_id   UUID NOT NULL REFERENCES users(id),
    content     TEXT NOT NULL CHECK (char_length(content) BETWEEN 1 AND 2000),
    msg_type    VARCHAR(20) NOT NULL DEFAULT 'TEXT',  -- TEXT, IMAGE, SYSTEM
    is_deleted  BOOLEAN NOT NULL DEFAULT FALSE,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE INDEX idx_chat_messages_room ON chat_messages(room_id, created_at DESC);

-- Phone verifications
CREATE TABLE IF NOT EXISTS user_phone_verifications (
    id           BIGSERIAL PRIMARY KEY,
    user_id      UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    phone_number VARCHAR(20) NOT NULL,
    code         CHAR(6) NOT NULL,
    is_verified  BOOLEAN NOT NULL DEFAULT FALSE,
    expires_at   TIMESTAMPTZ NOT NULL,
    verified_at  TIMESTAMPTZ,
    created_at   TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE UNIQUE INDEX idx_phone_verification_user ON user_phone_verifications(user_id) WHERE is_verified = TRUE;

-- Add phone_verified + account_age to users
ALTER TABLE users
    ADD COLUMN IF NOT EXISTS phone_number  VARCHAR(20),
    ADD COLUMN IF NOT EXISTS phone_verified BOOLEAN NOT NULL DEFAULT FALSE,
    ADD COLUMN IF NOT EXISTS account_created_at_snapshot TIMESTAMPTZ;

-- New account trade limit: 30,000 KRW cap for accounts < 30 days old & unverified phone
CREATE TABLE IF NOT EXISTS new_account_trade_limits (
    user_id          UUID PRIMARY KEY REFERENCES users(id),
    daily_limit_krw  NUMERIC(15,2) NOT NULL DEFAULT 30000,
    used_today_krw   NUMERIC(15,2) NOT NULL DEFAULT 0,
    last_reset_date  DATE NOT NULL DEFAULT CURRENT_DATE
);
