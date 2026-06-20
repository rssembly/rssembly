package handler

import (
	"net/http"

	"github.com/go-chi/chi/v5"

	"github.com/rssembly/rssembly/internal/middleware"
)

// Handlers is a bundle of all HTTP handlers, injected from main.go.
type Handlers struct {
	Auth     *AuthHandler
	Feeds    *FeedHandler
	Articles *ArticleHandler
	Folders  *FolderHandler
	Users    *UserHandler
	Health   *HealthHandler
}

// RegisterRoutes wires all routes onto the Chi router.
// Routes at root level (health, ready, metrics) are not authenticated.
// All /api/v1/* routes use the auth middleware, with per-resource scope checks.
func RegisterRoutes(r chi.Router, h *Handlers, authMW *middleware.Auth, metricsHandler http.HandlerFunc) {
	// ── Health / system (no auth) ───────────────────────────────────
	r.Get("/health", h.Health.Liveness)
	r.Get("/ready", h.Health.Readiness)
	r.Get("/metrics", metricsHandler)

	// ── API v1 ─────────────────────────────────────────────────────
	r.Route("/api/v1", func(r chi.Router) {
		// Public auth endpoints (no auth middleware).
		r.Post("/auth/register", h.Auth.Register)
		r.Post("/auth/login", h.Auth.Login)

		// Authenticated + scoped endpoints.
		r.Group(func(r chi.Router) {
			r.Use(authMW.Middleware)

			// Feeds
			r.Route("/feeds", func(r chi.Router) {
				r.With(authMW.RequireScope("feeds:read")).Get("/", h.Feeds.List)
				r.With(authMW.RequireScope("feeds:write")).Post("/", h.Feeds.Create)
				r.With(authMW.RequireScope("feeds:read")).Get("/{feedID}", h.Feeds.Get)
				r.With(authMW.RequireScope("feeds:write")).Put("/{feedID}", h.Feeds.Update)
				r.With(authMW.RequireScope("feeds:delete")).Delete("/{feedID}", h.Feeds.Delete)
			})

			// Articles
			r.Route("/articles", func(r chi.Router) {
				r.With(authMW.RequireScope("articles:read")).Get("/", h.Articles.List)
				r.With(authMW.RequireScope("articles:read")).Get("/{articleID}", h.Articles.Get)
				r.With(authMW.RequireScope("articles:write")).Put("/{articleID}/read-state", h.Articles.SetReadState)
			})

			// Folders
			r.Route("/folders", func(r chi.Router) {
				r.With(authMW.RequireScope("folders:read")).Get("/", h.Folders.List)
				r.With(authMW.RequireScope("folders:write")).Post("/", h.Folders.Create)
				r.With(authMW.RequireScope("folders:write")).Put("/{folderID}", h.Folders.Update)
				r.With(authMW.RequireScope("folders:delete")).Delete("/{folderID}", h.Folders.Delete)
			})

			// Users
			r.Route("/users", func(r chi.Router) {
				r.With(authMW.RequireScope("users:read")).Get("/me", h.Users.GetProfile)
				r.With(authMW.RequireScope("users:write")).Put("/me", h.Users.UpdateProfile)
			})
		})
	})
}
