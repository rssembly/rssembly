package handler

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/go-playground/validator/v10"

	"github.com/rssembly/rssembly/internal/auth"
	"github.com/rssembly/rssembly/internal/models"
	"github.com/rssembly/rssembly/internal/repo"
)

// ArticleHandler handles article listing and read state.
type ArticleHandler struct {
	articleRepo *repo.ArticleRepo
	validate    *validator.Validate
}

// NewArticleHandler creates an ArticleHandler.
func NewArticleHandler(articleRepo *repo.ArticleRepo) *ArticleHandler {
	return &ArticleHandler{
		articleRepo: articleRepo,
		validate:    validator.New(),
	}
}

// setReadStateRequest is the expected body for PUT /api/v1/articles/{articleID}/read-state.
type setReadStateRequest struct {
	State string `json:"state" validate:"required,oneof=unread read saved"`
}

// List handles GET /api/v1/articles.
func (h *ArticleHandler) List(w http.ResponseWriter, r *http.Request) {
	user, ok := auth.UserFromContext(r.Context())
	if !ok {
		RespondError(w, http.StatusUnauthorized, "unauthorized", "authentication required")
		return
	}

	params := repo.ListArticlesParams{
		UserID:    user.UserID,
		Cursor:    r.URL.Query().Get("cursor"),
		ReadState: r.URL.Query().Get("read_state"),
		Search:    r.URL.Query().Get("search"),
	}

	if feedIDStr := r.URL.Query().Get("feed_id"); feedIDStr != "" {
		feedID, err := models.ParseUUIDv7(feedIDStr)
		if err != nil {
			RespondError(w, http.StatusBadRequest, "invalid_feed_id", "invalid feed ID")
			return
		}
		params.FeedID = feedID
	}

	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	if limit <= 0 || limit > 100 {
		limit = 50
	}
	params.Limit = limit

	articles, nextCursor, hasMore, err := h.articleRepo.ListArticles(r.Context(), params)
	if err != nil {
		RespondError(w, http.StatusInternalServerError, "internal", "failed to list articles")
		return
	}

	RespondPaginated(w, articles, nextCursor, hasMore)
}

// Get handles GET /api/v1/articles/{articleID}.
func (h *ArticleHandler) Get(w http.ResponseWriter, r *http.Request) {
	id, err := models.ParseUUIDv7(chi.URLParam(r, "articleID"))
	if err != nil {
		RespondError(w, http.StatusBadRequest, "invalid_id", "invalid article ID")
		return
	}

	article, err := h.articleRepo.GetArticleByID(r.Context(), id)
	if err != nil {
		if errors.Is(err, repo.ErrNotFound) {
			RespondError(w, http.StatusNotFound, "not_found", "article not found")
			return
		}
		RespondError(w, http.StatusInternalServerError, "internal", "failed to get article")
		return
	}

	Respond(w, http.StatusOK, article)
}

// SetReadState handles PUT /api/v1/articles/{articleID}/read-state.
func (h *ArticleHandler) SetReadState(w http.ResponseWriter, r *http.Request) {
	user, ok := auth.UserFromContext(r.Context())
	if !ok {
		RespondError(w, http.StatusUnauthorized, "unauthorized", "authentication required")
		return
	}

	articleID, err := models.ParseUUIDv7(chi.URLParam(r, "articleID"))
	if err != nil {
		RespondError(w, http.StatusBadRequest, "invalid_id", "invalid article ID")
		return
	}

	var req setReadStateRequest
	if err := decodeJSON(r, &req); err != nil {
		RespondError(w, http.StatusBadRequest, "invalid_body", "invalid request body")
		return
	}

	if err := h.validate.Struct(req); err != nil {
		RespondError(w, http.StatusBadRequest, "validation_error", err.Error())
		return
	}

	if err := h.articleRepo.SetReadState(r.Context(), user.UserID, articleID, models.ReadState(req.State)); err != nil {
		RespondError(w, http.StatusInternalServerError, "internal", "failed to set read state")
		return
	}

	Respond(w, http.StatusOK, map[string]string{"status": "ok"})
}
