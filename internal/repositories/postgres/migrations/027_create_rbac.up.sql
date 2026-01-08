-- +goose Up

CREATE TABLE roles (
    id UUID PRIMARY KEY,
    name TEXT UNIQUE NOT NULL,
    description TEXT,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE role_permissions (
    role_id UUID REFERENCES roles(id) ON DELETE CASCADE,
    permission TEXT NOT NULL,
    PRIMARY KEY (role_id, permission)
);

-- Insert default roles
-- Admin: *
INSERT INTO roles (id, name, description) VALUES ('00000000-0000-0000-0000-000000000001', 'admin', 'Full platform access');
INSERT INTO role_permissions (role_id, permission) VALUES ('00000000-0000-0000-0000-000000000001', '*');

-- Developer: basic resource management
INSERT INTO roles (id, name, description) VALUES ('00000000-0000-0000-0000-000000000002', 'developer', 'Access to manage compute, storage and network resources');
INSERT INTO role_permissions (role_id, permission) VALUES ('00000000-0000-0000-0000-000000000002', 'instance:launch');
INSERT INTO role_permissions (role_id, permission) VALUES ('00000000-0000-0000-0000-000000000002', 'instance:terminate');
INSERT INTO role_permissions (role_id, permission) VALUES ('00000000-0000-0000-0000-000000000002', 'instance:read');
INSERT INTO role_permissions (role_id, permission) VALUES ('00000000-0000-0000-0000-000000000002', 'vpc:read');
INSERT INTO role_permissions (role_id, permission) VALUES ('00000000-0000-0000-0000-000000000002', 'volume:read');
INSERT INTO role_permissions (role_id, permission) VALUES ('00000000-0000-0000-0000-000000000002', 'snapshot:read');

-- Viewer: read only
INSERT INTO roles (id, name, description) VALUES ('00000000-0000-0000-0000-000000000003', 'viewer', 'Read-only access to all resources');
INSERT INTO role_permissions (role_id, permission) VALUES ('00000000-0000-0000-0000-000000000003', 'instance:read');
INSERT INTO role_permissions (role_id, permission) VALUES ('00000000-0000-0000-0000-000000000003', 'vpc:read');
INSERT INTO role_permissions (role_id, permission) VALUES ('00000000-0000-0000-0000-000000000003', 'volume:read');
INSERT INTO role_permissions (role_id, permission) VALUES ('00000000-0000-0000-0000-000000000003', 'snapshot:read');
