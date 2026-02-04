-- +goose Up
ALTER TABLE load_balancers ADD COLUMN IF NOT EXISTS ip VARCHAR(50);

-- +goose Down
ALTER TABLE load_balancers DROP COLUMN IF EXISTS ip;
