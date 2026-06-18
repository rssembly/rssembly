package handler

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/go-playground/validator/v10"
)

func TestAuthHandler_MissingBody(t *testing.T) {
	h := &AuthHandler{}

	w := httptest.NewRecorder()
	r := httptest.NewRequest("POST", "/api/v1/auth/register", nil)

	router := chi.NewRouter()
	router.Post("/api/v1/auth/register", h.Register)
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

func TestAuthHandler_ValidationErrors(t *testing.T) {
	h := &AuthHandler{validate: validator.New()}

	body := `{"username":"a","email":"not-an-email","password":"short"}`
	w := httptest.NewRecorder()
	r := httptest.NewRequest("POST", "/api/v1/auth/register", strings.NewReader(body))
	r.Header.Set("Content-Type", "application/json")

	router := chi.NewRouter()
	router.Post("/api/v1/auth/register", h.Register)
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

func TestAuthHandler_LoginInvalidBody(t *testing.T) {
	h := &AuthHandler{}

	body := `not json`
	w := httptest.NewRecorder()
	r := httptest.NewRequest("POST", "/api/v1/auth/login", strings.NewReader(body))
	r.Header.Set("Content-Type", "application/json")

	router := chi.NewRouter()
	router.Post("/api/v1/auth/login", h.Login)
	router.ServeHTTP(w, r)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", w.Code)
	}
}
