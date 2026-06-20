package handler

import (
	"errors"
	"net/http"

	"github.com/go-playground/validator/v10"

	"github.com/rssembly/rssembly/internal/auth"
	"github.com/rssembly/rssembly/internal/repo"
)

// UserHandler handles user profile operations.
type UserHandler struct {
	repo     *repo.UserRepo
	validate *validator.Validate
}

// NewUserHandler creates a UserHandler.
func NewUserHandler(userRepo *repo.UserRepo) *UserHandler {
	return &UserHandler{
		repo:     userRepo,
		validate: validator.New(),
	}
}

// updateProfileRequest is the expected body for PUT /api/v1/users/me.
type updateProfileRequest struct {
	DisplayName string `json:"display_name,omitempty"`
	Password    string `json:"password,omitempty" validate:"omitempty,min=8,max=128"`
}

// GetProfile handles GET /api/v1/users/me.
func (h *UserHandler) GetProfile(w http.ResponseWriter, r *http.Request) {
	user, ok := auth.UserFromContext(r.Context())
	if !ok {
		RespondError(w, http.StatusUnauthorized, "unauthorized", "authentication required")
		return
	}

	u, err := h.repo.GetUserByID(r.Context(), user.UserID)
	if err != nil {
		RespondError(w, http.StatusInternalServerError, "internal", "failed to get profile")
		return
	}

	Respond(w, http.StatusOK, u)
}

// UpdateProfile handles PUT /api/v1/users/me.
func (h *UserHandler) UpdateProfile(w http.ResponseWriter, r *http.Request) {
	currentUser, ok := auth.UserFromContext(r.Context())
	if !ok {
		RespondError(w, http.StatusUnauthorized, "unauthorized", "authentication required")
		return
	}

	var req updateProfileRequest
	if err := decodeJSON(r, &req); err != nil {
		RespondError(w, http.StatusBadRequest, "invalid_body", "invalid request body")
		return
	}

	if err := h.validate.Struct(req); err != nil {
		RespondError(w, http.StatusBadRequest, "validation_error", err.Error())
		return
	}

	u, err := h.repo.GetUserByID(r.Context(), currentUser.UserID)
	if err != nil {
		RespondError(w, http.StatusInternalServerError, "internal", "failed to get profile")
		return
	}

	if req.DisplayName != "" {
		u.Username = req.DisplayName
	}
	if req.Password != "" {
		hash, err := auth.HashPassword(req.Password)
		if err != nil {
			RespondError(w, http.StatusInternalServerError, "internal", "failed to process password")
			return
		}
		u.PasswordHash = hash
	}

	updated, err := h.repo.UpdateUser(r.Context(), u)
	if err != nil {
		if errors.Is(err, repo.ErrConflict) {
			RespondError(w, http.StatusConflict, "username_taken", "this username is already taken")
			return
		}
		RespondError(w, http.StatusInternalServerError, "internal", "failed to update profile")
		return
	}

	Respond(w, http.StatusOK, updated)
}