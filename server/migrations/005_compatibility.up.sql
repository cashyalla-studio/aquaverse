-- 수조 테이블 (기존 없음, 신규 생성)
CREATE TABLE IF NOT EXISTS tanks (
    id          BIGSERIAL PRIMARY KEY,
    user_id     UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    name        VARCHAR(100) NOT NULL,
    volume_l    INT,            -- 수조 용량 (리터)
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE INDEX idx_tanks_user ON tanks(user_id);

-- 수조 입주 어종
CREATE TABLE IF NOT EXISTS tank_inhabitants (
    tank_id      BIGINT NOT NULL REFERENCES tanks(id) ON DELETE CASCADE,
    fish_data_id BIGINT NOT NULL REFERENCES fish_data(id),
    quantity     INT NOT NULL DEFAULT 1,
    added_at     TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    PRIMARY KEY (tank_id, fish_data_id)
);

-- 어종 호환성 매트릭스
CREATE TABLE IF NOT EXISTS fish_compatibility (
    fish_a_id    BIGINT NOT NULL REFERENCES fish_data(id),
    fish_b_id    BIGINT NOT NULL REFERENCES fish_data(id),
    compatible   BOOLEAN NOT NULL,
    caution      BOOLEAN NOT NULL DEFAULT FALSE,
    reason       TEXT,
    PRIMARY KEY (fish_a_id, fish_b_id),
    CHECK (fish_a_id < fish_b_id)
);

-- 샘플 데이터: 잘 알려진 호환성 관계 (Rule-based seed)
-- 구피(1)와 네온테트라(2) 호환 예시 (실제 fish_data id에 맞게 조정 필요)
-- INSERT는 id가 확정되지 않으므로 주석 처리 후 별도 seed 파일로 관리
-- INSERT INTO fish_compatibility ...
