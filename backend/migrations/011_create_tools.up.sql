-- Migration 011: Admin Tools Hub

CREATE TABLE IF NOT EXISTS admin_tools (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(255) NOT NULL,
    description TEXT,
    url VARCHAR(500) NOT NULL,
    icon VARCHAR(255),
    category VARCHAR(100),
    is_active BOOLEAN DEFAULT true,
    min_role VARCHAR(50) DEFAULT 'admin',
    order_index INTEGER DEFAULT 0
);

CREATE INDEX IF NOT EXISTS idx_admin_tools_category ON admin_tools(category);
CREATE INDEX IF NOT EXISTS idx_admin_tools_is_active ON admin_tools(is_active);
CREATE INDEX IF NOT EXISTS idx_admin_tools_order_index ON admin_tools(order_index);

-- Seed default tools
INSERT INTO admin_tools (name, description, url, icon, category, order_index) VALUES
    ('pgweb', 'PostgreSQL Web UI', '', 'database', 'database', 1),
    ('Redis Commander', 'Redis Web UI', '', 'cache', 'cache', 2),
    ('MinIO Console', 'Object Storage UI', '', 'storage', 'storage', 3),
    ('Grafana', 'Metrics Dashboard', '', 'chart', 'monitoring', 4),
    ('Prometheus', 'Metrics Collection', '', 'metrics', 'monitoring', 5)
ON CONFLICT DO NOTHING;
