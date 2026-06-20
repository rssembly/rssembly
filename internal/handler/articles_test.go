package handler

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/go-chi/chi/v5"
)

func TestArticleHandler_ListUnauthenticated(t *testing.T) {
	h := NewArticleHandler(nil)

	r := httptest.NewRequest("GET", "/api/v1/articles", nil)

	router := chi.NewRouter()
	router.Get("/api/v1/articles", h.List)
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

func TestArticleHandler_GetInvalidID(t *testing.T) {
	h := NewArticleHandler(nil)

	r := httptest.NewRequest("GET", "/api/v1/articles/bad-id", nil)

	router := chi.NewRouter()
	router.Get("/api/v1/articles/{articleID}", h.Get)
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

func TestArticleHandler_SetReadStateInvalidBody(t *testing.T) {
	id := "00000000-0000-0000-0000-000000000001"
	h := NewArticleHandler(nil)

	r := httptest.NewRequest("PUT", "/api/v1/articles/"+id+"/read-state", nil)
	r = r.WithContext(testCtx())

	router := chi.NewRouter()
	router.Put("/api/v1/articles/{articleID}/read-state", h.SetReadState)
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

func TestArticleHandler_SetReadStateInvalidID(t *testing.T) {
	h := NewArticleHandler(nil)

	body := `{"state":"read"}`
	r := httptest.NewRequest("PUT", "/api/v1/articles/bad-id/read-state", strings.NewReader(body))
	r.Header.Set("Content-Type", "application/json")
	r = r.WithContext(testCtx())

	router := chi.NewRouter()
	router.Put("/api/v1/articles/{articleID}/read-state", h.SetReadState)
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

func TestArticleHandler_SetReadStateInvalidStateValue(t *testing.T) {
	id := "00000000-0000-0000-0000-000000000001"
	h := NewArticleHandler(nil)

	body := `{"state":"invalid_state"}`
	r := httptest.NewRequest("PUT", "/api/v1/articles/"+id+"/read-state", strings.NewReader(body))
	r.Header.Set("Content-Type", "application/json")
	r = r.WithContext(testCtx())

	router := chi.NewRouter()
	router.Put("/api/v1/articles/{articleID}/read-state", h.SetReadState)
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
