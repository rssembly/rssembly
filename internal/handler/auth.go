package handler

import "net/http"

// AuthHandler handles user registration and authentication.
type AuthHandler struct{}

// NewAuthHandler creates an AuthHandler.
func NewAuthHandler() *AuthHandler {
	return &AuthHandler{}
}

// Register handles POST /api/v1/auth/register.
func (h *AuthHandler) Register(w http.ResponseWriter, r *http.Request) {
	RespondError(w, http.StatusNotImplemented, "not_implemented", "registration not yet implemented")
}

// Login handles POST /api/v1/auth/login.
func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	RespondError(w, http.StatusNotImplemented, "not_implemented", "login not yet implemented")
}
