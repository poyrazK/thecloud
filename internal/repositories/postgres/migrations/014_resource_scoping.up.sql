-- +goose Up

-- Add user_id to all resource tables
ALTER TABLE instances ADD COLUMN user_id UUID REFERENCES users(id) ON DELETE RESTRICT;
ALTER TABLE vpcs ADD COLUMN user_id UUID REFERENCES users(id) ON DELETE RESTRICT;
ALTER TABLE volumes ADD COLUMN user_id UUID REFERENCES users(id) ON DELETE RESTRICT;
ALTER TABLE objects ADD COLUMN user_id UUID REFERENCES users(id) ON DELETE RESTRICT;
ALTER TABLE load_balancers ADD COLUMN user_id UUID REFERENCES users(id) ON DELETE RESTRICT;
ALTER TABLE scaling_groups ADD COLUMN user_id UUID REFERENCES users(id) ON DELETE RESTRICT;
ALTER TABLE events ADD COLUMN user_id UUID REFERENCES users(id) ON DELETE RESTRICT;

-- Create indexes for performance
CREATE INDEX idx_instances_user_id ON instances(user_id);
CREATE INDEX idx_vpcs_user_id ON vpcs(user_id);
CREATE INDEX idx_volumes_user_id ON volumes(user_id);
CREATE INDEX idx_objects_user_id ON objects(user_id);
CREATE INDEX idx_load_balancers_user_id ON load_balancers(user_id);
CREATE INDEX idx_scaling_groups_user_id ON scaling_groups(user_id);
CREATE INDEX idx_events_user_id ON events(user_id);
