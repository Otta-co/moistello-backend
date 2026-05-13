CREATE TYPE penalty_type AS ENUM ('late', 'default', 'early_exit');

CREATE TABLE penalties (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    circle_id UUID NOT NULL REFERENCES circles(id) ON DELETE CASCADE,
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    round_number INTEGER NOT NULL,
    penalty_type penalty_type NOT NULL,
    amount NUMERIC(18,7) NOT NULL,
    strikes_applied INTEGER NOT NULL DEFAULT 0,
    reason TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_penalties_circle ON penalties(circle_id);
CREATE INDEX idx_penalties_user ON penalties(user_id);
