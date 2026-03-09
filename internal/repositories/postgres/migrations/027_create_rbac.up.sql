-- +goose Up

CREATE TABLE IF NOT EXISTS roles (
    id UUID PRIMARY KEY,
    name TEXT UNIQUE NOT NULL,
    description TEXT,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS role_permissions (
    role_id UUID REFERENCES roles(id) ON DELETE CASCADE,
    permission TEXT NOT NULL,
    PRIMARY KEY (role_id, permission)
);

-- Insert default roles
-- Admin: *
INSERT INTO roles (id, name, description) VALUES ('00000000-0000-0000-0000-000000000001', 'admin', 'Full platform access') ON CONFLICT DO NOTHING;
INSERT INTO role_permissions (role_id, permission) VALUES ('00000000-0000-0000-0000-000000000001', '*') ON CONFLICT DO NOTHING;

-- Developer: basic resource management
INSERT INTO roles (id, name, description) VALUES ('00000000-0000-0000-0000-000000000002', 'developer', 'Access to manage compute, storage and network resources') ON CONFLICT DO NOTHING;
INSERT INTO role_permissions (role_id, permission) VALUES ('00000000-0000-0000-0000-000000000002', 'instance:launch') ON CONFLICT DO NOTHING;
INSERT INTO role_permissions (role_id, permission) VALUES ('00000000-0000-0000-0000-000000000002', 'instance:terminate') ON CONFLICT DO NOTHING;
INSERT INTO role_permissions (role_id, permission) VALUES ('00000000-0000-0000-0000-000000000002', 'instance:read') ON CONFLICT DO NOTHING;
INSERT INTO role_permissions (role_id, permission) VALUES ('00000000-0000-0000-0000-000000000002', 'vpc:read') ON CONFLICT DO NOTHING;
INSERT INTO role_permissions (role_id, permission) VALUES ('00000000-0000-0000-0000-000000000002', 'volume:read') ON CONFLICT DO NOTHING;
INSERT INTO role_permissions (role_id, permission) VALUES ('00000000-0000-0000-0000-000000000002', 'snapshot:read') ON CONFLICT DO NOTHING;

-- Viewer: read only
INSERT INTO roles (id, name, description) VALUES ('00000000-0000-0000-0000-000000000003', 'viewer', 'Read-only access to all resources') ON CONFLICT DO NOTHING;
INSERT INTO role_permissions (role_id, permission) VALUES ('00000000-0000-0000-0000-000000000003', 'instance:read') ON CONFLICT DO NOTHING;
INSERT INTO role_permissions (role_id, permission) VALUES ('00000000-0000-0000-0000-000000000003', 'vpc:read') ON CONFLICT DO NOTHING;
INSERT INTO role_permissions (role_id, permission) VALUES ('00000000-0000-0000-0000-000000000003', 'volume:read') ON CONFLICT DO NOTHING;
INSERT INTO role_permissions (role_id, permission) VALUES ('00000000-0000-0000-0000-000000000003', 'snapshot:read') ON CONFLICT DO NOTHING;
