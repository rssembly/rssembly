// Package models provides the domain types for Rssembly.
//
// UUIDv7: we re-export github.com/google/uuid with convenience helpers.
package models

import (
	"github.com/google/uuid"
)

// UUIDv7 is a UUID conforming to RFC 9562 (UUIDv7), time-ordered for index-friendly
// primary keys. It is an alias for github.com/google/uuid.UUID.
type UUIDv7 = uuid.UUID

// NewUUIDv7 generates a new UUIDv7 from the current timestamp and random bits.
func NewUUIDv7() UUIDv7 {
	return uuid.Must(uuid.NewV7())
}

// ParseUUIDv7 parses a UUID string (standard 36-char with hyphens, or 32-char hex-only).
func ParseUUIDv7(s string) (UUIDv7, error) {
	return uuid.Parse(s)
}

// NilUUIDv7 is the zero-value UUID (all zeros).
var NilUUIDv7 = uuid.Nil

// UUIDv7FromBytes creates a UUIDv7 from a 16-byte slice (e.g. from pgx scanning BYTEA).
// Panics if the slice is not exactly 16 bytes.
func UUIDv7FromBytes(b []byte) UUIDv7 {
	return uuid.UUID(b)
}