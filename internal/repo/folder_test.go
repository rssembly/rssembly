package repo

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/rssembly/rssembly/internal/models"
)

func TestFolderRepo_CreateAndGet(t *testing.T) {
	db := testDB(t)
	repo := NewFolderRepo(db)
	ctx := context.Background()

	user := seedUser(t, db)

	folder, err := repo.CreateFolder(ctx, &models.Folder{
		UserID:    user.ID,
		Name:      "Tech Blogs",
		SortOrder: 1,
	})
	require.NoError(t, err)
	require.NotEqual(t, models.NilUUIDv7, folder.ID)
	require.Equal(t, "Tech Blogs", folder.Name)

	got, err := repo.GetFolderByID(ctx, folder.ID)
	require.NoError(t, err)
	require.Equal(t, folder.ID, got.ID)
	require.Equal(t, folder.Name, got.Name)
}

func TestFolderRepo_GetNotFound(t *testing.T) {
	db := testDB(t)
	repo := NewFolderRepo(db)
	ctx := context.Background()

	_, err := repo.GetFolderByID(ctx, models.NewUUIDv7())
	require.ErrorIs(t, err, ErrNotFound)
}

func TestFolderRepo_DuplicateName(t *testing.T) {
	db := testDB(t)
	repo := NewFolderRepo(db)
	ctx := context.Background()

	user := seedUser(t, db)

	_, err := repo.CreateFolder(ctx, &models.Folder{
		UserID: user.ID,
		Name:   "Duplicate",
	})
	require.NoError(t, err)

	_, err = repo.CreateFolder(ctx, &models.Folder{
		UserID: user.ID,
		Name:   "Duplicate",
	})
	require.ErrorIs(t, err, ErrConflict)
}

func TestFolderRepo_ListFoldersByUser(t *testing.T) {
	db := testDB(t)
	repo := NewFolderRepo(db)
	ctx := context.Background()

	user := seedUser(t, db)
	other := seedUser(t, db)

	_, err := repo.CreateFolder(ctx, &models.Folder{UserID: user.ID, Name: "News", SortOrder: 2})
	require.NoError(t, err)
	_, err = repo.CreateFolder(ctx, &models.Folder{UserID: user.ID, Name: "Tech", SortOrder: 1})
	require.NoError(t, err)
	_, err = repo.CreateFolder(ctx, &models.Folder{UserID: other.ID, Name: "Other", SortOrder: 1})
	require.NoError(t, err)

	folders, err := repo.ListFoldersByUser(ctx, user.ID)
	require.NoError(t, err)
	require.Len(t, folders, 2)
	require.Equal(t, "Tech", folders[0].Name) // sort_order 1 first
	require.Equal(t, "News", folders[1].Name) // sort_order 2 second
}

func TestFolderRepo_UpdateFolder(t *testing.T) {
	db := testDB(t)
	repo := NewFolderRepo(db)
	ctx := context.Background()

	user := seedUser(t, db)

	folder, err := repo.CreateFolder(ctx, &models.Folder{
		UserID: user.ID,
		Name:   "Old Name",
	})
	require.NoError(t, err)

	folder.Name = "New Name"
	folder.SortOrder = 5

	updated, err := repo.UpdateFolder(ctx, folder)
	require.NoError(t, err)
	require.Equal(t, "New Name", updated.Name)
	require.Equal(t, 5, updated.SortOrder)
}

func TestFolderRepo_DeleteFolder(t *testing.T) {
	db := testDB(t)
	repo := NewFolderRepo(db)
	ctx := context.Background()

	user := seedUser(t, db)

	folder, err := repo.CreateFolder(ctx, &models.Folder{
		UserID: user.ID,
		Name:   "Delete Me",
	})
	require.NoError(t, err)

	err = repo.DeleteFolder(ctx, folder.ID)
	require.NoError(t, err)

	_, err = repo.GetFolderByID(ctx, folder.ID)
	require.ErrorIs(t, err, ErrNotFound)
}

func TestFolderRepo_DeleteNotFound(t *testing.T) {
	db := testDB(t)
	repo := NewFolderRepo(db)
	ctx := context.Background()

	err := repo.DeleteFolder(ctx, models.NewUUIDv7())
	require.ErrorIs(t, err, ErrNotFound)
}

func TestFolderRepo_UpdateNotFound(t *testing.T) {
	db := testDB(t)
	repo := NewFolderRepo(db)
	ctx := context.Background()

	_, err := repo.UpdateFolder(ctx, &models.Folder{
		ID:   models.NewUUIDv7(),
		Name: "Ghost",
	})
	require.ErrorIs(t, err, ErrNotFound)
}
