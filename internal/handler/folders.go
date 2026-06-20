package handler

import (
	"errors"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-playground/validator/v10"

	"github.com/rssembly/rssembly/internal/auth"
	"github.com/rssembly/rssembly/internal/models"
	"github.com/rssembly/rssembly/internal/repo"
)

// FolderHandler handles folder CRUD.
type FolderHandler struct {
	repo     *repo.FolderRepo
	validate *validator.Validate
}

// NewFolderHandler creates a FolderHandler.
func NewFolderHandler(folderRepo *repo.FolderRepo) *FolderHandler {
	return &FolderHandler{
		repo:     folderRepo,
		validate: validator.New(),
	}
}

// createFolderRequest is the expected body for POST /api/v1/folders.
type createFolderRequest struct {
	Name     string `json:"name" validate:"required,min=1,max=64"`
	ParentID string `json:"parent_id,omitempty"`
}

// updateFolderRequest is the expected body for PUT /api/v1/folders/{folderID}.
type updateFolderRequest struct {
	Name     string `json:"name,omitempty"`
	ParentID string `json:"parent_id,omitempty"`
}

// List handles GET /api/v1/folders.
func (h *FolderHandler) List(w http.ResponseWriter, r *http.Request) {
	user, ok := auth.UserFromContext(r.Context())
	if !ok {
		RespondError(w, http.StatusUnauthorized, "unauthorized", "authentication required")
		return
	}

	folders, err := h.repo.ListFoldersByUser(r.Context(), user.UserID)
	if err != nil {
		RespondError(w, http.StatusInternalServerError, "internal", "failed to list folders")
		return
	}

	if folders == nil {
		folders = []*models.Folder{}
	}

	Respond(w, http.StatusOK, folders)
}

// Create handles POST /api/v1/folders.
func (h *FolderHandler) Create(w http.ResponseWriter, r *http.Request) {
	user, ok := auth.UserFromContext(r.Context())
	if !ok {
		RespondError(w, http.StatusUnauthorized, "unauthorized", "authentication required")
		return
	}

	var req createFolderRequest
	if err := decodeJSON(r, &req); err != nil {
		RespondError(w, http.StatusBadRequest, "invalid_body", "invalid request body")
		return
	}

	if err := h.validate.Struct(req); err != nil {
		RespondError(w, http.StatusBadRequest, "validation_error", err.Error())
		return
	}

	folder := &models.Folder{
		UserID: user.UserID,
		Name:   req.Name,
	}

	if req.ParentID != "" {
		parentID, err := models.ParseUUIDv7(req.ParentID)
		if err != nil {
			RespondError(w, http.StatusBadRequest, "invalid_parent_id", "invalid parent folder ID")
			return
		}
		folder.ParentID = &parentID
	}

	created, err := h.repo.CreateFolder(r.Context(), folder)
	if err != nil {
		if errors.Is(err, repo.ErrConflict) {
			RespondError(w, http.StatusConflict, "folder_exists", "a folder with this name already exists")
			return
		}
		RespondError(w, http.StatusInternalServerError, "internal", "failed to create folder")
		return
	}

	Respond(w, http.StatusCreated, created)
}

// Update handles PUT /api/v1/folders/{folderID}.
func (h *FolderHandler) Update(w http.ResponseWriter, r *http.Request) {
	id, err := models.ParseUUIDv7(chi.URLParam(r, "folderID"))
	if err != nil {
		RespondError(w, http.StatusBadRequest, "invalid_id", "invalid folder ID")
		return
	}

	var req updateFolderRequest
	if err := decodeJSON(r, &req); err != nil {
		RespondError(w, http.StatusBadRequest, "invalid_body", "invalid request body")
		return
	}

	// Fetch existing to avoid overwriting fields not in request.
	folder, err := h.repo.GetFolderByID(r.Context(), id)
	if err != nil {
		if errors.Is(err, repo.ErrNotFound) {
			RespondError(w, http.StatusNotFound, "not_found", "folder not found")
			return
		}
		RespondError(w, http.StatusInternalServerError, "internal", "failed to get folder")
		return
	}

	if req.Name != "" {
		folder.Name = req.Name
	}
	if req.ParentID != "" {
		parentID, err := models.ParseUUIDv7(req.ParentID)
		if err != nil {
			RespondError(w, http.StatusBadRequest, "invalid_parent_id", "invalid parent folder ID")
			return
		}
		folder.ParentID = &parentID
	}

	updated, err := h.repo.UpdateFolder(r.Context(), folder)
	if err != nil {
		if errors.Is(err, repo.ErrNotFound) {
			RespondError(w, http.StatusNotFound, "not_found", "folder not found")
			return
		}
		if errors.Is(err, repo.ErrConflict) {
			RespondError(w, http.StatusConflict, "folder_exists", "a folder with this name already exists")
			return
		}
		RespondError(w, http.StatusInternalServerError, "internal", "failed to update folder")
		return
	}

	Respond(w, http.StatusOK, updated)
}

// Delete handles DELETE /api/v1/folders/{folderID}.
func (h *FolderHandler) Delete(w http.ResponseWriter, r *http.Request) {
	id, err := models.ParseUUIDv7(chi.URLParam(r, "folderID"))
	if err != nil {
		RespondError(w, http.StatusBadRequest, "invalid_id", "invalid folder ID")
		return
	}

	if err := h.repo.DeleteFolder(r.Context(), id); err != nil {
		if errors.Is(err, repo.ErrNotFound) {
			RespondError(w, http.StatusNotFound, "not_found", "folder not found")
			return
		}
		RespondError(w, http.StatusInternalServerError, "internal", "failed to delete folder")
		return
	}

	RespondNoContent(w)
}