CREATE TYPE contribution_status AS ENUM ('pending', 'confirmed', 'failed');

CREATE TABLE contributions (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    circle_id UUID NOT NULL REFERENCES circles(id) ON DELETE CASCADE,
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    round_number INTEGER NOT NULL,
    amount NUMERIC(18,7) NOT NULL,
    txn_hash VARCHAR(64),
    status contribution_status NOT NULL DEFAULT 'pending',
    on_time BOOLEAN NOT NULL DEFAULT TRUE,
    submitted_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE(circle_id, user_id, round_number)
);

CREATE INDEX idx_contrib_circle ON contributions(circle_id);
CREATE INDEX idx_contrib_user ON contributions(user_id);
CREATE INDEX idx_contrib_txn ON contributions(txn_hash);
