package models

import "time"

// APIKey represents a machine-to-machine API key.
type APIKey struct {
	ID        UUIDv7    `json:"id"`
	UserID    UUIDv7    `json:"user_id"`
	Name      string    `json:"name"`       // human label, e.g. "CI deploy"
	Prefix    string    `json:"prefix"`     // first 8 chars of the key, for identification
	Hash      string    `json:"-"`          // argon2id hash of the full key
	LastUsedAt *time.Time `json:"last_used_at,omitempty"`
	ExpiresAt *time.Time `json:"expires_at,omitempty"`
	IsActive  bool      `json:"is_active"`
	CreatedAt time.Time `json:"created_at"`
}