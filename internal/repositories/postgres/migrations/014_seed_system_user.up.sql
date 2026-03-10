-- +goose Up
-- Seed system user for background tasks
-- Using bare ON CONFLICT DO NOTHING to handle both ID and EMAIL constraints
INSERT INTO users (id, email, password_hash, name, role)
VALUES ('00000000-0000-0000-0000-000000000001', 'system@thecloud.local', 'SYSTEM_ACCOUNT_NO_LOGIN', 'System User', 'admin')
ON CONFLICT DO NOTHING;
