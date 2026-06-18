// Package settings provides resolution of configuration values across
// global defaults and per-user overrides.
//
// Priority: user override > global default. If neither exists, an error
// is returned so the caller can fall back to a code-level default.
package settings

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
)

// ErrSettingNotFound is returned when a setting key has no global default
// and no user override.
var ErrSettingNotFound = errors.New("setting not found")

// Store defines the data access interface needed by the resolver.
type Store interface {
	GetGlobal(ctx context.Context, key string) (json.RawMessage, error)
	GetUser(ctx context.Context, userID, key string) (json.RawMessage, error)
}

// Resolve returns the effective value for a setting key, checking the user's
// override first, then falling back to the global default. If userID is empty,
// only the global default is checked.
func Resolve(ctx context.Context, store Store, userID, key string) (json.RawMessage, error) {
	// 1. Try user override.
	if userID != "" {
		val, err := store.GetUser(ctx, userID, key)
		if err == nil {
			return val, nil
		}
		if !errors.Is(err, ErrSettingNotFound) {
			return nil, fmt.Errorf("get user setting: %w", err)
		}
	}

	// 2. Fall back to global default.
	val, err := store.GetGlobal(ctx, key)
	if err != nil {
		return nil, fmt.Errorf("get global setting: %w", err)
	}
	return val, nil
}
