-- Migration 005: Coupons and User Groups

-- User Groups (for discounts)
CREATE TABLE IF NOT EXISTS user_groups (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(255) NOT NULL,
    discount_percentage DECIMAL(5,2) DEFAULT 0
);

CREATE INDEX IF NOT EXISTS idx_user_groups_name ON user_groups(name);

-- Coupons
CREATE TABLE IF NOT EXISTS coupons (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    code VARCHAR(50) UNIQUE NOT NULL,
    discount_type VARCHAR(20),
    discount_value DECIMAL(10,2) NOT NULL,
    min_purchase DECIMAL(10,2),
    valid_from TIMESTAMP,
    valid_until TIMESTAMP,
    max_uses INTEGER,
    used_count INTEGER DEFAULT 0,
    is_active BOOLEAN DEFAULT true
);

CREATE INDEX IF NOT EXISTS idx_coupons_code ON coupons(code);
CREATE INDEX IF NOT EXISTS idx_coupons_is_active ON coupons(is_active);
CREATE INDEX IF NOT EXISTS idx_coupons_valid_period ON coupons(valid_from, valid_until);