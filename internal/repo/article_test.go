package repo

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/rssembly/rssembly/internal/models"
)

func TestArticleRepo_CreateAndGet(t *testing.T) {
	db := testDB(t)
	repo := NewArticleRepo(db)
	ctx := context.Background()

	user := seedUser(t, db)
	feed := seedFeed(t, db, user.ID)

	article, err := repo.CreateArticle(ctx, &models.Article{
		FeedID: feed.ID,
		GUID:   "guid-1",
		URL:    "https://example.com/article-1",
		Title:  "Test Article",
		Content: "<p>Hello world</p>",
	})
	require.NoError(t, err)
	require.NotEqual(t, models.NilUUIDv7, article.ID)
	require.Equal(t, "guid-1", article.GUID)

	got, err := repo.GetArticleByID(ctx, article.ID)
	require.NoError(t, err)
	require.Equal(t, article.Title, got.Title)
}

func TestArticleRepo_DedupOnConflict(t *testing.T) {
	db := testDB(t)
	repo := NewArticleRepo(db)
	ctx := context.Background()

	user := seedUser(t, db)
	feed := seedFeed(t, db, user.ID)

	// Same (feed_id, guid) — second insert should return the existing row
	a1, err := repo.CreateArticle(ctx, &models.Article{
		FeedID: feed.ID,
		GUID:   "dedup-guid",
		Title:  "Original",
	})
	require.NoError(t, err)

	a2, err := repo.CreateArticle(ctx, &models.Article{
		FeedID: feed.ID,
		GUID:   "dedup-guid",
		Title:  "Duplicate",
	})
	require.NoError(t, err)
	require.Equal(t, a1.ID, a2.ID, "should return same ID on conflict")
}

func TestArticleRepo_GetNotFound(t *testing.T) {
	db := testDB(t)
	repo := NewArticleRepo(db)
	ctx := context.Background()

	_, err := repo.GetArticleByID(ctx, models.NewUUIDv7())
	require.ErrorIs(t, err, ErrNotFound)
}

func TestArticleRepo_ListArticles(t *testing.T) {
	db := testDB(t)
	repo := NewArticleRepo(db)
	ctx := context.Background()

	user := seedUser(t, db)
	feed := seedFeed(t, db, user.ID)

	// Insert 3 articles
	for i := range 3 {
		_, err := repo.CreateArticle(ctx, &models.Article{
			FeedID: feed.ID,
			GUID:   "guid-" + string(rune('0'+i)),
			Title:  "Article " + string(rune('A'+i)),
		})
		require.NoError(t, err)
	}

	articles, cursor, hasMore, err := repo.ListArticles(ctx, ListArticlesParams{
		UserID: user.ID,
		Limit:  10,
	})
	require.NoError(t, err)
	require.Len(t, articles, 3)
	require.Empty(t, cursor)
	require.False(t, hasMore)
}

func TestArticleRepo_ListArticlesByFeed(t *testing.T) {
	db := testDB(t)
	repo := NewArticleRepo(db)
	ctx := context.Background()

	user := seedUser(t, db)
	feed1 := seedFeed(t, db, user.ID)
	feed2 := seedFeed(t, db, user.ID)

	_, err := repo.CreateArticle(ctx, &models.Article{FeedID: feed1.ID, GUID: "f1a1", Title: "F1 Article"})
	require.NoError(t, err)
	_, err = repo.CreateArticle(ctx, &models.Article{FeedID: feed2.ID, GUID: "f2a1", Title: "F2 Article"})
	require.NoError(t, err)

	articles, _, _, err := repo.ListArticles(ctx, ListArticlesParams{
		FeedID: feed1.ID,
		Limit:  10,
	})
	require.NoError(t, err)
	require.Len(t, articles, 1)
	require.Equal(t, "F1 Article", articles[0].Title)
}

