package handler

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/go-chi/chi/v5"
)

func TestUserHandler_GetProfileUnauthenticated(t *testing.T) {
	h := NewUserHandler(nil)

	r := httptest.NewRequest("GET", "/api/v1/users/me", nil)

	router := chi.NewRouter()
	router.Get("/api/v1/users/me", h.GetProfile)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, r)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("expected 401, got %d", w.Code)
	}
}

func TestUserHandler_UpdateProfileInvalidBody(t *testing.T) {
	h := NewUserHandler(nil)

	r := httptest.NewRequest("PUT", "/api/v1/users/me", nil)
	r = r.WithContext(testCtx())

	router := chi.NewRouter()
	router.Put("/api/v1/users/me", h.UpdateProfile)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, r)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", w.Code)
	}

	var resp ErrorResponse
	json.NewDecoder(w.Body).Decode(&resp)
	if resp.Error.Code != "invalid_body" {
		t.Errorf("expected invalid_body, got %q", resp.Error.Code)
	}
}

func TestUserHandler_UpdateProfileValidationError(t *testing.T) {
	h := NewUserHandler(nil)

	body := `{"password":"short"}`
	r := httptest.NewRequest("PUT", "/api/v1/users/me", strings.NewReader(body))
	r.Header.Set("Content-Type", "application/json")
	r = r.WithContext(testCtx())

	router := chi.NewRouter()
	router.Put("/api/v1/users/me", h.UpdateProfile)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, r)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", w.Code)
	}

	var resp ErrorResponse
	json.NewDecoder(w.Body).Decode(&resp)
	if resp.Error.Code != "validation_error" {
		t.Errorf("expected validation_error, got %q", resp.Error.Code)
	}
}
