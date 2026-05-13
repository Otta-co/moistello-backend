CREATE TABLE indexer_cursor (
    id SERIAL PRIMARY KEY,
    chain VARCHAR(20) NOT NULL DEFAULT 'stellar',
    last_ledger BIGINT NOT NULL DEFAULT 0,
    last_processed_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

INSERT INTO indexer_cursor (chain, last_ledger) VALUES ('stellar', 0);
