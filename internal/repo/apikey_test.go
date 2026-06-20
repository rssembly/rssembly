package repo

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/rssembly/rssembly/internal/models"
)

func TestAPIKeyRepo_CreateAndGet(t *testing.T) {
	db := testDB(t)
	repo := NewAPIKeyRepo(db)
	ctx := context.Background()

	user := seedUser(t, db)

	key := &models.APIKey{
		CreatedBy: user.ID,
		Name:      "CI Deploy Key",
		Prefix:    "abc12345",
		Hash:      "argon2hashvalue",
		Scopes:    []string{"feeds:read", "articles:read"},
		IsActive:  true,
	}

	created, err := repo.CreateAPIKey(ctx, key)
	require.NoError(t, err)
	require.NotEqual(t, models.NilUUIDv7, created.ID)

	got, err := repo.GetAPIKeyByID(ctx, created.ID)
	require.NoError(t, err)
	require.Equal(t, created.Name, got.Name)
	require.Equal(t, created.Hash, got.Hash)
}

func TestAPIKeyRepo_GetNotFound(t *testing.T) {
	db := testDB(t)
	repo := NewAPIKeyRepo(db)
	ctx := context.Background()

	_, err := repo.GetAPIKeyByID(ctx, models.NewUUIDv7())
	require.ErrorIs(t, err, ErrNotFound)
}

func TestAPIKeyRepo_ListByUser(t *testing.T) {
	db := testDB(t)
	repo := NewAPIKeyRepo(db)
	ctx := context.Background()

	user := seedUser(t, db)

	_, err := repo.CreateAPIKey(ctx, &models.APIKey{
		CreatedBy: user.ID,
		Name:      "Key 1",
		Prefix:    "pref1111",
		Hash:      "hash1",
		Scopes:    []string{},
		IsActive:  true,
	})
	require.NoError(t, err)

	_, err = repo.CreateAPIKey(ctx, &models.APIKey{
		CreatedBy: user.ID,
		Name:      "Key 2",
		Prefix:    "pref2222",
		Hash:      "hash2",
		Scopes:    []string{},
		IsActive:  true,
	})
	require.NoError(t, err)

	keys, err := repo.ListAPIKeysByUser(ctx, user.ID)
	require.NoError(t, err)
	require.Len(t, keys, 2)
}

func TestAPIKeyRepo_GetByPrefix(t *testing.T) {
	db := testDB(t)
	repo := NewAPIKeyRepo(db)
	ctx := context.Background()

	user := seedUser(t, db)

	_, err := repo.CreateAPIKey(ctx, &models.APIKey{
		CreatedBy: user.ID,
		Name:      "My Key",
		Prefix:    "mykey123",
		Hash:      "hashvalue",
		Scopes:    []string{},
		IsActive:  true,
	})
	require.NoError(t, err)

	got, err := repo.GetAPIKeyByPrefix(ctx, "mykey123")
	require.NoError(t, err)
	require.Equal(t, "My Key", got.Name)
}

func TestAPIKeyRepo_GetByPrefixNotFound(t *testing.T) {
	db := testDB(t)
	repo := NewAPIKeyRepo(db)
	ctx := context.Background()

	_, err := repo.GetAPIKeyByPrefix(ctx, "nope1234")
	require.ErrorIs(t, err, ErrNotFound)
}

func TestAPIKeyRepo_Revoke(t *testing.T) {
	db := testDB(t)
	repo := NewAPIKeyRepo(db)
	ctx := context.Background()

	user := seedUser(t, db)

	key, err := repo.CreateAPIKey(ctx, &models.APIKey{
		CreatedBy: user.ID,
		Name:      "Revoke Me",
		Prefix:    "revok123",
		Hash:      "hash",
		Scopes:    []string{},
		IsActive:  true,
	})
	require.NoError(t, err)

	err = repo.RevokeAPIKey(ctx, key.ID)
	require.NoError(t, err)

	// Should not be findable by ID
	_, err = repo.GetAPIKeyByID(ctx, key.ID)
	require.ErrorIs(t, err, ErrNotFound)

	// Should not be findable by prefix (is_active = false)
	_, err = repo.GetAPIKeyByPrefix(ctx, "revok123")
	require.ErrorIs(t, err, ErrNotFound)
}
