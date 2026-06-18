package auth

import (
	"testing"
)

func TestEncryptDecrypt_RoundTrip(t *testing.T) {
	key := "abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789" // 64 hex chars = 32 bytes
	plaintext := "my-feed-password"

	ciphertext, err := Encrypt(plaintext, key)
	if err != nil {
		t.Fatalf("Encrypt() error: %v", err)
	}
	if ciphertext == "" {
		t.Fatal("expected non-empty ciphertext")
	}
	if ciphertext == plaintext {
		t.Fatal("ciphertext should not equal plaintext")
	}

	decrypted, err := Decrypt(ciphertext, key)
	if err != nil {
		t.Fatalf("Decrypt() error: %v", err)
	}
	if decrypted != plaintext {
		t.Errorf("expected %q, got %q", plaintext, decrypted)
	}
}

func TestEncrypt_DifferentCiphertexts(t *testing.T) {
	key := "abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789"
	c1, _ := Encrypt("password", key)
	c2, _ := Encrypt("password", key)
	if c1 == c2 {
		t.Error("same plaintext should produce different ciphertexts (random nonce)")
	}
}

func TestEncrypt_InvalidKeyLength(t *testing.T) {
	_, err := Encrypt("password", "tooshort")
	if err == nil {
		t.Error("expected error for short key")
	}
}

func TestDecrypt_TamperedCiphertext(t *testing.T) {
	key := "abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789"
	ciphertext, _ := Encrypt("password", key)

	// Tamper with the ciphertext.
	tampered := ciphertext[:len(ciphertext)-1] + "x"
	_, err := Decrypt(tampered, key)
	if err == nil {
		t.Error("expected error for tampered ciphertext")
	}
}

func TestDecrypt_WrongKey(t *testing.T) {
	key1 := "abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789"
	key2 := "0000000000000000000000000000000000000000000000000000000000000000"
	ciphertext, _ := Encrypt("password", key1)
	_, err := Decrypt(ciphertext, key2)
	if err == nil {
		t.Error("expected error for wrong key")
	}
}
