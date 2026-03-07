-- ============================================================
-- AquaVerse - 001 Initial Schema
-- PostgreSQL 16+
-- ============================================================

-- UUID 확장
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";
-- PostGIS (위치 기반 검색)
CREATE EXTENSION IF NOT EXISTS postgis;

-- ============================================================
-- 사용자 시스템
-- ============================================================

CREATE TABLE users (
    id               UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    email            VARCHAR(320) NOT NULL UNIQUE,
    username         VARCHAR(50) NOT NULL UNIQUE,
    password_hash    TEXT NOT NULL,
    display_name     VARCHAR(100) NOT NULL,
    avatar_url       TEXT,
    bio              TEXT,
    preferred_locale VARCHAR(10) NOT NULL DEFAULT 'en-US',
    role             VARCHAR(20) NOT NULL DEFAULT 'USER'
                     CHECK (role IN ('USER', 'MODERATOR', 'ADMIN')),
    email_verified   BOOLEAN NOT NULL DEFAULT FALSE,
    phone_verified   BOOLEAN NOT NULL DEFAULT FALSE,
    phone_number     VARCHAR(30),
    trust_score      DECIMAL(5,2) NOT NULL DEFAULT 36.5,
    country_code     VARCHAR(2),
    city             VARCHAR(100),
    is_active        BOOLEAN NOT NULL DEFAULT TRUE,
    is_banned        BOOLEAN NOT NULL DEFAULT FALSE,
    ban_reason       TEXT,
    created_at       TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at       TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    last_login_at    TIMESTAMPTZ
);

CREATE INDEX idx_users_email ON users (email);
CREATE INDEX idx_users_username ON users (username);
CREATE INDEX idx_users_role ON users (role);

CREATE TABLE refresh_tokens (
    id           UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id      UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    token_hash   VARCHAR(64) NOT NULL UNIQUE,
    device_info  TEXT,
    expires_at   TIMESTAMPTZ NOT NULL,
    created_at   TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    revoked_at   TIMESTAMPTZ
);

CREATE INDEX idx_refresh_tokens_user ON refresh_tokens (user_id);

-- ============================================================
-- 열대어 백과사전
-- ============================================================

CREATE TABLE fish_data (
    id                   BIGSERIAL PRIMARY KEY,
    scientific_name      VARCHAR(255) NOT NULL UNIQUE,
    genus                VARCHAR(100),
    species              VARCHAR(100),
    family               VARCHAR(100),
    order_name           VARCHAR(100),
    class_name           VARCHAR(100),
    primary_common_name  VARCHAR(255),
    care_level           VARCHAR(20) CHECK (care_level IN ('BEGINNER', 'INTERMEDIATE', 'EXPERT')),
    temperament          VARCHAR(20) CHECK (temperament IN ('PEACEFUL', 'SEMI_AGGRESSIVE', 'AGGRESSIVE')),
    max_size_cm          DECIMAL(6,2),
    lifespan_years       DECIMAL(4,1),
    ph_min               DECIMAL(4,2),
    ph_max               DECIMAL(4,2),
    temp_min_c           DECIMAL(4,1),
    temp_max_c           DECIMAL(4,1),
    hardness_min_dgh     DECIMAL(6,2),
    hardness_max_dgh     DECIMAL(6,2),
    min_tank_size_liters INT,
    diet_type            VARCHAR(20) CHECK (diet_type IN ('OMNIVORE', 'CARNIVORE', 'HERBIVORE')),
    diet_notes           TEXT,
    breeding_notes       TEXT,
    care_notes           TEXT,
    primary_image_url    TEXT,
    image_license        VARCHAR(100),
    image_author         VARCHAR(255),
    primary_source       VARCHAR(50) NOT NULL DEFAULT 'manual',
    source_url           TEXT,
    license              VARCHAR(100),
    license_url          TEXT,
    attribution          TEXT,
    publish_status       VARCHAR(20) NOT NULL DEFAULT 'DRAFT'
                         CHECK (publish_status IN ('DRAFT','AI_PROCESSING','TRANSLATING','PUBLISHED','REJECTED')),
    quality_score        DECIMAL(5,2) NOT NULL DEFAULT 0,
    ai_enriched          BOOLEAN NOT NULL DEFAULT FALSE,
    ai_enriched_at       TIMESTAMPTZ,
    created_at           TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at           TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    published_at         TIMESTAMPTZ
);

