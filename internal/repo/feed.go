package repo

import (
	"context"
	"encoding/hex"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"

	"github.com/rssembly/rssembly/internal/database"
	"github.com/rssembly/rssembly/internal/models"
)

// FeedRepo handles feed subscription database operations.
type FeedRepo struct {
	db *database.Pool
}

// NewFeedRepo creates a FeedRepo.
func NewFeedRepo(db *database.Pool) *FeedRepo {
	return &FeedRepo{db: db}
}

// CreateFeed inserts a new feed subscription and returns it with its generated ID.
// If the user already subscribes to this feed_url, it returns ErrConflict.
func (r *FeedRepo) CreateFeed(ctx context.Context, feed *models.Feed) (*models.Feed, error) {
	feed.ID = models.NewUUIDv7()
	now := time.Now()
	feed.CreatedAt = now
	feed.UpdatedAt = now
	feed.NextPollAt = now

	if feed.PollInterval == 0 {
		feed.PollInterval = 15 * time.Minute
	}
	if feed.Status == "" {
		feed.Status = models.FeedStatusOK
	}

	var folderIDBytes []byte
	if feed.FolderID != nil {
		folderIDBytes = feed.FolderID[:]
	}

	_, err := r.db.Exec(ctx, `
		INSERT INTO feeds
			(id, created_by, title, description, feed_url, site_url, icon_url,
			 poll_interval, next_poll_at, status, etag, last_modified, is_paused,
			 max_entries, folder_id, username, password_encrypted,
			 created_at, updated_at)
		VALUES
			($1, $2, $3, $4, $5, $6, $7,
			 $8, $9, $10, $11, $12, $13,
			 $14, $15, $16, $17,
			 $18, $19)
	`, feed.ID[:], feed.CreatedBy[:], feed.Title, feed.Description, feed.FeedURL, feed.SiteURL, feed.IconURL,
		feed.PollInterval, feed.NextPollAt, string(feed.Status), feed.ETag, feed.LastModified, feed.IsPaused,
		feed.MaxEntries, folderIDBytes, feed.Username, feed.PasswordEncrypted,
		feed.CreatedAt, feed.UpdatedAt)

	if err != nil {
		if isUniqueViolation(err) {
			return nil, fmt.Errorf("create feed: %w", ErrConflict)
		}
		return nil, fmt.Errorf("create feed: %w", err)
	}

	return feed, nil
}

// GetFeedByID returns a feed by its primary key.
func (r *FeedRepo) GetFeedByID(ctx context.Context, id models.UUIDv7) (*models.Feed, error) {
	feed, err := scanFeed(ctx, r.db, `WHERE f.id = $1 AND f.deleted_at IS NULL`, id[:])
	if err != nil {
		return nil, err
	}
	return feed, nil
}

// ListFeedsByUser returns a cursor-paginated list of feeds owned by the given user.
// Cursor is a hex-encoded UUIDv7 of the last visible item. Pass empty string for the first page.
func (r *FeedRepo) ListFeedsByUser(ctx context.Context, userID models.UUIDv7, cursor string, limit int) ([]*models.Feed, string, bool, error) {
	if limit <= 0 || limit > 100 {
		limit = 50
	}

	args := []any{userID[:]}
	argIdx := 2

	query := `
		SELECT f.id, f.created_by, f.title, f.description, f.feed_url, f.site_url, f.icon_url,
		       f.poll_interval, f.next_poll_at, f.last_fetched_at, f.status,
		       f.etag, f.last_modified, f.is_paused,
		       f.max_entries, f.folder_id, f.username, f.password_encrypted,
		       f.created_at, f.updated_at, f.deleted_at
		FROM feeds f
		WHERE f.created_by = $1 AND f.deleted_at IS NULL
	`

	if cursor != "" {
		decoded, err := hex.DecodeString(cursor)
		if err != nil {
			return nil, "", false, fmt.Errorf("decode cursor: %w", err)
		}
		if len(decoded) != 16 {
			return nil, "", false, fmt.Errorf("invalid cursor length: %d", len(decoded))
		}
		query += ` AND f.id < $` + fmt.Sprintf("%d", argIdx)
		args = append(args, decoded)
		argIdx++
	}

	query += ` ORDER BY f.id DESC LIMIT $` + fmt.Sprintf("%d", argIdx)
	args = append(args, limit+1) // fetch one extra to detect has_more

	rows, err := r.db.Query(ctx, query, args...)
	if err != nil {
		return nil, "", false, fmt.Errorf("list feeds: %w", err)
	}
	defer rows.Close()

	var feeds []*models.Feed
	for rows.Next() {
		f := &models.Feed{}
		if err := scanFeedRow(rows, f); err != nil {
			return nil, "", false, fmt.Errorf("scan feed: %w", err)
		}
		feeds = append(feeds, f)
	}
	if err := rows.Err(); err != nil {
		return nil, "", false, fmt.Errorf("rows iteration: %w", err)
	}

	hasMore := len(feeds) > limit
	if hasMore {
		feeds = feeds[:limit]
	}

	nextCursor := ""
	if hasMore && len(feeds) > 0 {
		nextCursor = hex.EncodeToString(feeds[len(feeds)-1].ID[:])
	}

	return feeds, nextCursor, hasMore, nil
}

