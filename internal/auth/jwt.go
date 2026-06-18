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
	"github.com/rssembly/rssembly/internal/models"
)

// DefaultTokenExpiry is the default JWT lifetime (30 days).
const DefaultTokenExpiry = 30 * 24 * time.Hour

// Context key for storing authenticated user info in request context.
type ctxKey string

const ctxKeyUser ctxKey = "user"

// AuthenticatedUser represents the identity extracted from a verified token
// or API key.
type AuthenticatedUser struct {
	UserID   models.UUIDv7
	Scopes   []string // permission scopes; ["*"] = superadmin
	IsAPIKey bool     // true when authenticated via API key (not JWT)
}

// ErrInvalidToken is returned when a JWT or API key cannot be validated.
var ErrInvalidToken = errors.New("invalid token")

// JWTManager handles Ed25519-signed JWT operations.
type JWTManager struct {
	privateKey ed25519.PrivateKey
	publicKey  ed25519.PublicKey
}

// NewJWTManagerFromPEM creates a JWTManager from Ed25519 keys in PEM-encoded bytes.
func NewJWTManagerFromPEM(pemPriv, pemPub []byte) (*JWTManager, error) {
	privBlock, _ := pem.Decode(pemPriv)
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

	pubBlock, _ := pem.Decode(pemPub)
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
	Scopes []string `json:"scopes"`
}

// GenerateToken creates a new signed JWT for the given user.
func (m *JWTManager) GenerateToken(userID models.UUIDv7, scopes []string, expiry time.Duration) (string, error) {
	now := time.Now()
	claims := Claims{
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   userID.String(),
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(expiry)),
			Issuer:    "rssembly",
		},
		Scopes: scopes,
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

	scopes := claims.Scopes
	if scopes == nil {
		scopes = []string{}
	}

	return &AuthenticatedUser{
		UserID: uid,
		Scopes: scopes,
	}, nil
}

// GenerateKeyPairPEM generates a new Ed25519 key pair and returns PEM-encoded bytes.
func GenerateKeyPairPEM() (privPEM, pubPEM []byte, _ error) {
	_, priv, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		return nil, nil, fmt.Errorf("generate key: %w", err)
	}

	privBytes, err := x509.MarshalPKCS8PrivateKey(priv)
	if err != nil {
		return nil, nil, fmt.Errorf("marshal private key: %w", err)
	}

	pubBytes, err := x509.MarshalPKIXPublicKey(priv.Public())
	if err != nil {
		return nil, nil, fmt.Errorf("marshal public key: %w", err)
	}

	privPEM = pem.EncodeToMemory(&pem.Block{Type: "PRIVATE KEY", Bytes: privBytes})
	pubPEM = pem.EncodeToMemory(&pem.Block{Type: "PUBLIC KEY", Bytes: pubBytes})

	return privPEM, pubPEM, nil
}

// GenerateKeyPair creates a new Ed25519 key pair and writes them as PEM files.
func GenerateKeyPair(privatePath, publicPath string) error {
	privPEM, pubPEM, err := GenerateKeyPairPEM()
	if err != nil {
		return err
	}

	if err := os.MkdirAll("data", 0700); err != nil {
		return fmt.Errorf("create data directory: %w", err)
	}

	if err := os.WriteFile(privatePath, privPEM, 0600); err != nil {
		return fmt.Errorf("write private key: %w", err)
	}
	if err := os.WriteFile(publicPath, pubPEM, 0644); err != nil {
		return fmt.Errorf("write public key: %w", err)
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