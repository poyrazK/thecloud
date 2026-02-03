CREATE TABLE global_load_balancers (
    id UUID PRIMARY KEY,
    user_id UUID NOT NULL,
    tenant_id UUID NOT NULL,
    name VARCHAR(255) NOT NULL,
    hostname VARCHAR(255) UNIQUE NOT NULL,
    policy VARCHAR(50) NOT NULL,
    health_check_protocol VARCHAR(10),
    health_check_port INT,
    health_check_path VARCHAR(255),
    health_check_interval INT DEFAULT 30,
    health_check_timeout INT DEFAULT 5,
    health_check_healthy_count INT DEFAULT 2,
    health_check_unhealthy_count INT DEFAULT 2,
    status VARCHAR(50) DEFAULT 'CREATING',
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW()
);

CREATE TABLE global_lb_endpoints (
    id UUID PRIMARY KEY,
    global_lb_id UUID NOT NULL REFERENCES global_load_balancers(id) ON DELETE CASCADE,
    region VARCHAR(50) NOT NULL,
    target_type VARCHAR(10) NOT NULL, -- 'LB' or 'IP'
    target_id UUID, -- Can be null if target_type is IP
    target_ip INET, -- Can be null if target_type is LB
    weight INT DEFAULT 1,
    priority INT DEFAULT 1,
    healthy BOOLEAN DEFAULT true,
    last_health_check TIMESTAMPTZ,
    created_at TIMESTAMPTZ DEFAULT NOW()
);

CREATE INDEX idx_global_lb_user_id ON global_load_balancers(user_id);
CREATE INDEX idx_global_lb_endpoints_glb_id ON global_lb_endpoints(global_lb_id);
