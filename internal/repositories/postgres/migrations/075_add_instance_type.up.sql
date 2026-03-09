-- +goose Up
-- Add instance_type column with default for backward compatibility
ALTER TABLE instances ADD COLUMN IF NOT EXISTS instance_type TEXT DEFAULT 'basic-2';

-- Create instance_types reference table
CREATE TABLE IF NOT EXISTS instance_types (
    id TEXT PRIMARY KEY,
    name TEXT NOT NULL,
    vcpus INT NOT NULL,
    memory_mb INT NOT NULL,
    disk_gb INT NOT NULL,
    network_mbps INT NOT NULL DEFAULT 1000,
    price_per_hour NUMERIC(10,4) NOT NULL DEFAULT 0.01,
    category TEXT NOT NULL DEFAULT 'general',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Seed default instance types
INSERT INTO instance_types (id, name, vcpus, memory_mb, disk_gb, category, price_per_hour) VALUES
    ('basic-1',       'Basic 1',       1,   512,  8,  'basic',       0.005),
    ('basic-2',       'Basic 2',       1,  1024, 10,  'basic',       0.01),
    ('standard-1',    'Standard 1',    2,  2048, 20,  'standard',    0.02),
    ('standard-2',    'Standard 2',    2,  4096, 40,  'standard',    0.04),
    ('standard-4',    'Standard 4',    4,  8192, 80,  'standard',    0.08),
    ('performance-1', 'Performance 1', 4, 16384, 160, 'performance', 0.16),
    ('performance-2', 'Performance 2', 8, 32768, 320, 'performance', 0.32)
ON CONFLICT (id) DO NOTHING;
