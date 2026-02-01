CREATE TABLE IF NOT EXISTS usage_records (
    id UUID PRIMARY KEY,
    user_id UUID NOT NULL,
    resource_id UUID NOT NULL,
    resource_type VARCHAR(50) NOT NULL,
    quantity DECIMAL(18, 6) NOT NULL,
    unit VARCHAR(20) NOT NULL,
    start_time TIMESTAMP WITH TIME ZONE NOT NULL,
    end_time TIMESTAMP WITH TIME ZONE NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_usage_user_period ON usage_records(user_id, start_time, end_time);
CREATE INDEX IF NOT EXISTS idx_usage_resource ON usage_records(resource_id);
