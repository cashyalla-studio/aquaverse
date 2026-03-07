-- 에스크로 테이블에 PG 연동 컬럼 추가
ALTER TABLE escrow_transactions
    ADD COLUMN IF NOT EXISTS pg_transaction_id VARCHAR(100),
    ADD COLUMN IF NOT EXISTS pg_provider      VARCHAR(20) DEFAULT 'toss',
    ADD COLUMN IF NOT EXISTS pg_payment_key   VARCHAR(200),
    ADD COLUMN IF NOT EXISTS pg_order_id      VARCHAR(100),
    ADD COLUMN IF NOT EXISTS checkout_url     TEXT,
    ADD COLUMN IF NOT EXISTS pg_raw_response  JSONB;

CREATE INDEX IF NOT EXISTS idx_escrow_pg_order ON escrow_transactions(pg_order_id);
