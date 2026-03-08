-- 뱃지 정의 테이블
CREATE TABLE IF NOT EXISTS badge_definitions (
    code         VARCHAR(50) PRIMARY KEY,
    name         TEXT NOT NULL,
    description  TEXT,
    icon_emoji   VARCHAR(10),
    category     VARCHAR(30) NOT NULL, -- 'care', 'community', 'market', 'collection', 'special'
    condition_type VARCHAR(50) NOT NULL, -- 'post_count', 'trade_count', 'species_count', 'streak_days', 'login_days'
    condition_value INT NOT NULL,
    is_active    BOOLEAN NOT NULL DEFAULT TRUE
);

-- 사용자 획득 뱃지
CREATE TABLE IF NOT EXISTS user_badges (
    id           BIGSERIAL PRIMARY KEY,
    user_id      UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    badge_code   VARCHAR(50) NOT NULL REFERENCES badge_definitions(code),
    earned_at    TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (user_id, badge_code)
);

-- 챌린지 테이블
CREATE TABLE IF NOT EXISTS challenges (
    id           BIGSERIAL PRIMARY KEY,
    title        TEXT NOT NULL,
    description  TEXT,
    badge_code   VARCHAR(50) REFERENCES badge_definitions(code),
    starts_at    TIMESTAMPTZ NOT NULL,
    ends_at      TIMESTAMPTZ NOT NULL,
    condition_type VARCHAR(50) NOT NULL,
    condition_value INT NOT NULL,
    is_active    BOOLEAN NOT NULL DEFAULT TRUE
);

-- 챌린지 참가
CREATE TABLE IF NOT EXISTS challenge_participants (
    challenge_id BIGINT NOT NULL REFERENCES challenges(id) ON DELETE CASCADE,
    user_id      UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    progress     INT NOT NULL DEFAULT 0,
    completed    BOOLEAN NOT NULL DEFAULT FALSE,
    joined_at    TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    completed_at TIMESTAMPTZ,
    PRIMARY KEY (challenge_id, user_id)
);

CREATE INDEX IF NOT EXISTS idx_user_badges_user ON user_badges(user_id);
CREATE INDEX IF NOT EXISTS idx_challenges_active ON challenges(ends_at) WHERE is_active = TRUE;

-- 기본 뱃지 시드 데이터
INSERT INTO badge_definitions (code, name, description, icon_emoji, category, condition_type, condition_value) VALUES
('first_post', '첫 게시물', '커뮤니티에 첫 게시물 작성', '📝', 'community', 'post_count', 1),
('prolific_writer', '다작 작가', '게시물 50개 작성', '✍️', 'community', 'post_count', 50),
('first_trade', '첫 거래', '첫 번째 거래 완료', '🤝', 'market', 'trade_count', 1),
('collector_10', '10종 수집가', '10가지 어종 사육', '🐠', 'collection', 'species_count', 10),
('care_streak_7', '7일 케어', '7일 연속 케어 기록', '🔥', 'care', 'streak_days', 7),
('care_streak_30', '30일 케어', '30일 연속 케어 기록', '💎', 'care', 'streak_days', 30)
ON CONFLICT (code) DO NOTHING;
