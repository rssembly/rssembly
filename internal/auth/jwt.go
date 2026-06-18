package auth

import (
	"context"
	"crypto/ed25519"
	"crypto/rand"
	"crypto/x509"
	"encoding/pem"
	"errors"
	"fmt"
	"os"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/RSSembly/rssembly/internal/models"
)

// Context key for storing authenticated user info in request context.
type ctxKey string

const ctxKeyUser ctxKey = "user"

// AuthenticatedUser represents the identity extracted from a verified token
// or API key.
type AuthenticatedUser struct {
	UserID models.UUIDv7
	IsAdmin bool
	IsAPIKey bool // true when authenticated via API key (not JWT)
}

// ErrInvalidToken is returned when a JWT or API key cannot be validated.
var ErrInvalidToken = errors.New("invalid token")

// JWTManager handles Ed25519-signed JWT operations.
type JWTManager struct {
	privateKey ed25519.PrivateKey
	publicKey  ed25519.PublicKey
}

// NewJWTManager loads Ed25519 keys from PEM files and returns a manager.
func NewJWTManager(privateKeyPath, publicKeyPath string) (*JWTManager, error) {
	pkData, err := os.ReadFile(privateKeyPath)
	if err != nil {
		return nil, fmt.Errorf("read private key: %w", err)
	}
	pubData, err := os.ReadFile(publicKeyPath)
	if err != nil {
		return nil, fmt.Errorf("read public key: %w", err)
	}

	privBlock, _ := pem.Decode(pkData)
	if privBlock == nil || privBlock.Type != "PRIVATE KEY" {
		return nil, errors.New("invalid private key PEM")
	}
	privKey, err := x509.ParsePKCS8PrivateKey(privBlock.Bytes)
	if err != nil {
		return nil, fmt.Errorf("parse private key: %w", err)
	}
	ed25519Priv, ok := privKey.(ed25519.PrivateKey)
	if !ok {
		return nil, errors.New("private key is not Ed25519")
	}

	pubBlock, _ := pem.Decode(pubData)
	if pubBlock == nil || pubBlock.Type != "PUBLIC KEY" {
		return nil, errors.New("invalid public key PEM")
	}
	pubKey, err := x509.ParsePKIXPublicKey(pubBlock.Bytes)
	if err != nil {
		return nil, fmt.Errorf("parse public key: %w", err)
	}
	ed25519Pub, ok := pubKey.(ed25519.PublicKey)
	if !ok {
		return nil, errors.New("public key is not Ed25519")
	}

	return &JWTManager{privateKey: ed25519Priv, publicKey: ed25519Pub}, nil
}

// Claims are the custom JWT claims for Rssembly.
type Claims struct {
	jwt.RegisteredClaims
	IsAdmin bool `json:"is_admin"`
}

// GenerateToken creates a new signed JWT for the given user.
func (m *JWTManager) GenerateToken(userID models.UUIDv7, isAdmin bool, expiry time.Duration) (string, error) {
	now := time.Now()
	claims := Claims{
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   userID.String(),
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(expiry)),
			Issuer:    "rssembly",
		},
		IsAdmin: isAdmin,
	}

	token := jwt.NewWithClaims(jwt.SigningMethodEdDSA, claims)
	return token.SignedString(m.privateKey)
}

// VerifyToken parses and validates a JWT string, returning the authenticated user.
func (m *JWTManager) VerifyToken(tokenString string) (*AuthenticatedUser, error) {
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(t *jwt.Token) (any, error) {
		if _, ok := t.Method.(*jwt.SigningMethodEd25519); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", t.Header["alg"])
		}
		return m.publicKey, nil
	})
	if err != nil {
		return nil, fmt.Errorf("%w: %w", ErrInvalidToken, err)
	}

	claims, ok := token.Claims.(*Claims)
	if !ok || !token.Valid {
		return nil, ErrInvalidToken
	}

	uid, err := models.ParseUUIDv7(claims.Subject)
	if err != nil {
		return nil, fmt.Errorf("%w: invalid subject: %w", ErrInvalidToken, err)
	}

	return &AuthenticatedUser{
		UserID:  uid,
		IsAdmin: claims.IsAdmin,
	}, nil
}

// GenerateKeyPair creates a new Ed25519 key pair and writes them as PEM files.
func GenerateKeyPair(privatePath, publicPath string) error {
	_, priv, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		return fmt.Errorf("generate key: %w", err)
	}

	privBytes, err := x509.MarshalPKCS8PrivateKey(priv)
	if err != nil {
		return fmt.Errorf("marshal private key: %w", err)
	}

	pubBytes, err := x509.MarshalPKIXPublicKey(priv.Public())
	if err != nil {
		return fmt.Errorf("marshal public key: %w", err)
	}

	if err := os.MkdirAll("data", 0700); err != nil {
		return fmt.Errorf("create data directory: %w", err)
	}

	privFile, err := os.Create(privatePath)
	if err != nil {
		return fmt.Errorf("create private key file: %w", err)
	}
	defer privFile.Close()
	if err := pem.Encode(privFile, &pem.Block{Type: "PRIVATE KEY", Bytes: privBytes}); err != nil {
		return fmt.Errorf("write private key PEM: %w", err)
	}

	pubFile, err := os.Create(publicPath)
	if err != nil {
		return fmt.Errorf("create public key file: %w", err)
	}
	defer pubFile.Close()
	if err := pem.Encode(pubFile, &pem.Block{Type: "PUBLIC KEY", Bytes: pubBytes}); err != nil {
		return fmt.Errorf("write public key PEM: %w", err)
	}

	return nil
}

// UserFromContext extracts the authenticated user from a request context.
func UserFromContext(ctx context.Context) (*AuthenticatedUser, bool) {
	u, ok := ctx.Value(ctxKeyUser).(*AuthenticatedUser)
	return u, ok
}

// ContextWithUser stores an authenticated user in a context.
func ContextWithUser(ctx context.Context, u *AuthenticatedUser) context.Context {
	return context.WithValue(ctx, ctxKeyUser, u)
}