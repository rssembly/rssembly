package handler

import "net/http"

// FolderHandler handles folder CRUD.
type FolderHandler struct{}

func NewFolderHandler() *FolderHandler {
	return &FolderHandler{}
}

func (h *FolderHandler) List(w http.ResponseWriter, r *http.Request)   { RespondError(w, http.StatusNotImplemented, "not_implemented", "") }
func (h *FolderHandler) Create(w http.ResponseWriter, r *http.Request) { RespondError(w, http.StatusNotImplemented, "not_implemented", "") }
func (h *FolderHandler) Update(w http.ResponseWriter, r *http.Request) { RespondError(w, http.StatusNotImplemented, "not_implemented", "") }
func (h *FolderHandler) Delete(w http.ResponseWriter, r *http.Request) { RespondError(w, http.StatusNotImplemented, "not_implemented", "") }
