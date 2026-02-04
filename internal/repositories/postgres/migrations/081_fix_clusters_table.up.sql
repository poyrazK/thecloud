-- Add missing columns to clusters table
ALTER TABLE clusters ADD COLUMN IF NOT EXISTS pod_cidr VARCHAR(50) DEFAULT '10.244.0.0/16';
ALTER TABLE clusters ADD COLUMN IF NOT EXISTS service_cidr VARCHAR(50) DEFAULT '10.96.0.0/12';
ALTER TABLE clusters ADD COLUMN IF NOT EXISTS ssh_private_key_encrypted TEXT;
ALTER TABLE clusters ADD COLUMN IF NOT EXISTS kubeconfig_encrypted TEXT;
ALTER TABLE clusters ADD COLUMN IF NOT EXISTS join_token TEXT;
ALTER TABLE clusters ADD COLUMN IF NOT EXISTS token_expires_at TIMESTAMPTZ;
ALTER TABLE clusters ADD COLUMN IF NOT EXISTS ca_cert_hash TEXT;

-- Move data if existing columns exist
DO $$ 
BEGIN
    IF EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name='clusters' AND column_name='ssh_key') THEN
        UPDATE clusters SET ssh_private_key_encrypted = ssh_key WHERE ssh_private_key_encrypted IS NULL;
    END IF;
    IF EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name='clusters' AND column_name='kubeconfig') THEN
        UPDATE clusters SET kubeconfig_encrypted = kubeconfig WHERE kubeconfig_encrypted IS NULL;
    END IF;
END $$;
