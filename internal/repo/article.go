package repo

import (
	"context"
	"encoding/hex"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"

	"github.com/rssembly/rssembly/internal/database"
	"github.com/rssembly/rssembly/internal/models"
)

// ArticleRepo handles article and read-state database operations.
type ArticleRepo struct {
	db *database.Pool
}

// NewArticleRepo creates an ArticleRepo.
func NewArticleRepo(db *database.Pool) *ArticleRepo {
	return &ArticleRepo{db: db}
}

// CreateArticle inserts a new article. If an article with the same (feed_id, guid)
// already exists, it returns the existing article (idempotent dedup).
// Uses RETURNING so the caller always gets the authoritative row.
func (r *ArticleRepo) CreateArticle(ctx context.Context, article *models.Article) (*models.Article, error) {
	article.ID = models.NewUUIDv7()

	row := r.db.QueryRow(ctx, `
		INSERT INTO articles
			(id, feed_id, guid, url, title, content, published_at, created_at)
		VALUES
			($1, $2, $3, $4, $5, $6, $7, now())
		ON CONFLICT (feed_id, guid) DO UPDATE SET
			id = articles.id  -- no-op to allow RETURNING
		RETURNING id, feed_id, guid, url, title, content, published_at, created_at
	`, article.ID[:], article.FeedID[:], article.GUID, article.URL, article.Title, article.Content, article.PublishedAt)

	created := &models.Article{}
	err := row.Scan(&created.ID, &created.FeedID, &created.GUID, &created.URL, &created.Title, &created.Content, &created.PublishedAt, &created.CreatedAt)
	if err != nil {
		return nil, fmt.Errorf("create article: %w", err)
	}

	return created, nil
}

// GetArticleByID returns a single article by its primary key.
func (r *ArticleRepo) GetArticleByID(ctx context.Context, id models.UUIDv7) (*models.Article, error) {
	article := &models.Article{}

	err := r.db.QueryRow(ctx, `
		SELECT id, feed_id, guid, url, title, content, published_at, created_at, deleted_at
		FROM articles
		WHERE id = $1 AND deleted_at IS NULL
	`, id[:]).Scan(&article.ID, &article.FeedID, &article.GUID, &article.URL, &article.Title, &article.Content, &article.PublishedAt, &article.CreatedAt, &article.DeletedAt)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("get article: %w", err)
	}
	return article, nil
}

// ListArticlesParams holds optional filters for listing articles.
type ListArticlesParams struct {
	UserID    models.UUIDv7
	FeedID    models.UUIDv7 // zero value means no filter
	ReadState string        // "unread", "read", "saved", or empty for all
	Search    string        // full-text search query, empty for no filter
	Cursor    string        // hex-encoded UUIDv7 cursor
	Limit     int           // default 50, max 100
}