// UpdateFeed updates mutable feed fields and returns the updated feed.
func (r *FeedRepo) UpdateFeed(ctx context.Context, feed *models.Feed) (*models.Feed, error) {
	feed.UpdatedAt = time.Now()

	var folderIDBytes []byte
	if feed.FolderID != nil {
		folderIDBytes = feed.FolderID[:]
	}

	tag, err := r.db.Exec(ctx, `
		UPDATE feeds SET
			title = $1, description = $2, site_url = $3, icon_url = $4,
			poll_interval = $5, status = $6, is_paused = $7,
			max_entries = $8, folder_id = $9,
			username = $10, password_encrypted = $11,
			updated_at = $12
		WHERE id = $13 AND deleted_at IS NULL
	`, feed.Title, feed.Description, feed.SiteURL, feed.IconURL,
		feed.PollInterval, string(feed.Status), feed.IsPaused,
		feed.MaxEntries, folderIDBytes,
		feed.Username, feed.PasswordEncrypted,
		feed.UpdatedAt, feed.ID[:])
	if err != nil {
		return nil, fmt.Errorf("update feed: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return nil, ErrNotFound
	}
	return feed, nil
}

// DeleteFeed soft-deletes a feed by setting deleted_at.
func (r *FeedRepo) DeleteFeed(ctx context.Context, id models.UUIDv7) error {
	tag, err := r.db.Exec(ctx, `UPDATE feeds SET deleted_at = now() WHERE id = $1 AND deleted_at IS NULL`, id[:])
	if err != nil {
		return fmt.Errorf("delete feed: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return ErrNotFound
	}
	return nil
}

// --- scan helpers ---

// scanFeed returns a single feed matching the WHERE clause suffix.
func scanFeed(ctx context.Context, db interface {
	QueryRow(ctx context.Context, sql string, args ...any) pgx.Row
}, whereClause string, args ...any) (*models.Feed, error) {
	query := `
		SELECT f.id, f.created_by, f.title, f.description, f.feed_url, f.site_url, f.icon_url,
		       f.poll_interval, f.next_poll_at, f.last_fetched_at, f.status,
		       f.etag, f.last_modified, f.is_paused,
		       f.max_entries, f.folder_id, f.username, f.password_encrypted,
		       f.created_at, f.updated_at, f.deleted_at
		FROM feeds f ` + whereClause

	f := &models.Feed{}
	row := db.QueryRow(ctx, query, args...)
	if err := scanFeedRow(row, f); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("scan feed: %w", err)
	}
	return f, nil
}

// feedScanner is satisfied by pgx.Row and pgx.Rows.
type feedScanner interface {
	Scan(dest ...any) error
}

func scanFeedRow(s feedScanner, f *models.Feed) error {
	var (
		folderID    []byte
		statusStr   string
		deletedAt   *time.Time
		lastFetched *time.Time
	)

	err := s.Scan(
		&f.ID, &f.CreatedBy, &f.Title, &f.Description, &f.FeedURL, &f.SiteURL, &f.IconURL,
		&f.PollInterval, &f.NextPollAt, &lastFetched, &statusStr,
		&f.ETag, &f.LastModified, &f.IsPaused,
		&f.MaxEntries, &folderID, &f.Username, &f.PasswordEncrypted,
		&f.CreatedAt, &f.UpdatedAt, &deletedAt,
	)
	if err != nil {
		return err
	}

	f.Status = models.FeedStatus(statusStr)
	f.LastFetchedAt = lastFetched
	f.DeletedAt = deletedAt

	if len(folderID) > 0 {
		uid := models.UUIDv7FromBytes(folderID)
		f.FolderID = &uid
	}

	return nil
}
