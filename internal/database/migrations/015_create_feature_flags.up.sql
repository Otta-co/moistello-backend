CREATE TABLE feature_flags (
    id SERIAL PRIMARY KEY,
    flag VARCHAR(100) UNIQUE NOT NULL,
    enabled BOOLEAN NOT NULL DEFAULT FALSE,
    description TEXT,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

INSERT INTO feature_flags (flag, enabled, description) VALUES ('auction_payouts', false, 'Enable auction-based payout type');
INSERT INTO feature_flags (flag, enabled, description) VALUES ('vote_payouts', false, 'Enable vote-based payout type');
INSERT INTO feature_flags (flag, enabled, description) VALUES ('premium_circles', false, 'Enable premium circle type');
INSERT INTO feature_flags (flag, enabled, description) VALUES ('kyc_required', false, 'Require KYC for withdrawals');
