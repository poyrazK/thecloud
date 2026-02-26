-- +goose Down

DROP INDEX IF EXISTS idx_pipeline_webhook_deliveries_pipeline_id;
DROP TABLE IF EXISTS pipeline_webhook_deliveries;
