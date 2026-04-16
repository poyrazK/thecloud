-- +goose Down
ALTER TABLE IF EXISTS volume_encryption_keys DROP CONSTRAINT IF EXISTS fk_volume_encryption_keys_volume;
DROP TABLE IF EXISTS volume_encryption_keys;
