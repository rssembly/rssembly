package models

import "time"

// FeedStatus represents the current health of a feed subscription.
type FeedStatus string

const (
	FeedStatusOK     FeedStatus = "ok"
	FeedStatusError  FeedStatus = "error"
	FeedStatusPaused FeedStatus = "paused"
)

// Feed represents an RSS/Atom feed subscription.
type Feed struct {
	ID        UUIDv7 `json:"id"`
	CreatedBy UUIDv7 `json:"created_by"`
	Title     string `json:"title"`
	Description string   `json:"description,omitempty"`
	FeedURL   string     `json:"feed_url"`    // the RSS/Atom XML URL
	SiteURL   string     `json:"site_url"`    // human-readable website URL
	IconURL   string     `json:"icon_url,omitempty"`

	// Feed authentication (basic auth).
	Username          string `json:"username,omitempty"`
	PasswordEncrypted string `json:"-"` // AES-256-GCM encrypted, never exposed

	// Polling state.
	PollInterval     time.Duration  `json:"poll_interval"`
	NextPollAt       time.Time      `json:"next_poll_at"`
	LastFetchedAt    *time.Time     `json:"last_fetched_at,omitempty"`
	Status           FeedStatus     `json:"status"`
	ETag             string         `json:"etag,omitempty"`
	LastModified     string         `json:"last_modified,omitempty"`
	IsPaused         bool           `json:"is_paused"`

	// Retention.
	MaxEntries int `json:"max_entries"`

	// Folder / organization.
	FolderID  *UUIDv7   `json:"folder_id,omitempty"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	DeletedAt *time.Time `json:"deleted_at,omitempty"`
}