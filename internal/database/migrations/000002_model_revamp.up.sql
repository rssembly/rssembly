-- Migration 000002: Model revamp — soft delete, scopes, status, auth
-- Up

-- ── Users ────────────────────────────────────────────────────────

ALTER TABLE users ADD COLUMN username TEXT NOT NULL DEFAULT '';
UPDATE users SET username = display_name WHERE username = '';
ALTER TABLE users ALTER COLUMN username DROP DEFAULT;

ALTER TABLE users ADD COLUMN deleted_at TIMESTAMPTZ;
CREATE UNIQUE INDEX idx_users_username ON users (username) WHERE deleted_at IS NULL;

-- ── API Keys ──────────────────────────────────────────────────────

ALTER TABLE api_keys RENAME COLUMN user_id TO created_by;
ALTER TABLE api_keys ADD COLUMN scopes TEXT[] NOT NULL DEFAULT '{}';
ALTER TABLE api_keys ADD COLUMN updated_at TIMESTAMPTZ NOT NULL DEFAULT now();
ALTER TABLE api_keys ADD COLUMN deleted_at TIMESTAMPTZ;

-- ── Folders ──────────────────────────────────────────────────────

ALTER TABLE folders ADD COLUMN deleted_at TIMESTAMPTZ;

-- ── Feeds ────────────────────────────────────────────────────────

ALTER TABLE feeds RENAME COLUMN user_id TO created_by;

-- Replace consecutive_failures with status column
CREATE TYPE feed_status AS ENUM ('ok', 'error', 'paused');
ALTER TABLE feeds ADD COLUMN status feed_status NOT NULL DEFAULT 'ok';
UPDATE feeds SET status = 'paused' WHERE is_paused = true;
UPDATE feeds SET status = 'error' WHERE consecutive_failures > 3;
ALTER TABLE feeds DROP COLUMN consecutive_failures;

-- Feed authentication
ALTER TABLE feeds ADD COLUMN username TEXT NOT NULL DEFAULT '';
ALTER TABLE feeds ADD COLUMN password_encrypted TEXT NOT NULL DEFAULT '';

-- Retention
ALTER TABLE feeds ADD COLUMN max_entries INT NOT NULL DEFAULT 0;

ALTER TABLE feeds ADD COLUMN deleted_at TIMESTAMPTZ;

-- ── Articles ─────────────────────────────────────────────────────

-- Remove unused columns, add soft delete
ALTER TABLE articles DROP COLUMN IF EXISTS summary;
ALTER TABLE articles DROP COLUMN IF EXISTS author;
ALTER TABLE articles ADD COLUMN deleted_at TIMESTAMPTZ;

-- Update the full-text search to exclude deleted articles.
DROP INDEX IF EXISTS idx_articles_search;
CREATE INDEX idx_articles_search ON articles USING GIN (search_vector)
    WHERE deleted_at IS NULL;

-- ── Feed scheduling index ─────────────────────────────────────────

CREATE INDEX idx_feeds_next_poll_active ON feeds (next_poll_at)
    WHERE status != 'paused' AND deleted_at IS NULL;
