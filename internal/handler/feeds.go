package handler

import "net/http"

// FeedHandler handles feed subscription CRUD.
type FeedHandler struct{}

// NewFeedHandler creates a FeedHandler.
func NewFeedHandler() *FeedHandler {
	return &FeedHandler{}
}

func (h *FeedHandler) List(w http.ResponseWriter, r *http.Request)   { RespondError(w, http.StatusNotImplemented, "not_implemented", "") }
func (h *FeedHandler) Create(w http.ResponseWriter, r *http.Request) { RespondError(w, http.StatusNotImplemented, "not_implemented", "") }
func (h *FeedHandler) Get(w http.ResponseWriter, r *http.Request)    { RespondError(w, http.StatusNotImplemented, "not_implemented", "") }
func (h *FeedHandler) Update(w http.ResponseWriter, r *http.Request) { RespondError(w, http.StatusNotImplemented, "not_implemented", "") }
func (h *FeedHandler) Delete(w http.ResponseWriter, r *http.Request) { RespondError(w, http.StatusNotImplemented, "not_implemented", "") }
