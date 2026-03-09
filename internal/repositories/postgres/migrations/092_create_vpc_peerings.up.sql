CREATE TABLE IF NOT EXISTS vpc_peerings (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    requester_vpc_id UUID NOT NULL REFERENCES vpcs(id),
    accepter_vpc_id UUID NOT NULL REFERENCES vpcs(id),
    tenant_id UUID NOT NULL,
    status VARCHAR(32) NOT NULL DEFAULT 'pending-acceptance',
    arn VARCHAR(256),
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW(),
    CONSTRAINT no_self_peering CHECK (requester_vpc_id != accepter_vpc_id)
);

-- Prevent duplicate active/pending peerings between the same pair of VPCs (order-insensitive).
CREATE UNIQUE INDEX idx_active_peering_pair
    ON vpc_peerings (LEAST(requester_vpc_id, accepter_vpc_id), GREATEST(requester_vpc_id, accepter_vpc_id))
    WHERE status IN ('pending-acceptance', 'active');
