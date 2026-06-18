package models

import "time"

// User represents a registered user of the system.
type User struct {
	ID           UUIDv7     `json:"id"`
	Username     string     `json:"username"`
	Email        string     `json:"email"`
	PasswordHash string     `json:"-"` // never exposed in API responses
	Scopes       []string   `json:"scopes,omitempty"` // e.g. ["*"] for superadmin
	CreatedAt    time.Time  `json:"created_at"`
	UpdatedAt    time.Time  `json:"updated_at"`
	DeletedAt    *time.Time `json:"deleted_at,omitempty"`
}
