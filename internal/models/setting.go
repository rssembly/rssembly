package models

import (
	"encoding/json"
	"time"
)

// Setting defines a global configuration option with a default value, type hint,
// and human-readable description. The stored value is JSONB, meaning it can hold
// any JSON-compatible type (string, number, boolean, array, object).
type Setting struct {
	ID           UUIDv7            `json:"id"`
	Key          string            `json:"key"`                    // machine name, e.g. "feed.default_poll_interval"
	DefaultValue json.RawMessage   `json:"default_value"`          // JSONB — any type
	Type         string            `json:"type"`                   // "string", "number", "boolean", "duration", "array"
	Description  string            `json:"description,omitempty"`
	CreatedAt    time.Time         `json:"created_at"`
	UpdatedAt    time.Time         `json:"updated_at"`
}

// UserSetting represents a per-user override of a global setting.
// When present, the user's value takes priority over the global default.
type UserSetting struct {
	UserID    UUIDv7          `json:"user_id"`
	Key       string          `json:"key"`
	Value     json.RawMessage `json:"value"` // JSONB — the user's override
	UpdatedAt time.Time       `json:"updated_at"`
}
