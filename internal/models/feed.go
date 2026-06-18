package models

import "time"

// Feed represents an RSS/Atom feed subscription.
type Feed struct {
	ID          UUIDv7    `json:"id"`
	UserID      UUIDv7    `json:"user_id"`
	Title       string    `json:"title"`
	Description string    `json:"description,omitempty"`
	FeedURL     string    `json:"feed_url"`    // the RSS/Atom XML URL
	SiteURL     string    `json:"site_url"`    // the human-readable website URL
	IconURL     string    `json:"icon_url,omitempty"`

	// Polling state.
	PollInterval      time.Duration `json:"poll_interval"`
	NextPollAt        time.Time     `json:"next_poll_at"`
	LastFetchedAt     *time.Time    `json:"last_fetched_at,omitempty"`
	ConsecutiveFailures int         `json:"consecutive_failures"`
	ETag              string        `json:"etag,omitempty"`
	LastModified      string        `json:"last_modified,omitempty"`
	IsPaused          bool          `json:"is_paused"`

	// Folder / organization.
	FolderID  *UUIDv7   `json:"folder_id,omitempty"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}