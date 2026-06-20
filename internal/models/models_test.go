package models

import (
	"encoding/json"
	"testing"
	"time"
)

func TestUser_JSONRoundTrip(t *testing.T) {
	u := User{
		ID:           NewUUIDv7(),
		Username:     "testuser",
		Email:        "test@example.com",
		PasswordHash: "argon2hash",
		Scopes:       []string{"feeds:read"},
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
		DeletedAt:    nil,
	}
	data, err := json.Marshal(u)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}

	var decoded User
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if decoded.Email != u.Email {
		t.Errorf("expected email %q, got %q", u.Email, decoded.Email)
	}
	if decoded.PasswordHash != "" {
		t.Error("PasswordHash should be omitted from JSON")
	}
	if len(decoded.Scopes) != 1 || decoded.Scopes[0] != "feeds:read" {
		t.Errorf("expected scopes [feeds:read], got %v", decoded.Scopes)
	}
}

func TestAPIKey_JSONOmitsHash(t *testing.T) {
	k := APIKey{
		ID:        NewUUIDv7(),
		Name:      "test-key",
		Prefix:    "abc12345",
		Hash:      "argon2hashvalue",
		IsActive:  true,
		CreatedAt: time.Now(),
	}
	data, err := json.Marshal(k)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}

	var decoded map[string]any
	json.Unmarshal(data, &decoded)
	if _, ok := decoded["hash"]; ok {
		t.Error("hash field should not be in JSON output")
	}
}

func TestReadState_Values(t *testing.T) {
	if ReadStateUnread != "unread" {
		t.Errorf("expected unread, got %q", ReadStateUnread)
	}
	if ReadStateRead != "read" {
		t.Errorf("expected read, got %q", ReadStateRead)
	}
	if ReadStateSaved != "saved" {
		t.Errorf("expected saved, got %q", ReadStateSaved)
	}
}

func TestFolder_JSONRoundTrip(t *testing.T) {
	f := Folder{
		ID:        NewUUIDv7(),
		UserID:    NewUUIDv7(),
		Name:      "Tech Blogs",
		SortOrder: 1,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	data, err := json.Marshal(f)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}

	var decoded Folder
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if decoded.Name != f.Name {
		t.Errorf("expected name %q, got %q", f.Name, decoded.Name)
	}
	if decoded.SortOrder != 1 {
		t.Errorf("expected sort_order 1, got %d", decoded.SortOrder)
	}
}
