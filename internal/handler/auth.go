package handler

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/go-playground/validator/v10"

	"github.com/rssembly/rssembly/internal/auth"
	"github.com/rssembly/rssembly/internal/database"
	"github.com/rssembly/rssembly/internal/models"
	"github.com/rssembly/rssembly/internal/repo"
)

// AuthHandler handles user registration and authentication.
type AuthHandler struct {
	repo       *repo.UserRepo
	jwtManager *auth.JWTManager
	validate   *validator.Validate
	db         *database.Pool
}

// NewAuthHandler creates an AuthHandler with its dependencies.
func NewAuthHandler(db *database.Pool, jwtManager *auth.JWTManager) *AuthHandler {
	return &AuthHandler{
		repo:       repo.NewUserRepo(db),
		jwtManager: jwtManager,
		validate:   validator.New(),
		db:         db,
	}
}

// registerRequest is the expected body for POST /auth/register.
type registerRequest struct {
	Username string `json:"username" validate:"required,min=2,max=32"`
	Email    string `json:"email"    validate:"required,email"`
	Password string `json:"password" validate:"required,min=8,max=128"`
}

// loginRequest is the expected body for POST /auth/login.
type loginRequest struct {
	Email    string `json:"email"    validate:"required,email"`
	Password string `json:"password" validate:"required"`
}

// registerResponse is returned on successful registration.
type registerResponse struct {
	Token string       `json:"token"`
	User  *models.User `json:"user"`
}

// loginResponse is returned on successful login.
type loginResponse struct {
	Token string       `json:"token"`
	User  *models.User `json:"user"`
}

// Register handles POST /api/v1/auth/register.
func (h *AuthHandler) Register(w http.ResponseWriter, r *http.Request) {
	var req registerRequest
	if err := decodeJSON(r, &req); err != nil {
		RespondError(w, http.StatusBadRequest, "invalid_body", "invalid request body")
		return
	}

	if err := h.validate.Struct(req); err != nil {
		RespondError(w, http.StatusBadRequest, "validation_error", err.Error())
		return
	}

	// Check if email already exists.
	if _, err := h.repo.GetUserByEmail(r.Context(), req.Email); err == nil {
		RespondError(w, http.StatusConflict, "email_exists", "a user with this email already exists")
		return
	}

	// Hash the password.
	hash, err := auth.HashPassword(req.Password)
	if err != nil {
		RespondError(w, http.StatusInternalServerError, "internal", "failed to process password")
		return
	}

	// Default scopes for new users.
	defaultScopes := []string{
		"feeds:read", "feeds:write",
		"articles:read", "articles:write",
		"folders:read", "folders:write",
	}

	user := &models.User{
		Username:     req.Username,
		Email:        req.Email,
		PasswordHash: hash,
		Scopes:       defaultScopes,
	}

	created, err := h.repo.CreateUser(r.Context(), user)
	if err != nil {
		RespondError(w, http.StatusInternalServerError, "internal", "failed to create user")
		return
	}

	// Generate JWT.
	token, err := h.jwtManager.GenerateToken(created.ID, created.Scopes, auth.DefaultTokenExpiry)
	if err != nil {
		RespondError(w, http.StatusInternalServerError, "internal", "failed to generate token")
		return
	}

	Respond(w, http.StatusCreated, registerResponse{
		Token: token,
		User:  created,
	})
}

// Login handles POST /api/v1/auth/login.
func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	var req loginRequest
	if err := decodeJSON(r, &req); err != nil {
		RespondError(w, http.StatusBadRequest, "invalid_body", "invalid request body")
		return
	}

	if err := h.validate.Struct(req); err != nil {
		RespondError(w, http.StatusBadRequest, "validation_error", err.Error())
		return
	}

	user, err := h.repo.GetUserByEmail(r.Context(), req.Email)
	if err != nil {
		if errors.Is(err, repo.ErrNotFound) {
			RespondError(w, http.StatusUnauthorized, "invalid_credentials", "invalid email or password")
			return
		}
		RespondError(w, http.StatusInternalServerError, "internal", "failed to look up user")
		return
	}

	match, err := auth.VerifyPassword(req.Password, user.PasswordHash)
	if err != nil || !match {
		RespondError(w, http.StatusUnauthorized, "invalid_credentials", "invalid email or password")
		return
	}

	token, err := h.jwtManager.GenerateToken(user.ID, user.Scopes, auth.DefaultTokenExpiry)
	if err != nil {
		RespondError(w, http.StatusInternalServerError, "internal", "failed to generate token")
		return
	}

	Respond(w, http.StatusOK, loginResponse{
		Token: token,
		User:  user,
	})
}

// decodeJSON decodes a JSON request body into the target value.
func decodeJSON(r *http.Request, v any) error {
	defer r.Body.Close()
	return json.NewDecoder(r.Body).Decode(v)
}
