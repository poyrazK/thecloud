-- +goose Up

-- CloudQueue: Message Queue Service
CREATE TABLE IF NOT EXISTS queues (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    name VARCHAR(255) NOT NULL,
    arn VARCHAR(512) NOT NULL,
    visibility_timeout INT NOT NULL DEFAULT 30,
    retention_days INT NOT NULL DEFAULT 4,
    max_message_size INT NOT NULL DEFAULT 262144,
    status VARCHAR(50) NOT NULL DEFAULT 'ACTIVE',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE(user_id, name)
);

CREATE INDEX IF NOT EXISTS idx_queues_user_id ON queues(user_id);

CREATE TABLE IF NOT EXISTS queue_messages (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    queue_id UUID NOT NULL REFERENCES queues(id) ON DELETE CASCADE,
    body TEXT NOT NULL,
    receipt_handle VARCHAR(255),
    visible_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    received_count INT NOT NULL DEFAULT 0,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Index for efficient polling
CREATE INDEX IF NOT EXISTS idx_messages_queue_visible ON queue_messages(queue_id, visible_at);
