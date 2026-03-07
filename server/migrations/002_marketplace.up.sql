-- ============================================================
-- AquaVerse - 002 Marketplace (분양/입양)
-- ============================================================

-- 분양 매물
CREATE TABLE listings (
    id                  BIGSERIAL PRIMARY KEY,
    seller_id           UUID NOT NULL REFERENCES users(id),
    fish_data_id        BIGINT REFERENCES fish_data(id),
    scientific_name     VARCHAR(255),
    common_name         VARCHAR(255) NOT NULL,
    quantity            SMALLINT NOT NULL DEFAULT 1,
    age_months          SMALLINT,
    size_cm             DECIMAL(5,1),
    sex                 VARCHAR(10) NOT NULL DEFAULT 'UNKNOWN'
                        CHECK (sex IN ('MALE','FEMALE','UNKNOWN','MIXED')),
    health_status       VARCHAR(20) NOT NULL
                        CHECK (health_status IN ('EXCELLENT','GOOD','DISEASE_HISTORY','UNDER_TREATMENT')),
    disease_history     TEXT,
    bred_by_seller      BOOLEAN NOT NULL DEFAULT FALSE,
    tank_size_liters    SMALLINT,
    tank_mates          JSONB DEFAULT '[]',
    feeding_type        VARCHAR(100),
    water_ph            DECIMAL(3,1),
    water_temp_c        DECIMAL(4,1),
    price               DECIMAL(14,2) NOT NULL DEFAULT 0,
    price_usd           DECIMAL(12,2),
    currency            VARCHAR(3) NOT NULL DEFAULT 'KRW',
    price_negotiable    BOOLEAN NOT NULL DEFAULT FALSE,
    trade_type          VARCHAR(20) NOT NULL DEFAULT 'ALL'
                        CHECK (trade_type IN ('DIRECT','COURIER','AQUA_COURIER','ALL')),
    allow_international BOOLEAN NOT NULL DEFAULT FALSE,
    allowed_countries   JSONB DEFAULT '[]',
    location            GEOGRAPHY(POINT, 4326),
    location_text       VARCHAR(200) NOT NULL DEFAULT '',
    country_code        VARCHAR(2) NOT NULL DEFAULT 'KR',
    title               VARCHAR(200) NOT NULL,
    description         TEXT,
    image_urls          JSONB NOT NULL DEFAULT '[]',
    status              VARCHAR(20) NOT NULL DEFAULT 'ACTIVE'
                        CHECK (status IN ('DRAFT','ACTIVE','RESERVED','SOLD','EXPIRED','HIDDEN','DELETED')),
    view_count          INT NOT NULL DEFAULT 0,
    favorite_count      INT NOT NULL DEFAULT 0,
    auto_hold           BOOLEAN NOT NULL DEFAULT FALSE,
    hold_reason         VARCHAR(200),
    ip_address          INET,
    device_fingerprint  VARCHAR(64),
    created_at          TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at          TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    expires_at          TIMESTAMPTZ,
    sold_at             TIMESTAMPTZ
);

CREATE INDEX idx_listings_status ON listings (status, created_at DESC);
CREATE INDEX idx_listings_seller ON listings (seller_id);
CREATE INDEX idx_listings_fish ON listings (fish_data_id);
CREATE INDEX idx_listings_country ON listings (country_code, status);
CREATE INDEX idx_listings_price ON listings (price);
CREATE INDEX idx_listings_location ON listings USING GIST(location);
-- 전문 검색
CREATE INDEX idx_listings_fts ON listings
    USING GIN(to_tsvector('simple', common_name || ' ' || coalesce(scientific_name,'')));

-- 즐겨찾기
CREATE TABLE listing_favorites (
    listing_id BIGINT NOT NULL REFERENCES listings(id) ON DELETE CASCADE,
    user_id    UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    PRIMARY KEY (listing_id, user_id)
);

