-- 014_social.up.sql
-- Social Graph: 팔로우 + 활동 피드

-- 팔로우 관계 테이블
CREATE TABLE IF NOT EXISTS user_follows (
    follower_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    following_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    PRIMARY KEY (follower_id, following_id),
    CHECK (follower_id != following_id)
);
CREATE INDEX idx_user_follows_following ON user_follows(following_id, created_at DESC);

-- 활동 피드 이벤트
CREATE TABLE IF NOT EXISTS activity_feed (
    id          BIGSERIAL PRIMARY KEY,
    actor_id    UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    verb        VARCHAR(50) NOT NULL,  -- 'LISTED', 'SOLD', 'REVIEWED', 'JOINED'
    object_type VARCHAR(50),           -- 'LISTING', 'TRADE', 'FISH'
    object_id   BIGINT,
    object_data JSONB,                 -- 어종명, 가격 등 스냅샷
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE INDEX idx_activity_feed_actor ON activity_feed(actor_id, created_at DESC);
CREATE INDEX idx_activity_feed_created ON activity_feed(created_at DESC);
