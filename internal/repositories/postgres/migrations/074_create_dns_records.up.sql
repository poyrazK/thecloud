CREATE TABLE IF NOT EXISTS dns_records (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    zone_id UUID NOT NULL REFERENCES dns_zones(id) ON DELETE CASCADE,
    name VARCHAR(255) NOT NULL,
    type VARCHAR(10) NOT NULL,
    content TEXT NOT NULL,
    ttl INTEGER NOT NULL DEFAULT 300,
    priority INTEGER,
    disabled BOOLEAN NOT NULL DEFAULT FALSE,
    auto_managed BOOLEAN NOT NULL DEFAULT FALSE,
    instance_id UUID REFERENCES instances(id) ON DELETE SET NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_dns_records_zone ON dns_records(zone_id);
CREATE INDEX IF NOT EXISTS idx_dns_records_instance ON dns_records(instance_id) WHERE instance_id IS NOT NULL;
CREATE INDEX IF NOT EXISTS idx_dns_records_lookup ON dns_records(zone_id, name, type);