CREATE INDEX idx_fish_data_status ON fish_data (publish_status);
CREATE INDEX idx_fish_data_family ON fish_data (family);
CREATE INDEX idx_fish_data_quality ON fish_data (quality_score DESC);
CREATE INDEX idx_fish_data_scientific ON fish_data (scientific_name);
-- 전문 검색
CREATE INDEX idx_fish_data_fts ON fish_data
    USING GIN(to_tsvector('english', coalesce(scientific_name,'') || ' ' || coalesce(primary_common_name,'')));

CREATE TABLE fish_translations (
    id                 BIGSERIAL PRIMARY KEY,
    fish_data_id       BIGINT NOT NULL REFERENCES fish_data(id) ON DELETE CASCADE,
    locale             VARCHAR(10) NOT NULL,
    common_name        VARCHAR(255),
    care_level_label   VARCHAR(50),
    temperament_label  VARCHAR(50),
    diet_notes         TEXT,
    breeding_notes     TEXT,
    care_notes         TEXT,
    translation_source VARCHAR(20) NOT NULL DEFAULT 'AI'
                       CHECK (translation_source IN ('AI','HUMAN','COMMUNITY')),
    translated_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    verified           BOOLEAN NOT NULL DEFAULT FALSE,
    verified_at        TIMESTAMPTZ,
    UNIQUE (fish_data_id, locale)
);

CREATE INDEX idx_fish_translations_locale ON fish_translations (locale);
CREATE INDEX idx_fish_translations_fish ON fish_translations (fish_data_id);

CREATE TABLE fish_synonyms (
    id           BIGSERIAL PRIMARY KEY,
    fish_data_id BIGINT NOT NULL REFERENCES fish_data(id) ON DELETE CASCADE,
    synonym_type VARCHAR(20) NOT NULL CHECK (synonym_type IN ('SCIENTIFIC','COMMON','TRADE')),
    name         VARCHAR(255) NOT NULL,
    locale       VARCHAR(10),
    source       VARCHAR(50),
    UNIQUE (fish_data_id, name, synonym_type)
);

CREATE INDEX idx_fish_synonyms_name ON fish_synonyms (name);

-- ============================================================
-- 크롤링 시스템
-- ============================================================

