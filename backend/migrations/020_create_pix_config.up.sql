CREATE TABLE IF NOT EXISTS pix_config (
    id SERIAL PRIMARY KEY,
    pix_key VARCHAR(255) NOT NULL DEFAULT '',
    key_type VARCHAR(20) NOT NULL DEFAULT 'random',
    beneficiary VARCHAR(25) NOT NULL DEFAULT '',
    city VARCHAR(15) NOT NULL DEFAULT '',
    updated_at TIMESTAMP DEFAULT NOW()
);

INSERT INTO pix_config (pix_key, beneficiary, city) VALUES ('', '', '')
ON CONFLICT DO NOTHING;
