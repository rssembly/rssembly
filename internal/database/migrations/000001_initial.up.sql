-- Initial schema for Rssembly.
-- Migration 000001 (up)

CREATE EXTENSION IF NOT EXISTS pgcrypto;

-- ── Users ─────────────────────────────────────────────────────────

CREATE TABLE users (
    id           BYTEA PRIMARY KEY,  -- UUIDv7 (16 bytes)
    email        TEXT NOT NULL UNIQUE,
    password_hash TEXT NOT NULL,
    display_name TEXT NOT NULL DEFAULT '',
    is_admin     BOOLEAN NOT NULL DEFAULT false,
    created_at   TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at   TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX idx_users_email ON users (email);

-- ── API Keys ───────────────────────────────────────────────────────

CREATE TABLE api_keys (
    id          BYTEA PRIMARY KEY,
    user_id     BYTEA NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    name        TEXT NOT NULL,
    prefix      TEXT NOT NULL,
    hash        TEXT NOT NULL,
    last_used_at TIMESTAMPTZ,
    expires_at  TIMESTAMPTZ,
    is_active   BOOLEAN NOT NULL DEFAULT true,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX idx_api_keys_user_id ON api_keys (user_id);
CREATE INDEX idx_api_keys_prefix ON api_keys (prefix);

-- ── Folders ────────────────────────────────────────────────────────

CREATE TABLE folders (
    id          BYTEA PRIMARY KEY,
    user_id     BYTEA NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    name        TEXT NOT NULL,
    parent_id   BYTEA REFERENCES folders(id) ON DELETE SET NULL,
    sort_order  INT NOT NULL DEFAULT 0,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT now(),
    UNIQUE (user_id, name)
);

CREATE INDEX idx_folders_user_id ON folders (user_id);

-- ── Feeds ──────────────────────────────────────────────────────────

CREATE TABLE feeds (
    id                   BYTEA PRIMARY KEY,
    user_id              BYTEA NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    title                TEXT NOT NULL DEFAULT '',
    description          TEXT NOT NULL DEFAULT '',
    feed_url             TEXT NOT NULL,
    site_url             TEXT NOT NULL DEFAULT '',
    icon_url             TEXT NOT NULL DEFAULT '',

    -- Polling state
    poll_interval        INTERVAL NOT NULL DEFAULT '15 minutes',
    next_poll_at         TIMESTAMPTZ NOT NULL DEFAULT now(),
    last_fetched_at      TIMESTAMPTZ,
    consecutive_failures INT NOT NULL DEFAULT 0,
    etag                 TEXT NOT NULL DEFAULT '',
    last_modified        TEXT NOT NULL DEFAULT '',
    is_paused            BOOLEAN NOT NULL DEFAULT false,

    folder_id            BYTEA REFERENCES folders(id) ON DELETE SET NULL,
    created_at           TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at           TIMESTAMPTZ NOT NULL DEFAULT now(),

    UNIQUE (user_id, feed_url)
);

CREATE INDEX idx_feeds_user_id ON feeds (user_id);
CREATE INDEX idx_feeds_next_poll ON feeds (next_poll_at) WHERE NOT is_paused;

-- ── Articles ───────────────────────────────────────────────────────

CREATE TABLE articles (
    id           BYTEA PRIMARY KEY,
    feed_id      BYTEA NOT NULL REFERENCES feeds(id) ON DELETE CASCADE,
    guid         TEXT NOT NULL,
    url          TEXT NOT NULL DEFAULT '',
    title        TEXT NOT NULL DEFAULT '',
    content      TEXT NOT NULL DEFAULT '',
    summary      TEXT NOT NULL DEFAULT '',
    author       TEXT NOT NULL DEFAULT '',
    published_at TIMESTAMPTZ,
    updated_at   TIMESTAMPTZ,
    created_at   TIMESTAMPTZ NOT NULL DEFAULT now(),

    UNIQUE (feed_id, guid)
);

CREATE INDEX idx_articles_feed_id ON articles (feed_id);
CREATE INDEX idx_articles_published ON articles (published_at DESC);

-- ── Read States (per-user per-article) ────────────────────────────

CREATE TYPE read_state AS ENUM ('unread', 'read', 'saved');

CREATE TABLE read_states (
    user_id    BYTEA NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    article_id BYTEA NOT NULL REFERENCES articles(id) ON DELETE CASCADE,
    state      read_state NOT NULL DEFAULT 'unread',
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    PRIMARY KEY (user_id, article_id)
);

CREATE INDEX idx_read_states_user ON read_states (user_id, state);

-- ── Full-text search ───────────────────────────────────────────────

ALTER TABLE articles ADD COLUMN search_vector tsvector
    GENERATED ALWAYS AS (
        to_tsvector('english', coalesce(title, '') || ' ' || coalesce(content, ''))
    ) STORED;

CREATE INDEX idx_articles_search ON articles USING GIN (search_vector);