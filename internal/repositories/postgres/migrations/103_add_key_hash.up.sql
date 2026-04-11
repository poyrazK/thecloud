-- +goose Up

ALTER TABLE api_keys ADD COLUMN IF NOT EXISTS key_hash TEXT NOT NULL DEFAULT '';
CREATE UNIQUE INDEX IF NOT EXISTS idx_api_keys_key_hash ON api_keys(key_hash) WHERE key_hash != '';

-- Backfill existing keys: hash the plaintext and store it.
-- This runs once; future writes set key_hash via application code.
UPDATE api_keys SET key_hash = encode(sha256(key::bytea), 'hex') WHERE key_hash = '';

-- +goose Down

DROP INDEX IF EXISTS idx_api_keys_key_hash;
ALTER TABLE api_keys DROP COLUMN IF EXISTS key_hash;
