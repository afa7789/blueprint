-- Migration 002: Feature Flags

CREATE TABLE IF NOT EXISTS feature_flags (
    id SERIAL PRIMARY KEY,
    key VARCHAR(100) UNIQUE NOT NULL,
    enabled BOOLEAN DEFAULT true,
    updated_at TIMESTAMP DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_feature_flags_key ON feature_flags(key);

-- Seed default feature flags
INSERT INTO feature_flags (key, enabled) VALUES
    ('store_enabled', true),
    ('blog_enabled', true),
    ('waitlist_enabled', true),
    ('payments_stripe', true),
    ('payments_pix', true),
    ('pwa_enabled', true),
    ('ai_blog_enabled', true)
ON CONFLICT (key) DO NOTHING;