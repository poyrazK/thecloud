-- +goose Up

CREATE TABLE IF NOT EXISTS snapshots (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id),
    volume_id UUID NOT NULL REFERENCES volumes(id) ON DELETE CASCADE,
    volume_name VARCHAR(255) NOT NULL,
    size_gb INT NOT NULL,
    status VARCHAR(50) NOT NULL DEFAULT 'CREATING',
    description TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_snapshots_user_id ON snapshots(user_id);
CREATE INDEX IF NOT EXISTS idx_snapshots_volume_id ON snapshots(volume_id);
