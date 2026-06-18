package auth

import (
	"crypto/ed25519"
	"crypto/rand"
	"crypto/x509"
	"encoding/pem"
	"testing"

	"github.com/google/uuid"
)

func TestNewJWTManagerFromPEM_ValidKeys(t *testing.T) {
	pub, priv, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		t.Fatalf("generate key: %v", err)
	}

	privBytes, _ := x509.MarshalPKCS8PrivateKey(priv)
	pubBytes, _ := x509.MarshalPKIXPublicKey(pub)

	privPEM := pem.EncodeToMemory(&pem.Block{Type: "PRIVATE KEY", Bytes: privBytes})
	pubPEM := pem.EncodeToMemory(&pem.Block{Type: "PUBLIC KEY", Bytes: pubBytes})

	mgr, err := NewJWTManagerFromPEM(privPEM, pubPEM)
	if err != nil {
		t.Fatalf("NewJWTManagerFromPEM() error: %v", err)
	}
	if mgr == nil {
		t.Fatal("expected non-nil manager")
	}

	// Round-trip: generate a token and verify it.
	token, err := mgr.GenerateToken(uuid.Must(uuid.NewV7()), nil, DefaultTokenExpiry)
	if err != nil {
		t.Fatalf("GenerateToken() error: %v", err)
	}

	user, err := mgr.VerifyToken(token)
	if err != nil {
		t.Fatalf("VerifyToken() error: %v", err)
	}
	if user == nil {
		t.Fatal("expected non-nil user")
	}
}

func TestNewJWTManagerFromPEM_InvalidPrivateKey(t *testing.T) {
	_, err := NewJWTManagerFromPEM(
		[]byte("invalid pem"),
		[]byte("invalid pem"),
	)
	if err == nil {
		t.Fatal("expected error for invalid PEM")
	}
}

func TestNewJWTManagerFromPEM_MismatchedKeys(t *testing.T) {
	_, priv1, _ := ed25519.GenerateKey(rand.Reader)
	pub2, _, _ := ed25519.GenerateKey(rand.Reader)

	privBytes, _ := x509.MarshalPKCS8PrivateKey(priv1)
	pubBytes, _ := x509.MarshalPKIXPublicKey(pub2)

	privPEM := pem.EncodeToMemory(&pem.Block{Type: "PRIVATE KEY", Bytes: privBytes})
	pubPEM := pem.EncodeToMemory(&pem.Block{Type: "PUBLIC KEY", Bytes: pubBytes})

	mgr, err := NewJWTManagerFromPEM(privPEM, pubPEM)
	if err != nil {
		t.Fatalf("NewJWTManagerFromPEM() error: %v", err)
	}

	token, err := mgr.GenerateToken(uuid.Must(uuid.NewV7()), nil, DefaultTokenExpiry)
	if err != nil {
		t.Fatalf("GenerateToken() error: %v", err)
	}

	_, err = mgr.VerifyToken(token)
	if err == nil {
		t.Fatal("expected verification error for mismatched keys")
	}
}

func TestDefaultTokenExpiry(t *testing.T) {
	if DefaultTokenExpiry <= 0 {
		t.Fatal("DefaultTokenExpiry must be positive")
	}
}

func TestGenerateKeyPairPEM_ProducesValidKeys(t *testing.T) {
	privPEM, pubPEM, err := GenerateKeyPairPEM()
	if err != nil {
		t.Fatalf("GenerateKeyPairPEM() error: %v", err)
	}
	if len(privPEM) == 0 || len(pubPEM) == 0 {
		t.Fatal("expected non-empty PEM bytes")
	}

	mgr, err := NewJWTManagerFromPEM(privPEM, pubPEM)
	if err != nil {
		t.Fatalf("NewJWTManagerFromPEM() error: %v", err)
	}

	token, err := mgr.GenerateToken(uuid.Must(uuid.NewV7()), nil, DefaultTokenExpiry)
	if err != nil {
		t.Fatalf("GenerateToken() error: %v", err)
	}

	user, err := mgr.VerifyToken(token)
	if err != nil {
		t.Fatalf("VerifyToken() error: %v", err)
	}
	if user == nil {
		t.Fatal("expected non-nil user")
	}
}