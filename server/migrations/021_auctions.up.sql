CREATE TABLE IF NOT EXISTS auctions (
    id              BIGSERIAL PRIMARY KEY,
    listing_id      BIGINT REFERENCES listings(id) ON DELETE SET NULL,
    seller_id       UUID NOT NULL REFERENCES users(id),
    title           TEXT NOT NULL,
    description     TEXT,
    image_url       TEXT,
    start_price     BIGINT NOT NULL,
    current_price   BIGINT,
    reserve_price   BIGINT,
    bid_increment   BIGINT NOT NULL DEFAULT 1000,
    starts_at       TIMESTAMPTZ NOT NULL,
    ends_at         TIMESTAMPTZ NOT NULL,
    status          VARCHAR(20) NOT NULL DEFAULT 'scheduled',
    -- scheduled, live, ended, cancelled
    winner_id       UUID REFERENCES users(id),
    final_price     BIGINT,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS auction_bids (
    id          BIGSERIAL PRIMARY KEY,
    auction_id  BIGINT NOT NULL REFERENCES auctions(id) ON DELETE CASCADE,
    bidder_id   UUID NOT NULL REFERENCES users(id),
    amount      BIGINT NOT NULL,
    is_winning  BOOLEAN NOT NULL DEFAULT FALSE,
    bid_at      TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_auctions_status ON auctions(status, ends_at);
CREATE INDEX IF NOT EXISTS idx_auction_bids_auction ON auction_bids(auction_id, bid_at DESC);
