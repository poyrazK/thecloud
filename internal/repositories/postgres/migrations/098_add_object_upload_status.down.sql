-- +goose Down
DROP INDEX IF EXISTS idx_objects_pending;
ALTER TABLE objects DROP COLUMN IF EXISTS upload_status;
