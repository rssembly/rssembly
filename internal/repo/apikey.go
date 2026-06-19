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

// APIKeyRepo handles API key database operations.
type APIKeyRepo struct {
	db *database.Pool
}

// NewAPIKeyRepo creates an APIKeyRepo.
func NewAPIKeyRepo(db *database.Pool) *APIKeyRepo {
	return &APIKeyRepo{db: db}
}

// CreateAPIKey inserts a new API key and returns it.
func (r *APIKeyRepo) CreateAPIKey(ctx context.Context, key *models.APIKey) (*models.APIKey, error) {
	key.ID = models.NewUUIDv7()
	now := time.Now()
	key.CreatedAt = now
	key.UpdatedAt = now

	_, err := r.db.Exec(ctx, `
		INSERT INTO api_keys
			(id, created_by, name, prefix, hash, scopes, last_used_at, expires_at, is_active, created_at, updated_at)
		VALUES
			($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
	`, key.ID[:], key.CreatedBy[:], key.Name, key.Prefix, key.Hash, key.Scopes, key.LastUsedAt, key.ExpiresAt, key.IsActive, key.CreatedAt, key.UpdatedAt)

	if err != nil {
		return nil, fmt.Errorf("create api key: %w", err)
	}

	return key, nil
}

// GetAPIKeyByID returns an API key by its primary key.
func (r *APIKeyRepo) GetAPIKeyByID(ctx context.Context, id models.UUIDv7) (*models.APIKey, error) {
	k := &models.APIKey{}

	err := r.db.QueryRow(ctx, `
		SELECT id, created_by, name, prefix, hash, scopes, last_used_at, expires_at, is_active, created_at, updated_at, deleted_at
		FROM api_keys
		WHERE id = $1 AND deleted_at IS NULL
	`, id[:]).Scan(&k.ID, &k.CreatedBy, &k.Name, &k.Prefix, &k.Hash, &k.Scopes, &k.LastUsedAt, &k.ExpiresAt, &k.IsActive, &k.CreatedAt, &k.UpdatedAt, &k.DeletedAt)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("get api key: %w", err)
	}

	return k, nil
}

// ListAPIKeysByUser returns all active (non-deleted) API keys for a user.
func (r *APIKeyRepo) ListAPIKeysByUser(ctx context.Context, userID models.UUIDv7) ([]*models.APIKey, error) {
	rows, err := r.db.Query(ctx, `
		SELECT id, created_by, name, prefix, hash, scopes, last_used_at, expires_at, is_active, created_at, updated_at, deleted_at
		FROM api_keys
		WHERE created_by = $1 AND deleted_at IS NULL
		ORDER BY created_at DESC
	`, userID[:])
	if err != nil {
		return nil, fmt.Errorf("list api keys: %w", err)
	}
	defer rows.Close()

	var keys []*models.APIKey
	for rows.Next() {
		k := &models.APIKey{}
		if err := rows.Scan(&k.ID, &k.CreatedBy, &k.Name, &k.Prefix, &k.Hash, &k.Scopes, &k.LastUsedAt, &k.ExpiresAt, &k.IsActive, &k.CreatedAt, &k.UpdatedAt, &k.DeletedAt); err != nil {
			return nil, fmt.Errorf("scan api key: %w", err)
		}
		keys = append(keys, k)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("rows iteration: %w", err)
	}

	return keys, nil
}

// RevokeAPIKey soft-deletes an API key by setting deleted_at.
func (r *APIKeyRepo) RevokeAPIKey(ctx context.Context, id models.UUIDv7) error {
	tag, err := r.db.Exec(ctx, `UPDATE api_keys SET deleted_at = now(), is_active = false WHERE id = $1 AND deleted_at IS NULL`, id[:])
	if err != nil {
		return fmt.Errorf("revoke api key: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return ErrNotFound
	}
	return nil
}

// GetAPIKeyByPrefix looks up an API key by its prefix (for authentication flow).
// Returns ErrNotFound if no active key matches.
func (r *APIKeyRepo) GetAPIKeyByPrefix(ctx context.Context, prefix string) (*models.APIKey, error) {
	k := &models.APIKey{}

	err := r.db.QueryRow(ctx, `
		SELECT id, created_by, name, prefix, hash, scopes, last_used_at, expires_at, is_active, created_at, updated_at, deleted_at
		FROM api_keys
		WHERE prefix = $1 AND deleted_at IS NULL AND is_active = true
	`, prefix).Scan(&k.ID, &k.CreatedBy, &k.Name, &k.Prefix, &k.Hash, &k.Scopes, &k.LastUsedAt, &k.ExpiresAt, &k.IsActive, &k.CreatedAt, &k.UpdatedAt, &k.DeletedAt)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("get api key by prefix: %w", err)
	}

	return k, nil
}