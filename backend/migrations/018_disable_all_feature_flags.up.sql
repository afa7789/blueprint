-- Migration 018: All feature flags start disabled by default
-- Admin enables what they need via the panel

UPDATE feature_flags SET enabled = false;

-- Also change the table default for future flags
ALTER TABLE feature_flags ALTER COLUMN enabled SET DEFAULT false;
