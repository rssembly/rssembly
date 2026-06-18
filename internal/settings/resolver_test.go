package settings

import (
	"context"
	"encoding/json"
	"testing"
)

func TestResolve_UserOverrideWins(t *testing.T) {
	store := &mockStore{
		global: map[string]string{
			"feed.poll_interval": `"15m"`,
		},
		user: map[string]map[string]string{
			"user-1": {"feed.poll_interval": `"30m"`},
		},
	}

	val, err := Resolve(context.Background(), store, "user-1", "feed.poll_interval")
	if err != nil {
		t.Fatalf("Resolve() error: %v", err)
	}
	if string(val) != `"30m"` {
		t.Errorf("expected \"30m\", got %s", string(val))
	}
}

func TestResolve_GlobalDefaultWhenNoUserOverride(t *testing.T) {
	store := &mockStore{
		global: map[string]string{
			"feed.poll_interval": `"15m"`,
		},
		user: map[string]map[string]string{},
	}

	val, err := Resolve(context.Background(), store, "user-1", "feed.poll_interval")
	if err != nil {
		t.Fatalf("Resolve() error: %v", err)
	}
	if string(val) != `"15m"` {
		t.Errorf("expected \"15m\", got %s", string(val))
	}
}

func TestResolve_UnknownKey(t *testing.T) {
	store := &mockStore{
		global: map[string]string{},
		user:   map[string]map[string]string{},
	}

	_, err := Resolve(context.Background(), store, "user-1", "nonexistent.key")
	if err == nil {
		t.Fatal("expected error for unknown key")
	}
}

func TestResolve_NoUser(t *testing.T) {
	store := &mockStore{
		global: map[string]string{
			"system.max_feeds": `"100"`,
		},
		user: map[string]map[string]string{},
	}

	val, err := Resolve(context.Background(), store, "", "system.max_feeds")
	if err != nil {
		t.Fatalf("Resolve() error: %v", err)
	}
	if string(val) != `"100"` {
		t.Errorf("expected \"100\", got %s", string(val))
	}
}

type mockStore struct {
	global map[string]string
	user   map[string]map[string]string
}

func (m *mockStore) GetGlobal(_ context.Context, key string) (json.RawMessage, error) {
	v, ok := m.global[key]
	if !ok {
		return nil, ErrSettingNotFound
	}
	return json.RawMessage(v), nil
}

func (m *mockStore) GetUser(_ context.Context, userID, key string) (json.RawMessage, error) {
	if userID == "" {
		return nil, ErrSettingNotFound
	}
	userSettings, ok := m.user[userID]
	if !ok {
		return nil, ErrSettingNotFound
	}
	v, ok := userSettings[key]
	if !ok {
		return nil, ErrSettingNotFound
	}
	return json.RawMessage(v), nil
}
