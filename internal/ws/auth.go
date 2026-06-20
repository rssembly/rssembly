package ws

import (
	"github.com/rssembly/rssembly/internal/auth"
)

// NewJWTAdapter wraps an auth.JWTManager into the ws.Authenticator interface.
func NewJWTAdapter(jwtManager *auth.JWTManager) Authenticator {
	return &jwtAdapter{jwt: jwtManager}
}

type jwtAdapter struct {
	jwt *auth.JWTManager
}

func (a *jwtAdapter) VerifyToken(tokenString string) (*AuthUser, error) {
	user, err := a.jwt.VerifyToken(tokenString)
	if err != nil {
		return nil, err
	}
	return &AuthUser{
		UserID: user.UserID.String(),
	}, nil
}
