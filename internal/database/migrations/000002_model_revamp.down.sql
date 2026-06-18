-- Migration 000002: Model revamp — down
-- Reverts all changes made in the up migration.

DROP INDEX IF EXISTS idx_feeds_next_poll_active;

DROP INDEX IF EXISTS idx_articles_search;
CREATE INDEX idx_articles_search ON articles USING GIN (search_vector);

ALTER TABLE articles DROP COLUMN IF EXISTS deleted_at;
ALTER TABLE articles ADD COLUMN summary TEXT NOT NULL DEFAULT '';
ALTER TABLE articles ADD COLUMN author TEXT NOT NULL DEFAULT '';

ALTER TABLE feeds DROP COLUMN IF EXISTS deleted_at;
ALTER TABLE feeds DROP COLUMN IF EXISTS max_entries;
ALTER TABLE feeds DROP COLUMN IF EXISTS password_encrypted;
ALTER TABLE feeds DROP COLUMN IF EXISTS username;
ALTER TABLE feeds ADD COLUMN consecutive_failures INT NOT NULL DEFAULT 0;
UPDATE feeds SET consecutive_failures = 0;
ALTER TABLE feeds DROP COLUMN IF EXISTS status;
DROP TYPE IF EXISTS feed_status;
ALTER TABLE feeds RENAME COLUMN created_by TO user_id;

ALTER TABLE folders DROP COLUMN IF EXISTS deleted_at;

ALTER TABLE api_keys DROP COLUMN IF EXISTS deleted_at;
ALTER TABLE api_keys DROP COLUMN IF EXISTS updated_at;
ALTER TABLE api_keys DROP COLUMN IF EXISTS scopes;
ALTER TABLE api_keys RENAME COLUMN created_by TO user_id;

ALTER TABLE users DROP COLUMN IF EXISTS deleted_at;
ALTER TABLE users DROP COLUMN IF EXISTS username;
DROP INDEX IF EXISTS idx_users_username;
