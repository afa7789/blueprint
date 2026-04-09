-- Migration 013: Add missing feature flags

INSERT INTO feature_flags (key, enabled) VALUES
    ('linktree_enabled', true),
    ('brand_kit_enabled', true),
    ('helper_boxes_enabled', true),
    ('pix_auto_enabled', true),
    ('pix_manual_enabled', true),
    ('email_auto_enabled', true)
ON CONFLICT (key) DO NOTHING;
