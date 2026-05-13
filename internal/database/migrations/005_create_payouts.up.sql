CREATE TABLE payouts (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    circle_id UUID NOT NULL REFERENCES circles(id) ON DELETE CASCADE,
    recipient_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    round_number INTEGER NOT NULL,
    amount NUMERIC(18,7) NOT NULL,
    fee_amount NUMERIC(18,7) NOT NULL DEFAULT 0,
    txn_hash VARCHAR(64),
    payout_type VARCHAR(20) NOT NULL,
    executed_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_payouts_circle ON payouts(circle_id);
CREATE INDEX idx_payouts_recipient ON payouts(recipient_id);
