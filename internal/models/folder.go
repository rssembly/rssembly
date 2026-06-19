package models

import "time"

// Folder represents a user-created folder for organizing feeds.
type Folder struct {
	ID        UUIDv7     `json:"id"`
	UserID    UUIDv7     `json:"user_id"`
	Name      string     `json:"name"`
	ParentID  *UUIDv7    `json:"parent_id,omitempty"`
	SortOrder int        `json:"sort_order"`
	CreatedAt time.Time  `json:"created_at"`
	UpdatedAt time.Time  `json:"updated_at"`
	DeletedAt *time.Time `json:"deleted_at,omitempty"`
}