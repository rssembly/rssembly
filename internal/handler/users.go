package handler

import "net/http"

// UserHandler handles user profile operations.
type UserHandler struct{}

func NewUserHandler() *UserHandler {
	return &UserHandler{}
}

func (h *UserHandler) GetProfile(w http.ResponseWriter, r *http.Request)    { RespondError(w, http.StatusNotImplemented, "not_implemented", "") }
func (h *UserHandler) UpdateProfile(w http.ResponseWriter, r *http.Request) { RespondError(w, http.StatusNotImplemented, "not_implemented", "") }
