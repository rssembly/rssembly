package models

import "time"

// ReadState tracks a single user's read progress for an article.
type ReadState string

const (
	ReadStateUnread   ReadState = "unread"
	ReadStateRead     ReadState = "read"
	ReadStateSaved    ReadState = "saved" // bookmarked
)

// Article represents a single entry from a feed.
type Article struct {
	ID           UUIDv7    `json:"id"`
	FeedID       UUIDv7    `json:"feed_id"`
	GUID         string    `json:"guid"`       // globally unique, from feed or computed
	URL          string    `json:"url"`
	Title        string    `json:"title"`
	Content      string    `json:"content,omitempty"` // full HTML/plain content
	Summary      string    `json:"summary,omitempty"`
	Author       string    `json:"author,omitempty"`
	PublishedAt  *time.Time `json:"published_at,omitempty"`
	UpdatedAt    *time.Time `json:"updated_at,omitempty"`

	// Read state per user (populated when queried with user context).
	ReadState *ReadState `json:"read_state,omitempty"`

	CreatedAt time.Time `json:"created_at"`
}