package handler

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/go-chi/chi/v5"

	"github.com/rssembly/rssembly/internal/auth"
	"github.com/rssembly/rssembly/internal/models"
)

// testCtx returns a context with an authenticated user suitable for testing handlers.
func testCtx() context.Context {
	return auth.ContextWithUser(context.Background(), &auth.AuthenticatedUser{
		UserID: models.NewUUIDv7(),
		Scopes: []string{"*"},
	})
}

func TestFeedHandler_ListUnauthenticated(t *testing.T) {
	h := NewFeedHandler(nil)

	r := httptest.NewRequest("GET", "/api/v1/feeds", nil)

	router := chi.NewRouter()
	router.Get("/api/v1/feeds", h.List)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, r)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("expected 401, got %d", w.Code)
	}

	var resp ErrorResponse
	json.NewDecoder(w.Body).Decode(&resp)
	if resp.Error.Code != "unauthorized" {
		t.Errorf("expected unauthorized, got %q", resp.Error.Code)
	}
}

func TestFeedHandler_CreateInvalidBody(t *testing.T) {
	h := NewFeedHandler(nil)

	r := httptest.NewRequest("POST", "/api/v1/feeds", nil)
	r = r.WithContext(testCtx())

	router := chi.NewRouter()
	router.Post("/api/v1/feeds", h.Create)
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

func TestFeedHandler_CreateValidationError(t *testing.T) {
	h := NewFeedHandler(nil)

	body := `{"feed_url":"not-a-url"}`
	r := httptest.NewRequest("POST", "/api/v1/feeds", strings.NewReader(body))
	r.Header.Set("Content-Type", "application/json")
	r = r.WithContext(testCtx())

	router := chi.NewRouter()
	router.Post("/api/v1/feeds", h.Create)
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

func TestFeedHandler_GetInvalidID(t *testing.T) {
	h := NewFeedHandler(nil)

	r := httptest.NewRequest("GET", "/api/v1/feeds/not-a-uuid", nil)

	router := chi.NewRouter()
	router.Get("/api/v1/feeds/{feedID}", h.Get)
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

func TestFeedHandler_DeleteInvalidID(t *testing.T) {
	h := NewFeedHandler(nil)

	r := httptest.NewRequest("DELETE", "/api/v1/feeds/not-a-uuid", nil)

	router := chi.NewRouter()
	router.Delete("/api/v1/feeds/{feedID}", h.Delete)
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

func TestFeedHandler_UpdateInvalidBody(t *testing.T) {
	id := "00000000-0000-0000-0000-000000000001"
	h := NewFeedHandler(nil)

	r := httptest.NewRequest("PUT", "/api/v1/feeds/"+id, nil)
	r = r.WithContext(testCtx())

	router := chi.NewRouter()
	router.Put("/api/v1/feeds/{feedID}", h.Update)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, r)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", w.Code)
	}
}