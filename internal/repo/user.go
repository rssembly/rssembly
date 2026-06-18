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

// ErrNotFound is returned when a query returns zero rows.
var ErrNotFound = errors.New("not found")

// UserRepo handles user database operations.
type UserRepo struct {
	db *database.Pool
}

// NewUserRepo creates a UserRepo.
func NewUserRepo(db *database.Pool) *UserRepo {
	return &UserRepo{db: db}
}

// CreateUser inserts a new user and returns the created User with its generated ID.
func (r *UserRepo) CreateUser(ctx context.Context, user *models.User) (*models.User, error) {
	user.ID = models.NewUUIDv7()
	user.CreatedAt = time.Now()
	user.UpdatedAt = time.Now()

	_, err := r.db.Exec(ctx, `
		INSERT INTO users (id, username, email, password_hash, scopes, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
	`, user.ID[:], user.Username, user.Email, user.PasswordHash, user.Scopes, user.CreatedAt, user.UpdatedAt)

	if err != nil {
		return nil, fmt.Errorf("create user: %w", err)
	}

	return user, nil
}

// GetUserByEmail looks up a user by their email address.
// Returns ErrNotFound if no user matches.
func (r *UserRepo) GetUserByEmail(ctx context.Context, email string) (*models.User, error) {
	user := &models.User{}
	var scopeSlice []string

	err := r.db.QueryRow(ctx, `
		SELECT id, username, email, password_hash, scopes, created_at, updated_at, deleted_at
		FROM users
		WHERE email = $1 AND deleted_at IS NULL
	`, email).Scan(&user.ID, &user.Username, &user.Email, &user.PasswordHash, &scopeSlice, &user.CreatedAt, &user.UpdatedAt, &user.DeletedAt)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("get user by email: %w", err)
	}

	user.Scopes = scopeSlice
	return user, nil
}

// GetUserByID looks up a user by their primary key.
func (r *UserRepo) GetUserByID(ctx context.Context, id models.UUIDv7) (*models.User, error) {
	user := &models.User{}
	var scopeSlice []string

	err := r.db.QueryRow(ctx, `
		SELECT id, username, email, password_hash, scopes, created_at, updated_at, deleted_at
		FROM users
		WHERE id = $1 AND deleted_at IS NULL
	`, id[:]).Scan(&user.ID, &user.Username, &user.Email, &user.PasswordHash, &scopeSlice, &user.CreatedAt, &user.UpdatedAt, &user.DeletedAt)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("get user by id: %w", err)
	}

	user.Scopes = scopeSlice
	return user, nil
}
