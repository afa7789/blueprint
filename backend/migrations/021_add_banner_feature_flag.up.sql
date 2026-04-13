-- Migration 021: Add banner feature flag

INSERT INTO feature_flags (key, enabled) VALUES
    ('banners_enabled', false)
ON CONFLICT (key) DO NOTHING;
