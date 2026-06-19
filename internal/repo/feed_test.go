package repo

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/rssembly/rssembly/internal/models"
)

func TestFeedRepo_CreateAndGet(t *testing.T) {
	db := testDB(t)
	repo := NewFeedRepo(db)
	ctx := context.Background()

	user := seedUser(t, db)

	feed := &models.Feed{
		CreatedBy: user.ID,
		Title:     "Test Blog",
		FeedURL:   "https://example.com/feed.xml",
		SiteURL:   "https://example.com",
	}

	created, err := repo.CreateFeed(ctx, feed)
	require.NoError(t, err)
	require.NotEqual(t, models.NilUUIDv7, created.ID)
	require.Equal(t, feed.Title, created.Title)
	require.Equal(t, feed.FeedURL, created.FeedURL)
	require.Equal(t, models.FeedStatusOK, created.Status)
	require.Equal(t, 15*time.Minute, created.PollInterval)

	// Get by ID
	got, err := repo.GetFeedByID(ctx, created.ID)
	require.NoError(t, err)
	require.Equal(t, created.ID, got.ID)
	require.Equal(t, created.Title, got.Title)
}

func TestFeedRepo_CreateDuplicateFeedURL(t *testing.T) {
	db := testDB(t)
	repo := NewFeedRepo(db)
	ctx := context.Background()

	user := seedUser(t, db)

	feed := &models.Feed{
		CreatedBy: user.ID,
		FeedURL:   "https://example.com/dup.xml",
	}

	_, err := repo.CreateFeed(ctx, feed)
	require.NoError(t, err)

	dup := &models.Feed{
		CreatedBy: user.ID,
		FeedURL:   "https://example.com/dup.xml",
	}

	_, err = repo.CreateFeed(ctx, dup)
	require.ErrorIs(t, err, ErrConflict)
}

func TestFeedRepo_GetNotFound(t *testing.T) {
	db := testDB(t)
	repo := NewFeedRepo(db)
	ctx := context.Background()

	_, err := repo.GetFeedByID(ctx, models.NewUUIDv7())
	require.ErrorIs(t, err, ErrNotFound)
}

func TestFeedRepo_ListFeedsByUser(t *testing.T) {
	db := testDB(t)
	repo := NewFeedRepo(db)
	ctx := context.Background()

	user := seedUser(t, db)
	other := seedUser(t, db)

	for i := range 5 {
		_, err := repo.CreateFeed(ctx, &models.Feed{
			CreatedBy: user.ID,
			Title:     "Feed " + string(rune('A'+i)),
			FeedURL:   "https://example.com/feed" + string(rune('0'+i)) + ".xml",
		})
		require.NoError(t, err)
	}

	// Create one for other user (should not appear in user's list)
	_, err := repo.CreateFeed(ctx, &models.Feed{
		CreatedBy: other.ID,
		FeedURL:   "https://other.com/feed.xml",
	})
	require.NoError(t, err)

	// List all for user
	feeds, cursor, hasMore, err := repo.ListFeedsByUser(ctx, user.ID, "", 10)
	require.NoError(t, err)
	require.Len(t, feeds, 5)
	require.Empty(t, cursor)
	require.False(t, hasMore)
}

func TestFeedRepo_ListFeedsCursorPagination(t *testing.T) {
	db := testDB(t)
	repo := NewFeedRepo(db)
	ctx := context.Background()

	user := seedUser(t, db)

	// Insert 3 feeds
	for i := range 3 {
		_, err := repo.CreateFeed(ctx, &models.Feed{
			CreatedBy: user.ID,
			Title:     "Feed " + string(rune('A'+i)),
			FeedURL:   "https://example.com/feed" + string(rune('0'+i)) + ".xml",
		})
		require.NoError(t, err)
	}

	// Page size 2 — should get 2 feeds with hasMore=true
	feeds, cursor, hasMore, err := repo.ListFeedsByUser(ctx, user.ID, "", 2)
	require.NoError(t, err)
	require.Len(t, feeds, 2)
	require.True(t, hasMore)
	require.NotEmpty(t, cursor)

	// Second page
	feeds2, cursor2, hasMore2, err := repo.ListFeedsByUser(ctx, user.ID, cursor, 2)
	require.NoError(t, err)
	require.Len(t, feeds2, 1)
	require.False(t, hasMore2)
	require.Empty(t, cursor2)
}

func TestFeedRepo_UpdateFeed(t *testing.T) {
	db := testDB(t)
	repo := NewFeedRepo(db)
	ctx := context.Background()

	user := seedUser(t, db)

	feed := &models.Feed{
		CreatedBy: user.ID,
		Title:     "Original Title",
		FeedURL:   "https://example.com/feed.xml",
	}
	created, err := repo.CreateFeed(ctx, feed)
	require.NoError(t, err)

	created.Title = "Updated Title"
	created.IsPaused = true

	updated, err := repo.UpdateFeed(ctx, created)
	require.NoError(t, err)
	require.Equal(t, "Updated Title", updated.Title)
	require.True(t, updated.IsPaused)
}

func TestFeedRepo_DeleteFeed(t *testing.T) {
	db := testDB(t)
	repo := NewFeedRepo(db)
	ctx := context.Background()

	user := seedUser(t, db)

	feed := &models.Feed{
		CreatedBy: user.ID,
		FeedURL:   "https://example.com/feed.xml",
	}
	created, err := repo.CreateFeed(ctx, feed)
	require.NoError(t, err)

	err = repo.DeleteFeed(ctx, created.ID)
	require.NoError(t, err)

	// Should be gone
	_, err = repo.GetFeedByID(ctx, created.ID)
	require.ErrorIs(t, err, ErrNotFound)
}

func TestFeedRepo_DeleteNotFound(t *testing.T) {
	db := testDB(t)
	repo := NewFeedRepo(db)
	ctx := context.Background()

	err := repo.DeleteFeed(ctx, models.NewUUIDv7())
	require.ErrorIs(t, err, ErrNotFound)
}

func TestFeedRepo_UpdateNotFound(t *testing.T) {
	db := testDB(t)
	repo := NewFeedRepo(db)
	ctx := context.Background()

	feed := &models.Feed{
		ID:      models.NewUUIDv7(),
		Title:   "Ghost Feed",
		FeedURL: "https://ghost.com/feed.xml",
		Status:  models.FeedStatusOK,
	}

	_, err := repo.UpdateFeed(ctx, feed)
	require.ErrorIs(t, err, ErrNotFound)
}
