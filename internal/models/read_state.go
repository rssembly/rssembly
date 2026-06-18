package models

import "time"

// ReadState tracks a single user's read progress for an article.
type ReadState string

const (
	ReadStateUnread ReadState = "unread"
	ReadStateRead   ReadState = "read"
	ReadStateSaved  ReadState = "saved"
)

// ArticleReadState represents the join between a user and an article
// tracking read progress.
type ArticleReadState struct {
	UserID    UUIDv7    `json:"user_id"`
	ArticleID UUIDv7    `json:"article_id"`
	State     ReadState `json:"state"`
	UpdatedAt time.Time `json:"updated_at"`
}