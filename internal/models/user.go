package models

import "time"

// User represents a registered user of the system.
type User struct {
	ID           UUIDv7    `json:"id"`
	Email        string    `json:"email"`
	PasswordHash string    `json:"-"` // never exposed in API responses
	DisplayName  string    `json:"display_name"`
	IsAdmin      bool      `json:"is_admin"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}