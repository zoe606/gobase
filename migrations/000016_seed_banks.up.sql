-- Seed bank data
INSERT INTO banks (name, code, default_password) VALUES
    ('Bank Central Asia', 'BCA', NULL),
    ('Bank Rakyat Indonesia', 'BRI', NULL)
ON CONFLICT (code) DO NOTHING;
