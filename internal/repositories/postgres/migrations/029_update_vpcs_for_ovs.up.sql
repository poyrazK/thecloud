-- +goose Up

ALTER TABLE vpcs ADD COLUMN IF NOT EXISTS cidr_block VARCHAR(18);
ALTER TABLE vpcs ADD COLUMN IF NOT EXISTS vxlan_id INT;
ALTER TABLE vpcs ADD COLUMN IF NOT EXISTS status VARCHAR(50) DEFAULT 'active';
ALTER TABLE vpcs ADD COLUMN IF NOT EXISTS arn VARCHAR(512);

-- Set defaults for existing rows
UPDATE vpcs SET cidr_block = '10.0.0.0/16', vxlan_id = 100, status = 'active' WHERE cidr_block IS NULL;
