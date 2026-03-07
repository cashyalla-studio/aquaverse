-- 수조 영상 게시물
CREATE TABLE IF NOT EXISTS video_posts (
    id           BIGSERIAL PRIMARY KEY,
    user_id      UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    title        VARCHAR(200) NOT NULL,
    description  TEXT,
    video_key    VARCHAR(500) NOT NULL,  -- MinIO 오브젝트 키
    thumbnail_key VARCHAR(500),          -- 썸네일 (선택)
    duration_sec INT,                    -- 영상 길이 (초)
    view_count   INT NOT NULL DEFAULT 0,
    like_count   INT NOT NULL DEFAULT 0,
    status       VARCHAR(20) NOT NULL DEFAULT 'ACTIVE', -- ACTIVE, DELETED
    created_at   TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE INDEX IF NOT EXISTS idx_video_posts_user ON video_posts(user_id, created_at DESC);
CREATE INDEX IF NOT EXISTS idx_video_posts_feed ON video_posts(status, created_at DESC);

-- 영상 좋아요
CREATE TABLE IF NOT EXISTS video_likes (
    video_id  BIGINT NOT NULL REFERENCES video_posts(id) ON DELETE CASCADE,
    user_id   UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    PRIMARY KEY (video_id, user_id)
);
