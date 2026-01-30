-- +goose Down
DROP TABLE IF EXISTS instance_types;
ALTER TABLE instances DROP COLUMN IF EXISTS instance_type;
