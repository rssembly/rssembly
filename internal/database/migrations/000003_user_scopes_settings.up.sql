-- Migration 000003: User scopes + global/user settings
-- Up

-- ── Users: replace is_admin with scopes ───────────────────────────

ALTER TABLE users ADD COLUMN scopes TEXT[] NOT NULL DEFAULT '{}';
-- Migrate existing admins: is_admin = true → scope = ["*"]
UPDATE users SET scopes = ARRAY['*'] WHERE is_admin = true;

-- ── Global settings ───────────────────────────────────────────────

CREATE TABLE settings (
    id           BYTEA PRIMARY KEY,
    key          TEXT NOT NULL UNIQUE,
    default_value JSONB NOT NULL DEFAULT 'null',
    type         TEXT NOT NULL DEFAULT 'string',
    description  TEXT NOT NULL DEFAULT '',
    created_at   TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at   TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX idx_settings_key ON settings (key);

-- ── User settings overrides ───────────────────────────────────────

CREATE TABLE user_settings (
    user_id    BYTEA NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    key        TEXT NOT NULL,
    value      JSONB NOT NULL DEFAULT 'null',
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    PRIMARY KEY (user_id, key)
);
