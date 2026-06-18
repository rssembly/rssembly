-- Migration 000003: User scopes + global/user settings
-- Down

DROP TABLE IF EXISTS user_settings;
DROP TABLE IF EXISTS settings;
ALTER TABLE users DROP COLUMN IF EXISTS scopes;
