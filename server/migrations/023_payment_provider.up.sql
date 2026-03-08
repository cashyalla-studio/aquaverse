-- escrow_transactions에 다중 PSP provider 컬럼 추가
ALTER TABLE escrow_transactions
    ADD COLUMN IF NOT EXISTS provider VARCHAR(30) DEFAULT 'toss';

-- 기존 pg_provider 값을 provider 컬럼으로 복사 (존재하는 경우)
UPDATE escrow_transactions
    SET provider = pg_provider
    WHERE pg_provider IS NOT NULL AND provider IS NULL;

-- 결제 로그 테이블 (다중 PSP 공통)
CREATE TABLE IF NOT EXISTS payment_logs (
    id              BIGSERIAL PRIMARY KEY,
    trade_id        BIGINT REFERENCES trades(id),
    provider        VARCHAR(30) NOT NULL,           -- 'toss', 'stripe'
    provider_txn_id TEXT,
    amount          BIGINT NOT NULL,
    currency        VARCHAR(10) NOT NULL DEFAULT 'KRW',
    status          VARCHAR(20) NOT NULL,            -- 'initiated', 'confirmed', 'failed', 'refunded'
    payload         JSONB,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_payment_logs_trade    ON payment_logs(trade_id);
CREATE INDEX IF NOT EXISTS idx_payment_logs_provider ON payment_logs(provider, created_at DESC);
CREATE INDEX IF NOT EXISTS idx_payment_logs_txn      ON payment_logs(provider_txn_id);