func TestArticleRepo_ListArticlesWithReadState(t *testing.T) {
	db := testDB(t)
	repo := NewArticleRepo(db)
	ctx := context.Background()

	user := seedUser(t, db)
	feed := seedFeed(t, db, user.ID)

	article, err := repo.CreateArticle(ctx, &models.Article{FeedID: feed.ID, GUID: "rs-guid", Title: "Read State Test"})
	require.NoError(t, err)

	// Mark as read
	err = repo.SetReadState(ctx, user.ID, article.ID, models.ReadStateRead)
	require.NoError(t, err)

	// List with user context — read_state should be populated
	articles, _, _, err := repo.ListArticles(ctx, ListArticlesParams{
		UserID: user.ID,
		Limit:  10,
	})
	require.NoError(t, err)
	require.Len(t, articles, 1)
	require.NotNil(t, articles[0].ReadState)
	require.Equal(t, models.ReadStateRead, *articles[0].ReadState)
}

func TestArticleRepo_FilterByReadState(t *testing.T) {
	db := testDB(t)
	repo := NewArticleRepo(db)
	ctx := context.Background()

	user := seedUser(t, db)
	feed := seedFeed(t, db, user.ID)

	a1, err := repo.CreateArticle(ctx, &models.Article{FeedID: feed.ID, GUID: "unread-guid", Title: "Unread Article"})
	require.NoError(t, err)
	a2, err := repo.CreateArticle(ctx, &models.Article{FeedID: feed.ID, GUID: "saved-guid", Title: "Saved Article"})
	require.NoError(t, err)

	err = repo.SetReadState(ctx, user.ID, a2.ID, models.ReadStateSaved)
	require.NoError(t, err)
	err = repo.SetReadState(ctx, user.ID, a1.ID, models.ReadStateRead)
	require.NoError(t, err)

	// Filter by saved
	saved, _, _, err := repo.ListArticles(ctx, ListArticlesParams{
		UserID:    user.ID,
		ReadState: "saved",
		Limit:     10,
	})
	require.NoError(t, err)
	require.Len(t, saved, 1)
	require.Equal(t, a2.ID, saved[0].ID)
}

func TestArticleRepo_SetReadStateUpsert(t *testing.T) {
	db := testDB(t)
	repo := NewArticleRepo(db)
	ctx := context.Background()

	user := seedUser(t, db)
	feed := seedFeed(t, db, user.ID)

	article, err := repo.CreateArticle(ctx, &models.Article{FeedID: feed.ID, GUID: "upsert-guid"})
	require.NoError(t, err)

	// Set read
	err = repo.SetReadState(ctx, user.ID, article.ID, models.ReadStateRead)
	require.NoError(t, err)

	// Change to saved
	err = repo.SetReadState(ctx, user.ID, article.ID, models.ReadStateSaved)
	require.NoError(t, err)

	articles, _, _, err := repo.ListArticles(ctx, ListArticlesParams{
		UserID: user.ID,
		Limit:  10,
	})
	require.NoError(t, err)
	require.Equal(t, models.ReadStateSaved, *articles[0].ReadState)
}

func TestArticleRepo_GetArticleByGUID(t *testing.T) {
	db := testDB(t)
	repo := NewArticleRepo(db)
	ctx := context.Background()

	user := seedUser(t, db)
	feed := seedFeed(t, db, user.ID)

	created, err := repo.CreateArticle(ctx, &models.Article{
		FeedID: feed.ID,
		GUID:   "find-me-by-guid",
		Title:  "Found by GUID",
	})
	require.NoError(t, err)

	got, err := repo.GetArticleByGUID(ctx, feed.ID, "find-me-by-guid")
	require.NoError(t, err)
	require.Equal(t, created.ID, got.ID)
}

func TestArticleRepo_GetArticleByGUIDNotFound(t *testing.T) {
	db := testDB(t)
	repo := NewArticleRepo(db)
	ctx := context.Background()

	_, err := repo.GetArticleByGUID(ctx, models.NewUUIDv7(), "nonexistent")
	require.ErrorIs(t, err, ErrNotFound)
}