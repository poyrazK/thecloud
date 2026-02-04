-- +goose Down
ALTER TABLE load_balancers DROP COLUMN IF EXISTS ip;
