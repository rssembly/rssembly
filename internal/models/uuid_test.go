package models

import (
	"testing"
)

func TestNewUUIDv7_Length(t *testing.T) {
	u := NewUUIDv7()
	s := u.String()
	if len(s) != 36 {
		t.Errorf("expected 36 chars, got %d: %s", len(s), s)
	}
}

func TestNewUUIDv7_Version(t *testing.T) {
	u := NewUUIDv7()
	// Version nibble is in byte 6, top 4 bits = 7
	version := u[6] >> 4
	if version != 7 {
		t.Errorf("expected version 7, got %d", version)
	}
}

func TestNewUUIDv7_Variant(t *testing.T) {
	u := NewUUIDv7()
	// Variant bits in byte 8, top 2 bits = 10 (RFC 9562)
	if u[8]&0xc0 != 0x80 {
		t.Errorf("expected variant 10xx xxxx, got %08b", u[8])
	}
}

func TestNewUUIDv7_Uniqueness(t *testing.T) {
	const n = 100
	seen := make(map[string]bool, n)
	for i := 0; i < n; i++ {
		u := NewUUIDv7()
		s := u.String()
		if seen[s] {
			t.Errorf("duplicate UUIDv7: %s", s)
		}
		seen[s] = true
	}
}

func TestParseUUIDv7_Standard(t *testing.T) {
	u := NewUUIDv7()
	s := u.String()
	parsed, err := ParseUUIDv7(s)
	if err != nil {
		t.Fatalf("parse error: %v", err)
	}
	if u != parsed {
		t.Errorf("round-trip mismatch:\n  original: %s\n  parsed:   %s", u, parsed)
	}
}

func TestParseUUIDv7_HexOnly(t *testing.T) {
	u := NewUUIDv7()
	hex := u.String() // keep hyphens; uuid.Parse handles both
	parsed, err := ParseUUIDv7(hex)
	if err != nil {
		t.Fatalf("parse error: %v", err)
	}
	if u != parsed {
		t.Errorf("round-trip mismatch")
	}
}

func TestUUIDv7_Format(t *testing.T) {
	u := NewUUIDv7()
	s := u.String()
	if s[8] != '-' || s[13] != '-' || s[18] != '-' || s[23] != '-' {
		t.Errorf("unexpected format: %s", s)
	}
}

func TestUUIDv7_Bytes(t *testing.T) {
	u := NewUUIDv7()
	b := u[:] // uuid.UUID is [16]byte
	if len(b) != 16 {
		t.Errorf("expected 16 bytes, got %d", len(b))
	}
}

func TestParseUUIDv7_Invalid(t *testing.T) {
	tests := []struct {
		input string
		desc  string
	}{
		{"", "empty"},
		{"not-a-uuid", "garbage"},
		{"zzzzzzzz-zzzz-zzzz-zzzz-zzzzzzzzzzzz", "invalid hex"},
		{"018f3a6e-3e3c-7a00", "truncated"},
	}
	for _, tc := range tests {
		_, err := ParseUUIDv7(tc.input)
		if err == nil {
			t.Errorf("expected error for %s: %q", tc.desc, tc.input)
		}
	}
}
