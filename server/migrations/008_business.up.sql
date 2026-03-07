-- users 테이블에 account_type 추가
ALTER TABLE users
    ADD COLUMN IF NOT EXISTS account_type VARCHAR(20) NOT NULL DEFAULT 'PERSONAL',
    ADD COLUMN IF NOT EXISTS business_number VARCHAR(20);

-- 업체 프로필
CREATE TABLE IF NOT EXISTS business_profiles (
    id              BIGSERIAL PRIMARY KEY,
    user_id         UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE UNIQUE,
    store_name      VARCHAR(100) NOT NULL,
    description     TEXT,
    address         TEXT,
    city            VARCHAR(50),
    phone           VARCHAR(20),
    website         VARCHAR(200),
    logo_url        TEXT,
    is_verified     BOOLEAN NOT NULL DEFAULT FALSE,
    verified_at     TIMESTAMPTZ,
    lat             NUMERIC(10,7),
    lng             NUMERIC(10,7),
    business_hours  JSONB,   -- {"mon":"09:00-18:00", ...}
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE INDEX IF NOT EXISTS idx_business_profiles_city ON business_profiles(city);
CREATE INDEX IF NOT EXISTS idx_business_profiles_verified ON business_profiles(is_verified);

-- 업체 리뷰
CREATE TABLE IF NOT EXISTS business_reviews (
    id              BIGSERIAL PRIMARY KEY,
    business_id     BIGINT NOT NULL REFERENCES business_profiles(id) ON DELETE CASCADE,
    reviewer_id     UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    rating          SMALLINT NOT NULL CHECK (rating BETWEEN 1 AND 5),
    content         TEXT,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (business_id, reviewer_id)
);
CREATE INDEX IF NOT EXISTS idx_business_reviews_business ON business_reviews(business_id);
