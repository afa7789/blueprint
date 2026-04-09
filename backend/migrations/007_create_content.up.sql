-- Migration 007: Content (Banners, Linktree, Brand Kit)

-- Banners
CREATE TABLE IF NOT EXISTS banners (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    title VARCHAR(255),
    image_url VARCHAR(500) NOT NULL,
    link_url VARCHAR(500),
    target_profile VARCHAR(50),
    is_active BOOLEAN DEFAULT true,
    start_date TIMESTAMP,
    end_date TIMESTAMP,
    display_duration INTEGER,
    cache_key VARCHAR(100),
    order_index INTEGER DEFAULT 0
);

CREATE INDEX IF NOT EXISTS idx_banners_is_active ON banners(is_active);
CREATE INDEX IF NOT EXISTS idx_banners_target_profile ON banners(target_profile);
CREATE INDEX IF NOT EXISTS idx_banners_order_index ON banners(order_index);
CREATE INDEX IF NOT EXISTS idx_banners_cache_key ON banners(cache_key);

-- Linktree Items
CREATE TABLE IF NOT EXISTS linktree_items (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    title VARCHAR(255) NOT NULL,
    url VARCHAR(500) NOT NULL,
    icon VARCHAR(255),
    order_index INTEGER DEFAULT 0,
    is_active BOOLEAN DEFAULT true
);

CREATE INDEX IF NOT EXISTS idx_linktree_items_is_active ON linktree_items(is_active);
CREATE INDEX IF NOT EXISTS idx_linktree_items_order_index ON linktree_items(order_index);

-- Brand Kit (single row, use id as anchor for upserts)
CREATE TABLE IF NOT EXISTS brand_kit (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    primary_color VARCHAR(20) DEFAULT '#000000',
    secondary_color VARCHAR(20) DEFAULT '#ffffff',
    logo_url VARCHAR(500),
    favicon_url VARCHAR(500),
    font_family VARCHAR(255),
    updated_at TIMESTAMP DEFAULT NOW()
);

-- Insert default brand kit (upsert compatible with known id)
INSERT INTO brand_kit (id, primary_color, secondary_color) VALUES
    ('00000000-0000-0000-0000-000000000001', '#000000', '#ffffff')
ON CONFLICT (id) DO NOTHING;