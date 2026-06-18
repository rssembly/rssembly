package models

import "time"

// APIKey represents a machine-to-machine API key with granular scoping.
type APIKey struct {
	ID        UUIDv7     `json:"id"`
	CreatedBy UUIDv7     `json:"created_by"`
	Name      string     `json:"name"`   // human label, e.g. "CI deploy"
	Prefix    string     `json:"prefix"` // first 8 chars of the key, for identification
	Hash      string     `json:"-"`      // argon2id hash of the full key
	Scopes    []string   `json:"scopes"` // e.g. ["feeds:read", "articles:*"]
	LastUsedAt *time.Time `json:"last_used_at,omitempty"`
	ExpiresAt *time.Time `json:"expires_at,omitempty"`
	IsActive  bool       `json:"is_active"`
	CreatedAt time.Time  `json:"created_at"`
	UpdatedAt time.Time  `json:"updated_at"`
	DeletedAt *time.Time `json:"deleted_at,omitempty"`
}