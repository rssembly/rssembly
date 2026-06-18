package handler

import (
	"context"
	"net/http"
	"time"
)

// HealthChecker defines the interface for checking dependencies' health.
type HealthChecker interface {
	Ping(ctx context.Context) error
}

// HealthHandler handles liveness and readiness checks.
type HealthHandler struct {
	db HealthChecker
}

// NewHealthHandler creates a HealthHandler.
func NewHealthHandler(db HealthChecker) *HealthHandler {
	return &HealthHandler{db: db}
}

// Liveness responds to GET /health — always 200 if the process is alive.
func (h *HealthHandler) Liveness(w http.ResponseWriter, r *http.Request) {
	Respond(w, http.StatusOK, map[string]string{"status": "ok"})
}

// Readiness responds to GET /ready — 200 when dependencies are reachable, 503 otherwise.
func (h *HealthHandler) Readiness(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 3*time.Second)
	defer cancel()

	status := "ok"
	code := http.StatusOK

	if err := h.db.Ping(ctx); err != nil {
		status = "degraded"
		code = http.StatusServiceUnavailable
	}

	Respond(w, code, map[string]string{"status": status})
}
