package handler

import "net/http"

// ArticleHandler handles article listing and read state.
type ArticleHandler struct{}

func NewArticleHandler() *ArticleHandler {
	return &ArticleHandler{}
}

func (h *ArticleHandler) List(w http.ResponseWriter, r *http.Request)           { RespondError(w, http.StatusNotImplemented, "not_implemented", "") }
func (h *ArticleHandler) Get(w http.ResponseWriter, r *http.Request)            { RespondError(w, http.StatusNotImplemented, "not_implemented", "") }
func (h *ArticleHandler) SetReadState(w http.ResponseWriter, r *http.Request)   { RespondError(w, http.StatusNotImplemented, "not_implemented", "") }