CREATE TABLE crawl_jobs (
    id              BIGSERIAL PRIMARY KEY,
    source_name     VARCHAR(50) NOT NULL,
    source_url      TEXT NOT NULL,
    job_type        VARCHAR(30) NOT NULL CHECK (job_type IN ('FULL','INCREMENTAL','SINGLE')),
    status          VARCHAR(20) NOT NULL DEFAULT 'PENDING'
                    CHECK (status IN ('PENDING','RUNNING','COMPLETED','FAILED','SKIPPED')),
    items_found     INT NOT NULL DEFAULT 0,
    items_processed INT NOT NULL DEFAULT 0,
    items_failed    INT NOT NULL DEFAULT 0,
    error_message   TEXT,
    started_at      TIMESTAMPTZ,
    completed_at    TIMESTAMPTZ,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_crawl_jobs_status ON crawl_jobs (status);
CREATE INDEX idx_crawl_jobs_source ON crawl_jobs (source_name, created_at DESC);

CREATE TABLE raw_crawl_data (
    id           BIGSERIAL PRIMARY KEY,
    crawl_job_id BIGINT REFERENCES crawl_jobs(id),
    source_name  VARCHAR(50) NOT NULL,
    source_id    VARCHAR(500) NOT NULL,
    source_url   TEXT,
    raw_json     JSONB NOT NULL,
    content_hash VARCHAR(64) NOT NULL,
    parse_status VARCHAR(20) NOT NULL DEFAULT 'PENDING'
                 CHECK (parse_status IN ('PENDING','PROCESSED','FAILED','SKIPPED')),
    fish_data_id BIGINT REFERENCES fish_data(id),
    crawled_at   TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    processed_at TIMESTAMPTZ,
    UNIQUE (source_name, source_id),
    UNIQUE (content_hash)
);

CREATE INDEX idx_raw_crawl_parse_status ON raw_crawl_data (parse_status);
CREATE INDEX idx_raw_crawl_crawled_at ON raw_crawl_data (crawled_at DESC);

CREATE TABLE data_quality_log (
    id           BIGSERIAL PRIMARY KEY,
    fish_data_id BIGINT NOT NULL REFERENCES fish_data(id),
    scorer_type  VARCHAR(30) NOT NULL CHECK (scorer_type IN ('AUTO','AI','MANUAL')),
    quality_score DECIMAL(5,2) NOT NULL,
    field_scores JSONB,
    missing_fields JSONB,
    issues       JSONB,
    scored_at    TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_quality_log_fish ON data_quality_log (fish_data_id);

-- ============================================================
-- 커뮤니티 (로케일별 엄격 분리)
-- ============================================================

CREATE TABLE boards (
    id          BIGSERIAL PRIMARY KEY,
    locale      VARCHAR(10) NOT NULL,
    category    VARCHAR(30) NOT NULL
                CHECK (category IN ('GENERAL','QUESTION','SHOWCASE','BREEDING','DISEASES','EQUIPMENT','NEWS')),
    name        VARCHAR(100) NOT NULL,
    description TEXT,
    is_rtl      BOOLEAN NOT NULL DEFAULT FALSE,
    sort_order  INT NOT NULL DEFAULT 0,
    post_count  INT NOT NULL DEFAULT 0,
    is_active   BOOLEAN NOT NULL DEFAULT TRUE,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (locale, category)
);

CREATE INDEX idx_boards_locale ON boards (locale, is_active);

-- 기본 게시판 초기 데이터 (13개 로케일 × 7개 카테고리 = 91개)
INSERT INTO boards (locale, category, name, is_rtl, sort_order) VALUES
-- 한국어
('ko', 'GENERAL',   '자유게시판', FALSE, 1),
('ko', 'QUESTION',  '질문/답변',  FALSE, 2),
('ko', 'SHOWCASE',  '수조 자랑',  FALSE, 3),
('ko', 'BREEDING',  '번식 일지',  FALSE, 4),
('ko', 'DISEASES',  '질병/치료',  FALSE, 5),
('ko', 'EQUIPMENT', '장비/용품',  FALSE, 6),
('ko', 'NEWS',      '뉴스/공지',  FALSE, 7),
-- 영어 (US)
('en-US', 'GENERAL',   'General',           FALSE, 1),
('en-US', 'QUESTION',  'Q&A',               FALSE, 2),
('en-US', 'SHOWCASE',  'Tank Showcase',     FALSE, 3),
('en-US', 'BREEDING',  'Breeding Journals', FALSE, 4),
('en-US', 'DISEASES',  'Disease & Health',  FALSE, 5),
('en-US', 'EQUIPMENT', 'Equipment',         FALSE, 6),
('en-US', 'NEWS',      'News & Announcements', FALSE, 7),
-- 영어 (GB)
('en-GB', 'GENERAL',   'General',           FALSE, 1),
('en-GB', 'QUESTION',  'Q&A',               FALSE, 2),
('en-GB', 'SHOWCASE',  'Tank Showcase',     FALSE, 3),
('en-GB', 'BREEDING',  'Breeding Journals', FALSE, 4),
('en-GB', 'DISEASES',  'Disease & Health',  FALSE, 5),
('en-GB', 'EQUIPMENT', 'Equipment',         FALSE, 6),
('en-GB', 'NEWS',      'News & Announcements', FALSE, 7),
-- 영어 (AU)
('en-AU', 'GENERAL',   'General',           FALSE, 1),
('en-AU', 'QUESTION',  'Q&A',               FALSE, 2),
('en-AU', 'SHOWCASE',  'Tank Showcase',     FALSE, 3),
('en-AU', 'BREEDING',  'Breeding Journals', FALSE, 4),
('en-AU', 'DISEASES',  'Disease & Health',  FALSE, 5),
('en-AU', 'EQUIPMENT', 'Equipment',         FALSE, 6),
('en-AU', 'NEWS',      'News & Announcements', FALSE, 7),
-- 일본어
('ja', 'GENERAL',   '自由掲示板',       FALSE, 1),
('ja', 'QUESTION',  '質問・回答',       FALSE, 2),
('ja', 'SHOWCASE',  '水槽紹介',         FALSE, 3),
('ja', 'BREEDING',  '繁殖日記',         FALSE, 4),
('ja', 'DISEASES',  '病気・治療',       FALSE, 5),
('ja', 'EQUIPMENT', '機材・用品',       FALSE, 6),
('ja', 'NEWS',      'ニュース・お知らせ', FALSE, 7),
-- 중국어 간체
('zh-CN', 'GENERAL',   '自由讨论',    FALSE, 1),
('zh-CN', 'QUESTION',  '问答',        FALSE, 2),
('zh-CN', 'SHOWCASE',  '鱼缸展示',    FALSE, 3),
('zh-CN', 'BREEDING',  '繁殖日记',    FALSE, 4),
('zh-CN', 'DISEASES',  '疾病与治疗',  FALSE, 5),
('zh-CN', 'EQUIPMENT', '设备与器材',  FALSE, 6),
('zh-CN', 'NEWS',      '新闻公告',    FALSE, 7),
-- 중국어 번체
('zh-TW', 'GENERAL',   '自由討論',    FALSE, 1),
('zh-TW', 'QUESTION',  '問答',        FALSE, 2),
('zh-TW', 'SHOWCASE',  '魚缸展示',    FALSE, 3),
('zh-TW', 'BREEDING',  '繁殖日記',    FALSE, 4),
('zh-TW', 'DISEASES',  '疾病與治療',  FALSE, 5),
('zh-TW', 'EQUIPMENT', '設備與器材',  FALSE, 6),
('zh-TW', 'NEWS',      '新聞公告',    FALSE, 7),
-- 독일어
('de', 'GENERAL',   'Allgemein',         FALSE, 1),
('de', 'QUESTION',  'Fragen & Antworten', FALSE, 2),
('de', 'SHOWCASE',  'Aquarium Showcase', FALSE, 3),
('de', 'BREEDING',  'Zucht-Tagebuch',    FALSE, 4),
('de', 'DISEASES',  'Krankheiten',       FALSE, 5),
('de', 'EQUIPMENT', 'Technik & Zubehör', FALSE, 6),
('de', 'NEWS',      'News & Ankündigungen', FALSE, 7),
-- 프랑스어 (FR)
('fr-FR', 'GENERAL',   'Général',              FALSE, 1),
('fr-FR', 'QUESTION',  'Questions & Réponses', FALSE, 2),
('fr-FR', 'SHOWCASE',  'Vitrine Aquarium',     FALSE, 3),
('fr-FR', 'BREEDING',  'Journal de Élevage',   FALSE, 4),
('fr-FR', 'DISEASES',  'Maladies & Santé',     FALSE, 5),
('fr-FR', 'EQUIPMENT', 'Équipement',           FALSE, 6),
('fr-FR', 'NEWS',      'Actualités',           FALSE, 7),
-- 프랑스어 (CA)
('fr-CA', 'GENERAL',   'Général',              FALSE, 1),
('fr-CA', 'QUESTION',  'Questions & Réponses', FALSE, 2),
('fr-CA', 'SHOWCASE',  'Vitrine Aquarium',     FALSE, 3),
('fr-CA', 'BREEDING',  'Journal de Élevage',   FALSE, 4),
('fr-CA', 'DISEASES',  'Maladies & Santé',     FALSE, 5),
('fr-CA', 'EQUIPMENT', 'Équipement',           FALSE, 6),
('fr-CA', 'NEWS',      'Actualités',           FALSE, 7),
-- 스페인어
('es', 'GENERAL',   'General',             FALSE, 1),
('es', 'QUESTION',  'Preguntas',           FALSE, 2),
('es', 'SHOWCASE',  'Acuario Showcase',    FALSE, 3),
('es', 'BREEDING',  'Diario de Cría',      FALSE, 4),
('es', 'DISEASES',  'Enfermedades',        FALSE, 5),
('es', 'EQUIPMENT', 'Equipamiento',        FALSE, 6),
('es', 'NEWS',      'Noticias',            FALSE, 7),
-- 포르투갈어
('pt', 'GENERAL',   'Geral',               FALSE, 1),
('pt', 'QUESTION',  'Perguntas',           FALSE, 2),
('pt', 'SHOWCASE',  'Aquário Showcase',    FALSE, 3),
('pt', 'BREEDING',  'Diário de Criação',   FALSE, 4),
('pt', 'DISEASES',  'Doenças',             FALSE, 5),
('pt', 'EQUIPMENT', 'Equipamentos',        FALSE, 6),
('pt', 'NEWS',      'Notícias',            FALSE, 7),
-- 아랍어 (RTL)
('ar', 'GENERAL',   'عام',          TRUE, 1),
('ar', 'QUESTION',  'أسئلة وأجوبة', TRUE, 2),
('ar', 'SHOWCASE',  'عرض الأحواض',  TRUE, 3),
('ar', 'BREEDING',  'مذكرات التكاثر', TRUE, 4),
('ar', 'DISEASES',  'الأمراض والعلاج', TRUE, 5),
('ar', 'EQUIPMENT', 'المعدات',      TRUE, 6),
('ar', 'NEWS',      'أخبار وإعلانات', TRUE, 7),
-- 히브리어 (RTL)
('he', 'GENERAL',   'כללי',              TRUE, 1),
('he', 'QUESTION',  'שאלות ותשובות',     TRUE, 2),
('he', 'SHOWCASE',  'תצוגת אקווריום',    TRUE, 3),
('he', 'BREEDING',  'יומן רביה',         TRUE, 4),
('he', 'DISEASES',  'מחלות וטיפול',      TRUE, 5),
('he', 'EQUIPMENT', 'ציוד וחומרים',      TRUE, 6),
('he', 'NEWS',      'חדשות והכרזות',     TRUE, 7);

CREATE TABLE posts (
    id            BIGSERIAL PRIMARY KEY,
    board_id      BIGINT NOT NULL REFERENCES boards(id),
    author_id     UUID NOT NULL REFERENCES users(id),
    locale        VARCHAR(10) NOT NULL,   -- 게시판 locale과 반드시 일치
    title         VARCHAR(200) NOT NULL,
    content       TEXT NOT NULL,
    content_type  VARCHAR(20) NOT NULL DEFAULT 'MARKDOWN',
    image_urls    JSONB NOT NULL DEFAULT '[]',
    fish_data_id  BIGINT REFERENCES fish_data(id),
    view_count    INT NOT NULL DEFAULT 0,
    like_count    INT NOT NULL DEFAULT 0,
    comment_count INT NOT NULL DEFAULT 0,
    is_pinned     BOOLEAN NOT NULL DEFAULT FALSE,
    is_locked     BOOLEAN NOT NULL DEFAULT FALSE,
    is_deleted    BOOLEAN NOT NULL DEFAULT FALSE,
    created_at    TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at    TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at    TIMESTAMPTZ,
    -- 게시판-로케일 일관성 보장
    CONSTRAINT chk_post_locale_match CHECK (
        locale = (SELECT locale FROM boards WHERE id = board_id)
    )
);

CREATE INDEX idx_posts_board ON posts (board_id, is_deleted, created_at DESC);
CREATE INDEX idx_posts_author ON posts (author_id);
CREATE INDEX idx_posts_fish ON posts (fish_data_id);
CREATE INDEX idx_posts_fts ON posts
    USING GIN(to_tsvector('simple', title || ' ' || left(content, 500)));

CREATE TABLE comments (
    id          BIGSERIAL PRIMARY KEY,
    post_id     BIGINT NOT NULL REFERENCES posts(id) ON DELETE CASCADE,
    author_id   UUID NOT NULL REFERENCES users(id),
    parent_id   BIGINT REFERENCES comments(id),
    content     TEXT NOT NULL,
    like_count  INT NOT NULL DEFAULT 0,
    is_deleted  BOOLEAN NOT NULL DEFAULT FALSE,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_comments_post ON comments (post_id, is_deleted);
CREATE INDEX idx_comments_parent ON comments (parent_id);

CREATE TABLE post_likes (
    post_id    BIGINT NOT NULL REFERENCES posts(id) ON DELETE CASCADE,
    user_id    UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    PRIMARY KEY (post_id, user_id)
);

-- ============================================================
-- 수조 관리
-- ============================================================

CREATE TABLE tanks (
    id                 BIGSERIAL PRIMARY KEY,
    owner_id           UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    name               VARCHAR(100) NOT NULL,
    size_liters        INT NOT NULL,
    setup_date         DATE,
    description        TEXT,
    image_url          TEXT,
    current_ph         DECIMAL(3,1),
    current_temp_c     DECIMAL(4,1),
    current_nh3        DECIMAL(8,4),
    current_no2        DECIMAL(8,4),
    current_no3        DECIMAL(8,4),
    last_water_change  DATE,
    is_public          BOOLEAN NOT NULL DEFAULT TRUE,
    created_at         TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at         TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_tanks_owner ON tanks (owner_id);

CREATE TABLE tank_inhabitants (
    id           BIGSERIAL PRIMARY KEY,
    tank_id      BIGINT NOT NULL REFERENCES tanks(id) ON DELETE CASCADE,
    fish_data_id BIGINT REFERENCES fish_data(id),
    custom_name  VARCHAR(255) NOT NULL DEFAULT '',
    quantity     INT NOT NULL DEFAULT 1,
    added_at     TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    removed_at   TIMESTAMPTZ,
    notes        TEXT
);

CREATE INDEX idx_tank_inhabitants_tank ON tank_inhabitants (tank_id, removed_at);
