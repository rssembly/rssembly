package handler

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/go-chi/chi/v5"
)

func TestFolderHandler_ListUnauthenticated(t *testing.T) {
	h := NewFolderHandler(nil)

	r := httptest.NewRequest("GET", "/api/v1/folders", nil)

	router := chi.NewRouter()
	router.Get("/api/v1/folders", h.List)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, r)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("expected 401, got %d", w.Code)
	}
}

func TestFolderHandler_CreateInvalidBody(t *testing.T) {
	h := NewFolderHandler(nil)

	r := httptest.NewRequest("POST", "/api/v1/folders", nil)
	r = r.WithContext(testCtx())

	router := chi.NewRouter()
	router.Post("/api/v1/folders", h.Create)
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

func TestFolderHandler_CreateValidationError(t *testing.T) {
	h := NewFolderHandler(nil)

	body := `{"name":""}`
	r := httptest.NewRequest("POST", "/api/v1/folders", strings.NewReader(body))
	r.Header.Set("Content-Type", "application/json")
	r = r.WithContext(testCtx())

	router := chi.NewRouter()
	router.Post("/api/v1/folders", h.Create)
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

func TestFolderHandler_UpdateInvalidID(t *testing.T) {
	h := NewFolderHandler(nil)

	body := `{"name":"New Name"}`
	r := httptest.NewRequest("PUT", "/api/v1/folders/bad-id", strings.NewReader(body))
	r.Header.Set("Content-Type", "application/json")

	router := chi.NewRouter()
	router.Put("/api/v1/folders/{folderID}", h.Update)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, r)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", w.Code)
	}

	var resp ErrorResponse
	json.NewDecoder(w.Body).Decode(&resp)
	if resp.Error.Code != "invalid_id" {
		t.Errorf("expected invalid_id, got %q", resp.Error.Code)
	}
}

func TestFolderHandler_DeleteInvalidID(t *testing.T) {
	h := NewFolderHandler(nil)

	r := httptest.NewRequest("DELETE", "/api/v1/folders/bad-id", nil)

	router := chi.NewRouter()
	router.Delete("/api/v1/folders/{folderID}", h.Delete)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, r)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", w.Code)
	}

	var resp ErrorResponse
	json.NewDecoder(w.Body).Decode(&resp)
	if resp.Error.Code != "invalid_id" {
		t.Errorf("expected invalid_id, got %q", resp.Error.Code)
	}
}