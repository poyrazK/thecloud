-- +goose Up

CREATE TABLE IF NOT EXISTS users (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    email TEXT NOT NULL UNIQUE,
    password_hash TEXT NOT NULL,
    name TEXT,
    role TEXT NOT NULL DEFAULT 'user',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_users_email ON users(email);

-- Add foreign key to api_keys
-- Note: In a real system we might need to handle existing user_ids in api_keys
-- For now we assume we start fresh or the migration handles it
ALTER TABLE api_keys 
    ADD CONSTRAINT fk_api_keys_user 
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE RESTRICT;
