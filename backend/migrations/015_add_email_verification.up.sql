-- Migration 015: Email Verification

ALTER TABLE users ADD COLUMN IF NOT EXISTS email_verified BOOLEAN DEFAULT false;
ALTER TABLE users ADD COLUMN IF NOT EXISTS email_verified_at TIMESTAMP;

-- Existing users are considered verified
UPDATE users SET email_verified = true, email_verified_at = NOW() WHERE email_verified = false;

-- Add email verification feature flag
INSERT INTO feature_flags (key, enabled) VALUES ('email_verification_required', false)
ON CONFLICT (key) DO NOTHING;

-- Security settings table (admin-configurable rate limits etc.)
CREATE TABLE IF NOT EXISTS security_settings (
    id SERIAL PRIMARY KEY,
    key VARCHAR(100) UNIQUE NOT NULL,
    value VARCHAR(255) NOT NULL,
    description TEXT,
    updated_at TIMESTAMP DEFAULT NOW()
);

INSERT INTO security_settings (key, value, description) VALUES
    ('rate_limit_api', '60', 'Max API requests per minute per IP'),
    ('rate_limit_auth', '10', 'Max auth requests per minute per email'),
    ('rate_limit_register', '5', 'Max register requests per hour per email'),
    ('rate_limit_forgot', '3', 'Max forgot-password requests per hour per email'),
    ('max_login_attempts', '5', 'Max failed login attempts before temporary lock'),
    ('login_lock_duration', '15', 'Account lock duration in minutes after max attempts'),
    ('password_min_length', '6', 'Minimum password length'),
    ('session_max_age', '168', 'Max session age in hours (refresh token)')
ON CONFLICT (key) DO NOTHING;
