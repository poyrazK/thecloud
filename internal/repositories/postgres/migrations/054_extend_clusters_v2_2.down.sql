ALTER TABLE clusters 
DROP COLUMN ha_enabled,
DROP COLUMN api_server_lb_address,
DROP COLUMN job_id;

ALTER TABLE load_balancers
DROP COLUMN ip;
