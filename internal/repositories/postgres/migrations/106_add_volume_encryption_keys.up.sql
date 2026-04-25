-- +goose Up
CREATE TABLE IF NOT EXISTS volume_encryption_keys (
    volume_id      UUID PRIMARY KEY,
    encrypted_dek  BYTEA NOT NULL,
    kms_key_id     VARCHAR(500) NOT NULL,
    algorithm      VARCHAR(50) NOT NULL DEFAULT 'AES-256-GCM',
    created_at     TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Drop FK if exists (for idempotent re-runs)
ALTER TABLE volume_encryption_keys DROP CONSTRAINT IF EXISTS fk_volume_encryption_keys_volume;
ALTER TABLE volume_encryption_keys ADD CONSTRAINT fk_volume_encryption_keys_volume
    FOREIGN KEY (volume_id) REFERENCES volumes(id) ON DELETE CASCADE;

CREATE INDEX IF NOT EXISTS idx_volume_encryption_keys_kms_key_id ON volume_encryption_keys(kms_key_id);

-- +goose Down
DROP TABLE IF EXISTS volume_encryption_keys;
