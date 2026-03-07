-- CITES 어류 목록 (CITES 부속서 I, II 중 어류)
CREATE TABLE IF NOT EXISTS cites_fish (
    id           BIGSERIAL PRIMARY KEY,
    scientific_name VARCHAR(255) NOT NULL UNIQUE,  -- 학명 (Genus species)
    common_names    TEXT[],                          -- 영문 속명 배열
    appendix        CHAR(3) NOT NULL,                -- 'I', 'II', 'III'
    is_blocked      BOOLEAN NOT NULL DEFAULT TRUE,  -- I은 TRUE, II는 TRUE(경고+확인), III은 FALSE(경고만)
    notes           TEXT,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- 초기 데이터: CITES 부속서 I 어류 주요 종 (거래 완전 차단)
INSERT INTO cites_fish (scientific_name, common_names, appendix, is_blocked, notes) VALUES
('Carcharodon carcharias',   ARRAY['Great White Shark', '백상아리'],        'I',  TRUE, 'CITES Appendix II - 실제로는 II이지만 예시'),
('Rhincodon typus',          ARRAY['Whale Shark', '고래상어'],               'II', TRUE, 'CITES Appendix II'),
('Pristis spp.',             ARRAY['Sawfish', '톱상어'],                     'I',  TRUE, 'CITES Appendix I'),
('Latimeria spp.',           ARRAY['Coelacanth', '실러캔스'],               'I',  TRUE, 'CITES Appendix I - 살아있는 화석'),
('Hippocampus spp.',         ARRAY['Seahorse', '해마'],                      'II', FALSE, 'CITES Appendix II - 경고만'),
('Bolbometopon muricatum',   ARRAY['Bumphead Parrotfish', '혹돔앵무고기'], 'II', FALSE, 'CITES Appendix II'),
('Cheilinus undulatus',      ARRAY['Napoleon Wrasse', '나폴레옹 놀래기'],  'II', TRUE,  'CITES Appendix II - 상업적 거래 제한'),
('Epinephelus itajara',      ARRAY['Goliath Grouper', '거대 그루퍼'],      'II', FALSE, 'CITES Appendix II'),
('Thunnus thynnus',          ARRAY['Atlantic Bluefin Tuna', '대서양 참다랑어'], 'II', FALSE, 'CITES Appendix II'),
('Anguilla anguilla',        ARRAY['European Eel', '유럽 장어'],             'II', FALSE, 'CITES Appendix II'),
('Pangasianodon gigas',      ARRAY['Mekong Giant Catfish', '메콩 대형메기'], 'I', TRUE, 'CITES Appendix I'),
('Potamotrygon spp.',        ARRAY['Freshwater Stingray', '민물가오리'],    'III', FALSE, 'CITES Appendix III - 경고만'),
('Arapaima gigas',           ARRAY['Arapaima', '아라파이마'],               'II', TRUE, 'CITES Appendix II')
ON CONFLICT (scientific_name) DO NOTHING;

-- 한국 생태계 교란 어류 (환경부 고시)
CREATE TABLE IF NOT EXISTS invasive_species_kr (
    id              BIGSERIAL PRIMARY KEY,
    scientific_name VARCHAR(255) NOT NULL UNIQUE,
    korean_name     VARCHAR(100),
    common_names    TEXT[],
    designated_date DATE,
    notes           TEXT,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

INSERT INTO invasive_species_kr (scientific_name, korean_name, common_names, notes) VALUES
('Lepomis macrochirus',        '블루길',       ARRAY['Bluegill'],                    '환경부 지정 생태계 교란종'),
('Micropterus salmoides',      '배스',         ARRAY['Largemouth Bass'],             '환경부 지정 생태계 교란종'),
('Trachemys scripta elegans',  NULL,           ARRAY['Red-eared Slider'],            '생태계 교란 동물 (어류 아님, 참고용)'),
('Carassius auratus',          '황금붕어',     ARRAY['Goldfish'],                    '일부 지역 방류 금지'),
('Ameiurus nebulosus',         '메기(외래)',   ARRAY['Brown Bullhead'],              '주의 어종')
ON CONFLICT (scientific_name) DO NOTHING;

-- 에스크로 거래 테이블
CREATE TABLE IF NOT EXISTS escrow_transactions (
    id              BIGSERIAL PRIMARY KEY,
    trade_id        BIGINT NOT NULL UNIQUE REFERENCES trades(id) ON DELETE CASCADE,
    amount          NUMERIC(15,2) NOT NULL,
    currency        VARCHAR(3) NOT NULL DEFAULT 'KRW',
    status          VARCHAR(20) NOT NULL DEFAULT 'PENDING',
    -- PENDING → FUNDED → RELEASED / REFUNDED / DISPUTED
    funded_at       TIMESTAMPTZ,
    released_at     TIMESTAMPTZ,
    refunded_at     TIMESTAMPTZ,
    dispute_reason  TEXT,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_escrow_trade ON escrow_transactions(trade_id);

-- 이미지 처리 작업 큐
CREATE TABLE IF NOT EXISTS image_jobs (
    id          BIGSERIAL PRIMARY KEY,
    object_key  VARCHAR(500) NOT NULL,  -- MinIO object key
    bucket      VARCHAR(100) NOT NULL,
    status      VARCHAR(20) NOT NULL DEFAULT 'PENDING',  -- PENDING, PROCESSING, DONE, FAILED
    error_msg   TEXT,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    processed_at TIMESTAMPTZ
);

CREATE INDEX idx_image_jobs_status ON image_jobs(status, created_at);
