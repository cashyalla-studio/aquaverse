-- 수조 수질 기록
CREATE TABLE tank_water_params (
    id          BIGSERIAL PRIMARY KEY,
    tank_id     BIGINT NOT NULL REFERENCES tanks(id) ON DELETE CASCADE,
    temp_c      NUMERIC(4,1),
    ph          NUMERIC(3,1),
    ammonia_ppm NUMERIC(5,2),
    nitrite_ppm NUMERIC(5,2),
    nitrate_ppm NUMERIC(5,2),
    gh_dgh      NUMERIC(4,1),
    kh_dkh      NUMERIC(4,1),
    recorded_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    notes       TEXT
);
CREATE INDEX idx_tank_water_params ON tank_water_params(tank_id, recorded_at DESC);

-- AI 진단 결과 캐시
CREATE TABLE tank_diagnosis_cache (
    tank_id       BIGINT PRIMARY KEY REFERENCES tanks(id) ON DELETE CASCADE,
    diagnosis     JSONB NOT NULL,
    param_hash    VARCHAR(64),
    created_at    TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