-- 거래
CREATE TABLE trades (
    id                       BIGSERIAL PRIMARY KEY,
    listing_id               BIGINT NOT NULL REFERENCES listings(id),
    seller_id                UUID NOT NULL REFERENCES users(id),
    buyer_id                 UUID NOT NULL REFERENCES users(id),
    trade_type               VARCHAR(20) NOT NULL,
    agreed_price             DECIMAL(14,2) NOT NULL,
    currency                 VARCHAR(3) NOT NULL DEFAULT 'KRW',
    escrow_enabled           BOOLEAN NOT NULL DEFAULT FALSE,
    escrow_status            VARCHAR(20) CHECK (escrow_status IN ('HELD','RELEASED','REFUNDED')),
    payment_method           VARCHAR(30),
    payment_ref              VARCHAR(100),
    tracking_number          VARCHAR(100),
    courier_name             VARCHAR(50),
    delivery_notes           TEXT,
    status                   VARCHAR(20) NOT NULL DEFAULT 'NEGOTIATING'
                             CHECK (status IN ('NEGOTIATING','CONFIRMED','IN_DELIVERY','DELIVERED','COMPLETED','CANCELLED','DISPUTED')),
    meetup_location          GEOGRAPHY(POINT, 4326),
    meetup_confirmed_seller  BOOLEAN NOT NULL DEFAULT FALSE,
    meetup_confirmed_buyer   BOOLEAN NOT NULL DEFAULT FALSE,
    arrival_confirmed_at     TIMESTAMPTZ,
    arrival_photo_urls       JSONB DEFAULT '[]',
    health_confirmed         BOOLEAN,
    disputed_at              TIMESTAMPTZ,
    dispute_reason           TEXT,
    dispute_resolved_at      TIMESTAMPTZ,
    admin_note               TEXT,
    created_at               TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at               TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    completed_at             TIMESTAMPTZ
);

CREATE INDEX idx_trades_listing ON trades (listing_id);
CREATE INDEX idx_trades_seller ON trades (seller_id, status);
CREATE INDEX idx_trades_buyer ON trades (buyer_id, status);
CREATE INDEX idx_trades_status ON trades (status);

-- 거래 리뷰 (양방향)
CREATE TABLE trade_reviews (
    id                   BIGSERIAL PRIMARY KEY,
    trade_id             BIGINT NOT NULL REFERENCES trades(id),
    reviewer_id          UUID NOT NULL REFERENCES users(id),
    reviewee_id          UUID NOT NULL REFERENCES users(id),
    rating               DECIMAL(2,1) NOT NULL CHECK (rating >= 1 AND rating <= 5),
    rating_communication SMALLINT CHECK (rating_communication BETWEEN 1 AND 5),
    rating_accuracy      SMALLINT CHECK (rating_accuracy BETWEEN 1 AND 5),
    rating_packaging     SMALLINT CHECK (rating_packaging BETWEEN 1 AND 5),
    rating_health        SMALLINT CHECK (rating_health BETWEEN 1 AND 5),
    comment              TEXT,
    tags                 JSONB DEFAULT '[]',
    created_at           TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (trade_id, reviewer_id)
);

CREATE INDEX idx_reviews_reviewee ON trade_reviews (reviewee_id);

