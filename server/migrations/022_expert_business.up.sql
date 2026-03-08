-- Expert Connect
CREATE TABLE IF NOT EXISTS expert_profiles (
    id              BIGSERIAL PRIMARY KEY,
    user_id         UUID NOT NULL UNIQUE REFERENCES users(id) ON DELETE CASCADE,
    expert_type     VARCHAR(30) NOT NULL,  -- 'vet', 'breeder', 'aquarist', 'trainer'
    bio             TEXT,
    specialties     JSONB DEFAULT '[]',    -- 전문 어종/동물 (JSON 배열)
    hourly_rate     BIGINT,               -- 원 단위 (NULL = 무료)
    is_verified     BOOLEAN NOT NULL DEFAULT FALSE,
    verified_at     TIMESTAMPTZ,
    rating          NUMERIC(3,2),
    review_count    INT NOT NULL DEFAULT 0,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS consultations (
    id              BIGSERIAL PRIMARY KEY,
    user_id         UUID NOT NULL REFERENCES users(id),
    expert_id       BIGINT NOT NULL REFERENCES expert_profiles(id),
    scheduled_at    TIMESTAMPTZ,
    duration_min    INT DEFAULT 30,
    status          VARCHAR(20) NOT NULL DEFAULT 'pending',
    -- pending, confirmed, completed, cancelled
    question        TEXT,
    answer          TEXT,
    payment_amount  BIGINT,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS expert_reviews (
    id              BIGSERIAL PRIMARY KEY,
    consultation_id BIGINT NOT NULL REFERENCES consultations(id),
    reviewer_id     UUID NOT NULL REFERENCES users(id),
    expert_id       BIGINT NOT NULL REFERENCES expert_profiles(id),
    rating          INT NOT NULL CHECK (rating BETWEEN 1 AND 5),
    comment         TEXT,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Business Hub 업그레이드: 재고 관리
CREATE TABLE IF NOT EXISTS shop_inventory (
    id              BIGSERIAL PRIMARY KEY,
    business_id     BIGINT NOT NULL REFERENCES business_profiles(id) ON DELETE CASCADE,
    fish_data_id    BIGINT REFERENCES fish_data(id),
    custom_name     TEXT,               -- fish_data 없는 경우
    quantity        INT NOT NULL DEFAULT 0,
    price           BIGINT,             -- 판매가 (원)
    cites_status    VARCHAR(20),        -- 자동 체크 결과
    is_available    BOOLEAN NOT NULL DEFAULT TRUE,
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_expert_profiles_type ON expert_profiles(expert_type, is_verified);
CREATE INDEX IF NOT EXISTS idx_consultations_user ON consultations(user_id, created_at DESC);
CREATE INDEX IF NOT EXISTS idx_shop_inventory_business ON shop_inventory(business_id);
