-- +goose Up
DO $$ 
BEGIN
    IF EXISTS (SELECT 1 FROM information_schema.columns WHERE table_schema = current_schema() AND table_name='clusters' AND column_name='ssh_key') THEN
        ALTER TABLE clusters RENAME COLUMN ssh_key TO ssh_key_id;
    END IF;

    IF EXISTS (SELECT 1 FROM information_schema.columns WHERE table_schema = current_schema() AND table_name='clusters' AND column_name='kubeconfig') THEN
        ALTER TABLE clusters DROP COLUMN kubeconfig;
    END IF;
END $$;
