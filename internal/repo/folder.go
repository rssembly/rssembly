package repo

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"

	"github.com/rssembly/rssembly/internal/database"
	"github.com/rssembly/rssembly/internal/models"
)

// FolderRepo handles folder CRUD database operations.
type FolderRepo struct {
	db *database.Pool
}

// NewFolderRepo creates a FolderRepo.
func NewFolderRepo(db *database.Pool) *FolderRepo {
	return &FolderRepo{db: db}
}

// CreateFolder inserts a new folder. Returns ErrConflict if the user already
// has a folder with the same name (unique per user_id + name).
func (r *FolderRepo) CreateFolder(ctx context.Context, folder *models.Folder) (*models.Folder, error) {
	folder.ID = models.NewUUIDv7()
	now := time.Now()
	folder.CreatedAt = now
	folder.UpdatedAt = now

	var parentIDBytes []byte
	if folder.ParentID != nil {
		parentIDBytes = folder.ParentID[:]
	}

	_, err := r.db.Exec(ctx, `
		INSERT INTO folders (id, user_id, name, parent_id, sort_order, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
	`, folder.ID[:], folder.UserID[:], folder.Name, parentIDBytes, folder.SortOrder, folder.CreatedAt, folder.UpdatedAt)

	if err != nil {
		if isUniqueViolation(err) {
			return nil, fmt.Errorf("create folder: %w", ErrConflict)
		}
		return nil, fmt.Errorf("create folder: %w", err)
	}

	return folder, nil
}

// GetFolderByID returns a single folder by its primary key.
func (r *FolderRepo) GetFolderByID(ctx context.Context, id models.UUIDv7) (*models.Folder, error) {
	f := &models.Folder{}
	var parentIDBytes []byte

	err := r.db.QueryRow(ctx, `
		SELECT id, user_id, name, parent_id, sort_order, created_at, updated_at, deleted_at
		FROM folders
		WHERE id = $1 AND deleted_at IS NULL
	`, id[:]).Scan(&f.ID, &f.UserID, &f.Name, &parentIDBytes, &f.SortOrder, &f.CreatedAt, &f.UpdatedAt, &f.DeletedAt)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("get folder: %w", err)
	}

	if len(parentIDBytes) > 0 {
		uid := models.UUIDv7FromBytes(parentIDBytes)
		f.ParentID = &uid
	}

	return f, nil
}

// ListFoldersByUser returns all non-deleted folders for a user, ordered by sort_order then name.
func (r *FolderRepo) ListFoldersByUser(ctx context.Context, userID models.UUIDv7) ([]*models.Folder, error) {
	rows, err := r.db.Query(ctx, `
		SELECT id, user_id, name, parent_id, sort_order, created_at, updated_at, deleted_at
		FROM folders
		WHERE user_id = $1 AND deleted_at IS NULL
		ORDER BY sort_order ASC, name ASC
	`, userID[:])
	if err != nil {
		return nil, fmt.Errorf("list folders: %w", err)
	}
	defer rows.Close()

	var folders []*models.Folder
	for rows.Next() {
		f := &models.Folder{}
		var parentIDBytes []byte
		if err := rows.Scan(&f.ID, &f.UserID, &f.Name, &parentIDBytes, &f.SortOrder, &f.CreatedAt, &f.UpdatedAt, &f.DeletedAt); err != nil {
			return nil, fmt.Errorf("scan folder: %w", err)
		}
		if len(parentIDBytes) > 0 {
			uid := models.UUIDv7FromBytes(parentIDBytes)
			f.ParentID = &uid
		}
		folders = append(folders, f)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("rows iteration: %w", err)
	}

	return folders, nil
}

// UpdateFolder updates mutable folder fields and returns the updated folder.
func (r *FolderRepo) UpdateFolder(ctx context.Context, folder *models.Folder) (*models.Folder, error) {
	folder.UpdatedAt = time.Now()

	var parentIDBytes []byte
	if folder.ParentID != nil {
		parentIDBytes = folder.ParentID[:]
	}

	tag, err := r.db.Exec(ctx, `
		UPDATE folders SET
			name = $1, parent_id = $2, sort_order = $3, updated_at = $4
		WHERE id = $5 AND deleted_at IS NULL
	`, folder.Name, parentIDBytes, folder.SortOrder, folder.UpdatedAt, folder.ID[:])
	if err != nil {
		if isUniqueViolation(err) {
			return nil, fmt.Errorf("update folder: %w", ErrConflict)
		}
		return nil, fmt.Errorf("update folder: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return nil, ErrNotFound
	}
	return folder, nil
}

// DeleteFolder soft-deletes a folder by setting deleted_at.
func (r *FolderRepo) DeleteFolder(ctx context.Context, id models.UUIDv7) error {
	tag, err := r.db.Exec(ctx, `UPDATE folders SET deleted_at = now() WHERE id = $1 AND deleted_at IS NULL`, id[:])
	if err != nil {
		return fmt.Errorf("delete folder: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return ErrNotFound
	}
	return nil
}
