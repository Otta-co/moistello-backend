CREATE TABLE reputation_snapshots (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    score INTEGER NOT NULL DEFAULT 0,
    level VARCHAR(20) NOT NULL DEFAULT 'Bronze',
    breakdown JSONB NOT NULL DEFAULT '{}',
    month DATE NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE(user_id, month)
);

CREATE INDEX idx_reputation_user ON reputation_snapshots(user_id);
CREATE INDEX idx_reputation_month ON reputation_snapshots(month);
