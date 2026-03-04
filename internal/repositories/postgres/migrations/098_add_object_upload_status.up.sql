-- +goose Up
ALTER TABLE objects ADD COLUMN IF NOT EXISTS upload_status VARCHAR(50) DEFAULT 'AVAILABLE';

-- Set default for existing records
UPDATE objects SET upload_status = 'AVAILABLE' WHERE upload_status IS NULL;

-- Index for finding pending uploads efficiently
CREATE INDEX IF NOT EXISTS idx_objects_pending ON objects(upload_status, created_at) WHERE upload_status = 'PENDING';