-- 신뢰도 집계 (이벤트 발생 시 갱신)
CREATE TABLE user_trust_scores (
    user_id               UUID PRIMARY KEY REFERENCES users(id) ON DELETE CASCADE,
    trust_score           DECIMAL(5,2) NOT NULL DEFAULT 36.5,
    total_trades          INT NOT NULL DEFAULT 0,
    completed_trades      INT NOT NULL DEFAULT 0,
    avg_rating            DECIMAL(3,2),
    response_rate         DECIMAL(5,4),
    badges                JSONB NOT NULL DEFAULT '[]',
    fraud_report_count    INT NOT NULL DEFAULT 0,
    confirmed_fraud_count INT NOT NULL DEFAULT 0,
    updated_at            TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- 어종 알림 구독
CREATE TABLE fish_watch_subscriptions (
    id                    BIGSERIAL PRIMARY KEY,
    user_id               UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    fish_data_id          BIGINT REFERENCES fish_data(id),
    custom_species        VARCHAR(255),
    max_price             DECIMAL(14,2),
    location              GEOGRAPHY(POINT, 4326),
    radius_km             DECIMAL(6,1) NOT NULL DEFAULT 50,
    include_international BOOLEAN NOT NULL DEFAULT FALSE,
    notify_push           BOOLEAN NOT NULL DEFAULT TRUE,
    notify_email          BOOLEAN NOT NULL DEFAULT FALSE,
    notify_in_app         BOOLEAN NOT NULL DEFAULT TRUE,
    active                BOOLEAN NOT NULL DEFAULT TRUE,
    last_notified_at      TIMESTAMPTZ,
    match_count           INT NOT NULL DEFAULT 0,
    created_at            TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (user_id, fish_data_id, custom_species)
);

CREATE INDEX idx_watch_user ON fish_watch_subscriptions (user_id);
CREATE INDEX idx_watch_fish ON fish_watch_subscriptions (fish_data_id);
CREATE INDEX idx_watch_active ON fish_watch_subscriptions (active);
CREATE INDEX idx_watch_location ON fish_watch_subscriptions USING GIST(location);

-- 알림 로그
CREATE TABLE notification_log (
    id                BIGSERIAL PRIMARY KEY,
    user_id           UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    subscription_id   BIGINT REFERENCES fish_watch_subscriptions(id),
    listing_id        BIGINT REFERENCES listings(id),
    notification_type VARCHAR(30) NOT NULL,
    channel           VARCHAR(20) NOT NULL CHECK (channel IN ('PUSH','EMAIL','IN_APP')),
    title             VARCHAR(200),
    body              TEXT,
    sent_at           TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    read_at           TIMESTAMPTZ
);

CREATE INDEX idx_notif_user ON notification_log (user_id, sent_at DESC);
CREATE INDEX idx_notif_unread ON notification_log (user_id, read_at) WHERE read_at IS NULL;

-- 사기 신고
CREATE TABLE fraud_reports (
    id               BIGSERIAL PRIMARY KEY,
    reporter_id      UUID NOT NULL REFERENCES users(id),
    reported_user_id UUID NOT NULL REFERENCES users(id),
    listing_id       BIGINT REFERENCES listings(id),
    trade_id         BIGINT REFERENCES trades(id),
    report_type      VARCHAR(30) NOT NULL
                     CHECK (report_type IN ('FRAUD','NO_SHOW','DECEASED_FISH','WRONG_SPECIES','MISREPRESENTATION')),
    description      TEXT,
    evidence_urls    JSONB NOT NULL DEFAULT '[]',
    status           VARCHAR(20) NOT NULL DEFAULT 'PENDING'
                     CHECK (status IN ('PENDING','UNDER_REVIEW','CONFIRMED','REJECTED')),
    admin_note       TEXT,
    created_at       TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    resolved_at      TIMESTAMPTZ
);

CREATE INDEX idx_fraud_reported ON fraud_reports (reported_user_id);
CREATE INDEX idx_fraud_status ON fraud_reports (status);

-- ============================================================
-- 신뢰도 점수 자동 갱신 함수
-- ============================================================

CREATE OR REPLACE FUNCTION update_trust_score(p_user_id UUID)
RETURNS VOID AS $$
DECLARE
    v_completed   INT;
    v_total       INT;
    v_avg_rating  DECIMAL;
    v_fraud_count INT;
    v_score       DECIMAL := 36.5;
BEGIN
    SELECT
        COUNT(*) FILTER (WHERE status = 'COMPLETED'),
        COUNT(*)
    INTO v_completed, v_total
    FROM trades
    WHERE seller_id = p_user_id OR buyer_id = p_user_id;

    SELECT AVG(rating)
    INTO v_avg_rating
    FROM trade_reviews
    WHERE reviewee_id = p_user_id;

    SELECT COUNT(*)
    INTO v_fraud_count
    FROM fraud_reports
    WHERE reported_user_id = p_user_id AND status = 'CONFIRMED';

    -- 거래 완료 가산 (최대 +20)
    v_score := v_score + LEAST(v_completed * 0.5, 20);

    -- 평균 별점 반영
    IF v_avg_rating IS NOT NULL THEN
        v_score := v_score + (v_avg_rating - 3) * 2; -- 5점=+4, 3점=0, 1점=-4
    END IF;

    -- 사기 확정 감산
    v_score := v_score - (v_fraud_count * 20);

    -- 0~100 클램프
    v_score := GREATEST(0, LEAST(100, v_score));

    INSERT INTO user_trust_scores (user_id, trust_score, total_trades, completed_trades, avg_rating, fraud_report_count, confirmed_fraud_count, updated_at)
    VALUES (p_user_id, v_score, v_total, v_completed, v_avg_rating, 0, v_fraud_count, NOW())
    ON CONFLICT (user_id) DO UPDATE SET
        trust_score           = v_score,
        total_trades          = v_total,
        completed_trades      = v_completed,
        avg_rating            = v_avg_rating,
        confirmed_fraud_count = v_fraud_count,
        updated_at            = NOW();

    -- users 테이블도 동기화
    UPDATE users SET trust_score = v_score, updated_at = NOW()
    WHERE id = p_user_id;
END;
$$ LANGUAGE plpgsql;
