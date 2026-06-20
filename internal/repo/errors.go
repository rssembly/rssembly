package repo

import (
	"errors"

	"github.com/jackc/pgx/v5/pgconn"
)

// ErrConflict is returned when an insert violates a uniqueness constraint.
var ErrConflict = errors.New("conflict")

// isUniqueViolation checks if a pgx error is a PostgreSQL unique violation (code 23505).
func isUniqueViolation(err error) bool {
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		return pgErr.Code == "23505"
	}
	return false
}
