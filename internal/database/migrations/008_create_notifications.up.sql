CREATE TYPE notification_type AS ENUM ('circle_created', 'member_joined', 'contribution_due', 'contribution_received', 'contribution_late', 'payout_received', 'circle_completed', 'member_exited', 'dispute_raised', 'system');
CREATE TYPE notification_channel AS ENUM ('email', 'sms', 'push', 'inapp');

CREATE TABLE notifications (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    type notification_type NOT NULL,
    title VARCHAR(200) NOT NULL,
    body TEXT,
    data JSONB,
    is_read BOOLEAN NOT NULL DEFAULT FALSE,
    channel notification_channel NOT NULL DEFAULT 'inapp',
    sent_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_notifs_user ON notifications(user_id, is_read);
CREATE INDEX idx_notifs_created ON notifications(created_at DESC);
