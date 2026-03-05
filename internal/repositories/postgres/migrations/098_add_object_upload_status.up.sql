-- +goose Up
ALTER TABLE objects ADD COLUMN IF NOT EXISTS upload_status VARCHAR(50) DEFAULT 'AVAILABLE';

-- Set default for existing records and enforce constraints
UPDATE objects SET upload_status = 'AVAILABLE' WHERE upload_status IS NULL;
ALTER TABLE objects ALTER COLUMN upload_status SET NOT NULL;
ALTER TABLE objects ADD CONSTRAINT ck_objects_upload_status_allowed_values CHECK (upload_status IN ('PENDING', 'AVAILABLE'));

-- Optimized index for finding pending uploads
CREATE INDEX IF NOT EXISTS idx_objects_pending ON objects(created_at) WHERE upload_status = 'PENDING';
