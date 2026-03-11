-- +goose Down
ALTER TABLE objects DROP COLUMN IF EXISTS checksum;
