package handler

import (
	"encoding/json"
	"net/http"
)

// ErrorPayload is the standard error body.
type ErrorPayload struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

// ErrorResponse is the envelope for error responses.
type ErrorResponse struct {
	Error ErrorPayload `json:"error"`
}

// Respond writes a JSON success response with the given status code and payload.
func Respond(w http.ResponseWriter, status int, data any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if data != nil {
		json.NewEncoder(w).Encode(data)
	}
}

// RespondError writes a JSON error response.
func RespondError(w http.ResponseWriter, status int, code, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(ErrorResponse{
		Error: ErrorPayload{Code: code, Message: message},
	})
}

// RespondPaginated writes a cursor-paginated list response.
type PaginatedResponse struct {
	Data       any    `json:"data"`
	NextCursor string `json:"next_cursor,omitempty"`
	HasMore    bool   `json:"has_more"`
}

func RespondPaginated(w http.ResponseWriter, data any, nextCursor string, hasMore bool) {
	Respond(w, http.StatusOK, PaginatedResponse{
		Data:       data,
		NextCursor: nextCursor,
		HasMore:    hasMore,
	})
}

// RespondNoContent writes a 204 with no body.
func RespondNoContent(w http.ResponseWriter) {
	w.WriteHeader(http.StatusNoContent)
}
