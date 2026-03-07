-- admin_audit_logs: 관리자 행동 기록
CREATE TABLE IF NOT EXISTS admin_audit_logs (
    id          BIGSERIAL PRIMARY KEY,
    admin_id    UUID NOT NULL REFERENCES users(id),
    action      VARCHAR(100) NOT NULL,  -- 'BAN_USER', 'UNBAN_USER', 'DELETE_LISTING', etc.
    target_type VARCHAR(50),            -- 'USER', 'LISTING', 'TRADE'
    target_id   VARCHAR(100),           -- 대상 ID (UUID or BIGINT as string)
    detail      JSONB,                  -- 추가 정보
    ip_address  INET,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE INDEX idx_admin_audit_logs_admin ON admin_audit_logs(admin_id, created_at DESC);
CREATE INDEX idx_admin_audit_logs_action ON admin_audit_logs(action, created_at DESC);