// ListArticles returns a cursor-paginated list of articles matching the given filters.
// If UserID is set, it JOINs read_states and populates ReadState on each article.
func (r *ArticleRepo) ListArticles(ctx context.Context, params ListArticlesParams) ([]*models.Article, string, bool, error) {
	if params.Limit <= 0 || params.Limit > 100 {
		params.Limit = 50
	}

	args := make([]any, 0, 8)
	argIdx := 1

	query := `
		SELECT a.id, a.feed_id, a.guid, a.url, a.title, a.content, a.published_at, a.created_at, a.deleted_at
	`

	if params.UserID != models.NilUUIDv7 {
		query += `, rs.state AS read_state`
	} else {
		query += `, NULL::read_state AS read_state`
	}

	query += ` FROM articles a`

	if params.UserID != models.NilUUIDv7 {
		query += ` LEFT JOIN read_states rs ON rs.article_id = a.id AND rs.user_id = $` + fmt.Sprintf("%d", argIdx)
		args = append(args, params.UserID[:])
		argIdx++
	}

	query += ` WHERE a.deleted_at IS NULL`

	if params.FeedID != models.NilUUIDv7 {
		query += ` AND a.feed_id = $` + fmt.Sprintf("%d", argIdx)
		args = append(args, params.FeedID[:])
		argIdx++
	}

	if params.ReadState != "" && params.UserID != models.NilUUIDv7 {
		query += ` AND rs.state = $` + fmt.Sprintf("%d", argIdx)
		args = append(args, params.ReadState)
		argIdx++
	}

	if params.Search != "" {
		query += ` AND a.search_vector @@ plainto_tsquery('english', $` + fmt.Sprintf("%d", argIdx) + `)`
		args = append(args, params.Search)
		argIdx++
	}

	if params.Cursor != "" {
		decoded, err := hex.DecodeString(params.Cursor)
		if err != nil {
			return nil, "", false, fmt.Errorf("decode cursor: %w", err)
		}
		if len(decoded) != 16 {
			return nil, "", false, fmt.Errorf("invalid cursor length: %d", len(decoded))
		}
		query += ` AND (a.published_at, a.id) < (SELECT published_at, id FROM articles WHERE id = $` + fmt.Sprintf("%d", argIdx) + `)`
		args = append(args, decoded)
		argIdx++
	}

	query += ` ORDER BY a.published_at DESC NULLS LAST, a.id DESC LIMIT $` + fmt.Sprintf("%d", argIdx)
	args = append(args, params.Limit+1)

	rows, err := r.db.Query(ctx, query, args...)
	if err != nil {
		return nil, "", false, fmt.Errorf("list articles: %w", err)
	}
	defer rows.Close()

	var articles []*models.Article
	for rows.Next() {
		a := &models.Article{}
		var readState *string
		if err := rows.Scan(&a.ID, &a.FeedID, &a.GUID, &a.URL, &a.Title, &a.Content, &a.PublishedAt, &a.CreatedAt, &a.DeletedAt, &readState); err != nil {
			return nil, "", false, fmt.Errorf("scan article: %w", err)
		}
		if readState != nil {
			state := models.ReadState(*readState)
			a.ReadState = &state
		}
		articles = append(articles, a)
	}
	if err := rows.Err(); err != nil {
		return nil, "", false, fmt.Errorf("rows iteration: %w", err)
	}

	hasMore := len(articles) > params.Limit
	if hasMore {
		articles = articles[:params.Limit]
	}

	nextCursor := ""
	if hasMore && len(articles) > 0 {
		nextCursor = hex.EncodeToString(articles[len(articles)-1].ID[:])
	}

	return articles, nextCursor, hasMore, nil
}

// SetReadState creates or updates the read state for a user+article pair.
func (r *ArticleRepo) SetReadState(ctx context.Context, userID, articleID models.UUIDv7, state models.ReadState) error {
	_, err := r.db.Exec(ctx, `
		INSERT INTO read_states (user_id, article_id, state, updated_at)
		VALUES ($1, $2, $3, now())
		ON CONFLICT (user_id, article_id)
		DO UPDATE SET state = $3, updated_at = now()
	`, userID[:], articleID[:], string(state))
	if err != nil {
		return fmt.Errorf("set read state: %w", err)
	}
	return nil
}

// GetArticleByGUID looks up an article by feed_id + guid (for dedup checks).
// Returns ErrNotFound if no match.
func (r *ArticleRepo) GetArticleByGUID(ctx context.Context, feedID models.UUIDv7, guid string) (*models.Article, error) {
	a := &models.Article{}
	err := r.db.QueryRow(ctx, `
		SELECT id, feed_id, guid, url, title, content, published_at, created_at, deleted_at
		FROM articles
		WHERE feed_id = $1 AND guid = $2 AND deleted_at IS NULL
	`, feedID[:], guid).Scan(&a.ID, &a.FeedID, &a.GUID, &a.URL, &a.Title, &a.Content, &a.PublishedAt, &a.CreatedAt, &a.DeletedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("get article by guid: %w", err)
	}
	return a, nil
}
