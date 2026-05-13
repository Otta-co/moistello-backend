CREATE TYPE member_status AS ENUM ('pending', 'active', 'completed', 'defaulted', 'exited');

CREATE TABLE circle_members (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    circle_id UUID NOT NULL REFERENCES circles(id) ON DELETE CASCADE,
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    position INTEGER NOT NULL,
    status member_status NOT NULL DEFAULT 'pending',
    joined_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE(circle_id, user_id)
);

CREATE INDEX idx_cm_circle ON circle_members(circle_id);
CREATE INDEX idx_cm_user ON circle_members(user_id);
CREATE INDEX idx_cm_status ON circle_members(status);
