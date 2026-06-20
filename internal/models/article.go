package models

import "time"

// Article represents a single entry from a feed.
type Article struct {
	ID          UUIDv7     `json:"id"`
	FeedID      UUIDv7     `json:"feed_id"`
	GUID        string     `json:"guid"` // feed-provided unique ID, used for dedup
	URL         string     `json:"url"`
	Title       string     `json:"title"`
	Content     string     `json:"content,omitempty"` // full HTML/plain content
	PublishedAt *time.Time `json:"published_at,omitempty"`

	// Read state per user (populated when queried with user context).
	ReadState *ReadState `json:"read_state,omitempty"`

	CreatedAt time.Time  `json:"created_at"`
	DeletedAt *time.Time `json:"deleted_at,omitempty"`
}
