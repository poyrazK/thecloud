ALTER TABLE instances ADD COLUMN ssh_key_id UUID REFERENCES ssh_keys(id);
