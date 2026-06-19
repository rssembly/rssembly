package repo

import (
	"context"
	"testing"
	"time"

	"github.com/testcontainers/testcontainers-go/modules/postgres"

	"github.com/rssembly/rssembly/internal/database"
	"github.com/rssembly/rssembly/internal/models"
)

// testDB starts a PostgreSQL container, runs migrations, and returns a pool
// ready for use. The container is terminated and pool closed via t.Cleanup.
func testDB(t *testing.T) *database.Pool {
	t.Helper()

	ctx := context.Background()

	pg, err := postgres.Run(ctx,
		"postgres:16-alpine",
		postgres.WithDatabase("rssembly_test"),
		postgres.WithUsername("test"),
		postgres.WithPassword("test"),
		postgres.BasicWaitStrategies(),
	)
	if err != nil {
		t.Fatalf("testcontainers: start postgres: %v", err)
	}
	t.Cleanup(func() { _ = pg.Terminate(ctx) })

	connStr, err := pg.ConnectionString(ctx, "sslmode=disable")
	if err != nil {
		t.Fatalf("testcontainers: get connection string: %v", err)
	}

	pool, err := database.Connect(ctx, connStr)
	if err != nil {
		t.Fatalf("database: connect: %v", err)
	}
	t.Cleanup(pool.Close)

	if err := database.RunMigrations(connStr); err != nil {
		t.Fatalf("database: run migrations: %v", err)
	}

	return pool
}

// seedUser inserts a test user and returns it.
func seedUser(t *testing.T, db *database.Pool) *models.User {
	t.Helper()
	ctx := context.Background()

	u := &models.User{
		ID:           models.NewUUIDv7(),
		Username:     "testuser_" + t.Name(),
		Email:        t.Name() + "@test.com",
		PasswordHash: "argon2id$v=19$...",
		Scopes:       []string{"feeds:read", "feeds:write", "articles:read", "articles:write", "folders:read", "folders:write"},
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}

	_, err := db.Exec(ctx, `
		INSERT INTO users (id, username, email, password_hash, scopes, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
	`, u.ID[:], u.Username, u.Email, u.PasswordHash, u.Scopes, u.CreatedAt, u.UpdatedAt)
	if err != nil {
		t.Fatalf("seed user: %v", err)
	}
	return u
}

// seedFeed inserts a test feed owned by the given user and returns it.
func seedFeed(t *testing.T, db *database.Pool, userID models.UUIDv7) *models.Feed {
	t.Helper()
	ctx := context.Background()

	f := &models.Feed{
		ID:           models.NewUUIDv7(),
		CreatedBy:    userID,
		Title:        "Test Feed " + t.Name(),
		FeedURL:      "https://" + t.Name() + ".example.com/feed.xml",
		Status:       models.FeedStatusOK,
		PollInterval: 15 * time.Minute,
		NextPollAt:   time.Now(),
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}

	_, err := db.Exec(ctx, `
		INSERT INTO feeds
			(id, created_by, title, description, feed_url, site_url, icon_url,
			 poll_interval, next_poll_at, status, etag, last_modified, is_paused,
			 max_entries, created_at, updated_at)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,$15,$16)
	`, f.ID[:], f.CreatedBy[:], f.Title, "", f.FeedURL, "", "",
		f.PollInterval, f.NextPollAt, f.Status, "", "", false,
		0, f.CreatedAt, f.UpdatedAt)
	if err != nil {
		t.Fatalf("seed feed: %v", err)
	}
	return f
}