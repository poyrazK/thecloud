ALTER TABLE clusters DROP COLUMN IF EXISTS pod_cidr;
ALTER TABLE clusters DROP COLUMN IF EXISTS service_cidr;
ALTER TABLE clusters DROP COLUMN IF EXISTS ssh_private_key_encrypted;
ALTER TABLE clusters DROP COLUMN IF EXISTS kubeconfig_encrypted;
ALTER TABLE clusters DROP COLUMN IF EXISTS join_token;
ALTER TABLE clusters DROP COLUMN IF EXISTS token_expires_at;
ALTER TABLE clusters DROP COLUMN IF EXISTS ca_cert_hash;
