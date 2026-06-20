package handler

import (
	"errors"
	"net/http"
	"strconv"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-playground/validator/v10"

	"github.com/rssembly/rssembly/internal/auth"
	"github.com/rssembly/rssembly/internal/models"
	"github.com/rssembly/rssembly/internal/repo"
)

// FeedHandler handles feed subscription CRUD.
type FeedHandler struct {
	repo     *repo.FeedRepo
	validate *validator.Validate
}

// NewFeedHandler creates a FeedHandler.
func NewFeedHandler(feedRepo *repo.FeedRepo) *FeedHandler {
	return &FeedHandler{
		repo:     feedRepo,
		validate: validator.New(),
	}
}

// createFeedRequest is the expected body for POST /api/v1/feeds.
type createFeedRequest struct {
	FeedURL  string `json:"feed_url" validate:"required,url"`
	FolderID string `json:"folder_id,omitempty"`
}

// updateFeedRequest is the expected body for PUT /api/v1/feeds/{feedID}.
type updateFeedRequest struct {
	Title        string `json:"title,omitempty"`
	PollInterval string `json:"poll_interval,omitempty"`
	FolderID     string `json:"folder_id,omitempty"`
}

// List handles GET /api/v1/feeds.
func (h *FeedHandler) List(w http.ResponseWriter, r *http.Request) {
	user, ok := auth.UserFromContext(r.Context())
	if !ok {
		RespondError(w, http.StatusUnauthorized, "unauthorized", "authentication required")
		return
	}

	cursor := r.URL.Query().Get("cursor")
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	if limit <= 0 || limit > 100 {
		limit = 50
	}

	feeds, nextCursor, hasMore, err := h.repo.ListFeedsByUser(r.Context(), user.UserID, cursor, limit)
	if err != nil {
		RespondError(w, http.StatusInternalServerError, "internal", "failed to list feeds")
		return
	}

	RespondPaginated(w, feeds, nextCursor, hasMore)
}

// Create handles POST /api/v1/feeds.
func (h *FeedHandler) Create(w http.ResponseWriter, r *http.Request) {
	user, ok := auth.UserFromContext(r.Context())
	if !ok {
		RespondError(w, http.StatusUnauthorized, "unauthorized", "authentication required")
		return
	}

	var req createFeedRequest
	if err := decodeJSON(r, &req); err != nil {
		RespondError(w, http.StatusBadRequest, "invalid_body", "invalid request body")
		return
	}

	if err := h.validate.Struct(req); err != nil {
		RespondError(w, http.StatusBadRequest, "validation_error", err.Error())
		return
	}

	feed := &models.Feed{
		CreatedBy: user.UserID,
		FeedURL:   req.FeedURL,
	}

	if req.FolderID != "" {
		folderID, err := models.ParseUUIDv7(req.FolderID)
		if err != nil {
			RespondError(w, http.StatusBadRequest, "invalid_folder_id", "invalid folder ID")
			return
		}
		feed.FolderID = &folderID
	}

	created, err := h.repo.CreateFeed(r.Context(), feed)
	if err != nil {
		if errors.Is(err, repo.ErrConflict) {
			RespondError(w, http.StatusConflict, "feed_exists", "already subscribed to this feed")
			return
		}
		RespondError(w, http.StatusInternalServerError, "internal", "failed to create feed")
		return
	}

	Respond(w, http.StatusCreated, created)
}

// Get handles GET /api/v1/feeds/{feedID}.
func (h *FeedHandler) Get(w http.ResponseWriter, r *http.Request) {
	id, err := models.ParseUUIDv7(chi.URLParam(r, "feedID"))
	if err != nil {
		RespondError(w, http.StatusBadRequest, "invalid_id", "invalid feed ID")
		return
	}

	feed, err := h.repo.GetFeedByID(r.Context(), id)
	if err != nil {
		if errors.Is(err, repo.ErrNotFound) {
			RespondError(w, http.StatusNotFound, "not_found", "feed not found")
			return
		}
		RespondError(w, http.StatusInternalServerError, "internal", "failed to get feed")
		return
	}

	Respond(w, http.StatusOK, feed)
}

// Update handles PUT /api/v1/feeds/{feedID}.
func (h *FeedHandler) Update(w http.ResponseWriter, r *http.Request) {
	id, err := models.ParseUUIDv7(chi.URLParam(r, "feedID"))
	if err != nil {
		RespondError(w, http.StatusBadRequest, "invalid_id", "invalid feed ID")
		return
	}

	var req updateFeedRequest
	if err := decodeJSON(r, &req); err != nil {
		RespondError(w, http.StatusBadRequest, "invalid_body", "invalid request body")
		return
	}

	// Fetch existing feed to avoid overwriting fields not in the request body.
	feed, err := h.repo.GetFeedByID(r.Context(), id)
	if err != nil {
		if errors.Is(err, repo.ErrNotFound) {
			RespondError(w, http.StatusNotFound, "not_found", "feed not found")
			return
		}
		RespondError(w, http.StatusInternalServerError, "internal", "failed to get feed")
		return
	}

	if req.Title != "" {
		feed.Title = req.Title
	}
	if req.PollInterval != "" {
		d, err := time.ParseDuration(req.PollInterval)
		if err != nil {
			RespondError(w, http.StatusBadRequest, "invalid_poll_interval", "invalid poll interval format")
			return
		}
		feed.PollInterval = d
	}
	if req.FolderID != "" {
		folderID, err := models.ParseUUIDv7(req.FolderID)
		if err != nil {
			RespondError(w, http.StatusBadRequest, "invalid_folder_id", "invalid folder ID")
			return
		}
		feed.FolderID = &folderID
	}

	updated, err := h.repo.UpdateFeed(r.Context(), feed)
	if err != nil {
		if errors.Is(err, repo.ErrNotFound) {
			RespondError(w, http.StatusNotFound, "not_found", "feed not found")
			return
		}
		RespondError(w, http.StatusInternalServerError, "internal", "failed to update feed")
		return
	}

	Respond(w, http.StatusOK, updated)
}

// Delete handles DELETE /api/v1/feeds/{feedID}.
func (h *FeedHandler) Delete(w http.ResponseWriter, r *http.Request) {
	id, err := models.ParseUUIDv7(chi.URLParam(r, "feedID"))
	if err != nil {
		RespondError(w, http.StatusBadRequest, "invalid_id", "invalid feed ID")
		return
	}

	if err := h.repo.DeleteFeed(r.Context(), id); err != nil {
		if errors.Is(err, repo.ErrNotFound) {
			RespondError(w, http.StatusNotFound, "not_found", "feed not found")
			return
		}
		RespondError(w, http.StatusInternalServerError, "internal", "failed to delete feed")
		return
	}

	RespondNoContent(w)
}
