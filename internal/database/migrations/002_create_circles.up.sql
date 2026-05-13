CREATE TYPE circle_type AS ENUM ('public', 'private', 'org', 'community', 'premium');
CREATE TYPE payout_type AS ENUM ('random', 'fixed', 'auction', 'vote');
CREATE TYPE circle_frequency AS ENUM ('daily', 'weekly', 'biweekly', 'monthly');
CREATE TYPE circle_currency AS ENUM ('USDC', 'XLM');
CREATE TYPE circle_status AS ENUM ('pending', 'active', 'completed', 'cancelled', 'disputed');

CREATE TABLE circles (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    contract_id VARCHAR(64) UNIQUE,
    name VARCHAR(200) NOT NULL,
    description TEXT,
    circle_type circle_type NOT NULL DEFAULT 'public',
    payout_type payout_type NOT NULL DEFAULT 'random',
    contribution_amount NUMERIC(18,7) NOT NULL,
    currency circle_currency NOT NULL DEFAULT 'USDC',
    frequency circle_frequency NOT NULL DEFAULT 'monthly',
    max_members INTEGER NOT NULL CHECK (max_members >= 2 AND max_members <= 100),
    min_moi_score INTEGER NOT NULL DEFAULT 0,
    collateral_percent NUMERIC(5,2) NOT NULL DEFAULT 0,
    late_fee_percent NUMERIC(5,2) NOT NULL DEFAULT 5,
    grace_period_hours INTEGER NOT NULL DEFAULT 24,
    max_strikes INTEGER NOT NULL DEFAULT 3,
    start_date TIMESTAMPTZ,
    end_date TIMESTAMPTZ,
    status circle_status NOT NULL DEFAULT 'pending',
    current_round INTEGER NOT NULL DEFAULT 0,
    total_contributions NUMERIC(18,7) NOT NULL DEFAULT 0,
    organizer_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    metadata_ipfs_hash VARCHAR(64),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_circles_status ON circles(status);
CREATE INDEX idx_circles_type ON circles(circle_type);
CREATE INDEX idx_circles_organizer ON circles(organizer_id);
CREATE INDEX idx_circles_search ON circles USING GIN(to_tsvector('english', name || ' ' || COALESCE(description, '')));

CREATE TRIGGER trg_circles_updated_at BEFORE UPDATE ON circles FOR EACH ROW EXECUTE FUNCTION update_updated_at();
