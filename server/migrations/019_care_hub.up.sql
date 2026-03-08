-- 케어 일정 (먹이, 수질, 청소 등)
CREATE TABLE IF NOT EXISTS care_schedules (
    id           BIGSERIAL PRIMARY KEY,
    tank_id      BIGINT NOT NULL REFERENCES tanks(id) ON DELETE CASCADE,
    user_id      UUID NOT NULL REFERENCES users(id),
    schedule_type VARCHAR(50) NOT NULL, -- 'feeding', 'water_change', 'filter_clean', 'medication', 'custom'
    title        TEXT NOT NULL,
    description  TEXT,
    frequency    VARCHAR(50) NOT NULL, -- 'daily', 'weekly', 'biweekly', 'monthly', 'custom'
    interval_days INT,
    next_due_at  TIMESTAMPTZ NOT NULL,
    last_done_at TIMESTAMPTZ,
    is_active    BOOLEAN NOT NULL DEFAULT TRUE,
    created_at   TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- 케어 기록 로그
CREATE TABLE IF NOT EXISTS care_logs (
    id              BIGSERIAL PRIMARY KEY,
    schedule_id     BIGINT REFERENCES care_schedules(id) ON DELETE SET NULL,
    tank_id         BIGINT NOT NULL REFERENCES tanks(id) ON DELETE CASCADE,
    user_id         UUID NOT NULL REFERENCES users(id),
    care_type       VARCHAR(50) NOT NULL,
    notes           TEXT,
    photo_url       TEXT,
    done_at         TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- 케어 스트릭 (연속 기록)
CREATE TABLE IF NOT EXISTS care_streaks (
    user_id         UUID PRIMARY KEY REFERENCES users(id),
    current_streak  INT NOT NULL DEFAULT 0,
    longest_streak  INT NOT NULL DEFAULT 0,
    last_care_date  DATE,
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_care_schedules_tank ON care_schedules(tank_id);
CREATE INDEX IF NOT EXISTS idx_care_schedules_user_due ON care_schedules(user_id, next_due_at);
CREATE INDEX IF NOT EXISTS idx_care_logs_user ON care_logs(user_id, done_at DESC);
