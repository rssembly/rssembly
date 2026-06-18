package main

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/go-chi/chi/v5"
	chiMiddleware "github.com/go-chi/chi/v5/middleware"

	"github.com/rssembly/rssembly/internal/auth"
	"github.com/rssembly/rssembly/internal/config"
	"github.com/rssembly/rssembly/internal/database"
	"github.com/rssembly/rssembly/internal/handler"
	"github.com/rssembly/rssembly/internal/middleware"
	"github.com/rssembly/rssembly/internal/telemetry"
)

func main() {
	slog.SetDefault(slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo})))

	cfg, err := config.Load()
	if err != nil {
		slog.Error("failed to load config", "error", err)
		os.Exit(1)
	}

	// Parse log level.
	var logLevel slog.Level
	if err := logLevel.UnmarshalText([]byte(cfg.LogLevel)); err != nil {
		logLevel = slog.LevelInfo
	}
	slog.SetDefault(slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: logLevel})))

	// ── Database ──────────────────────────────────────────────────────
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	db, err := database.Connect(ctx, cfg.DatabaseURL)
	cancel()
	if err != nil {
		slog.Error("failed to connect to database", "error", err)
		os.Exit(1)
	}
	defer db.Close()
	slog.Info("connected to database")

	// Run migrations on startup.
	if err := database.RunMigrations(cfg.DatabaseURL); err != nil {
		slog.Error("failed to run migrations", "error", err)
		os.Exit(1)
	}

	// ── Telemetry ─────────────────────────────────────────────────────
	shutdownTelemetry, err := telemetry.Init("rssembly", "0.1.0")
	if err != nil {
		slog.Error("failed to init telemetry", "error", err)
		os.Exit(1)
	}
	defer shutdownTelemetry()
	slog.Info("telemetry initialized")

	// ── Auth: JWT Key Resolution ──────────────────────────────────────
	// Three-tier fallback:
	//   1. Inline PEM env vars (JWT_PRIVATE_KEY / JWT_PUBLIC_KEY)
	//   2. PEM files at configured paths
	//   3. Auto-generate and persist to default paths
	jwtManager, err := initJWT(cfg)
	if err != nil {
		slog.Error("failed to initialize JWT manager", "error", err)
		os.Exit(1)
	}
	slog.Info("JWT manager initialized")

	// ── Auth Middleware ─────────────────────────────────────────────
	// The JWT manager handles JWT verification. API key lookup is a stub
	// until the database layer is fully wired up.
	authHandler := &compositeAuth{
		JWT: jwtManager,
		APIKeyLookupFn: func(ctx context.Context, prefix, hash string) (*auth.AuthenticatedUser, error) {
			return nil, errors.New("API key lookup not yet implemented")
		},
	}
	authMiddleware := middleware.NewAuth(authHandler)

	// ── Router ────────────────────────────────────────────────────────
	r := chi.NewRouter()

	// Global middleware.
	r.Use(middleware.Recoverer)
	r.Use(middleware.Logging)
	r.Use(chiMiddleware.RequestID)
	r.Use(middleware.CORSMiddleware(cfg.CORSAllowedOrigins))
	rl := middleware.NewRateLimiter(cfg.RateLimitRequests, cfg.RateLimitWindow)
	r.Use(rl.Middleware)

	// Health endpoints (no auth required).
	healthHandler := handler.NewHealthHandler(db)
	r.Get("/health", healthHandler.Liveness)
	r.Get("/ready", healthHandler.Readiness)

	// Metrics (root-level — standard Prometheus scrape path).
	r.Get("/metrics", telemetry.MetricsHandler().ServeHTTP)

	// ── API v1 ────────────────────────────────────────────────────────
	r.Route("/api/v1", func(r chi.Router) {
		// Public endpoints.
		r.Post("/auth/login", func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusNotImplemented)
			w.Write([]byte(`{"error":{"code":"not_implemented","message":"login not yet implemented"}}`))
		})
		r.Post("/auth/register", func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusNotImplemented)
			w.Write([]byte(`{"error":{"code":"not_implemented","message":"registration not yet implemented"}}`))
		})

		// Authenticated endpoints.
		r.Group(func(r chi.Router) {
			r.Use(authMiddleware.Middleware)

			r.Get("/feeds", notImplemented)
			r.Post("/feeds", notImplemented)
			r.Get("/feeds/{feedID}", notImplemented)
			r.Put("/feeds/{feedID}", notImplemented)
			r.Delete("/feeds/{feedID}", notImplemented)

			r.Get("/articles", notImplemented)
			r.Get("/articles/{articleID}", notImplemented)
			r.Put("/articles/{articleID}/read-state", notImplemented)

			r.Get("/folders", notImplemented)
			r.Post("/folders", notImplemented)
			r.Put("/folders/{folderID}", notImplemented)
			r.Delete("/folders/{folderID}", notImplemented)

			r.Get("/users/me", notImplemented)
			r.Put("/users/me", notImplemented)
		})
	})

	// ── Server ────────────────────────────────────────────────────────
	srv := &http.Server{
		Addr:         cfg.Addr,
		Handler:      r,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 30 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// Graceful shutdown.
	shutdown := make(chan os.Signal, 1)
	signal.Notify(shutdown, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		slog.Info("starting server", "addr", cfg.Addr)
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			slog.Error("server error", "error", err)
			os.Exit(1)
		}
	}()

	<-shutdown
	slog.Info("shutting down server...")

	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer shutdownCancel()

	if err := srv.Shutdown(shutdownCtx); err != nil {
		slog.Error("server shutdown error", "error", err)
		os.Exit(1)
	}
	slog.Info("server stopped")
}

func notImplemented(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusNotImplemented)
	w.Write([]byte(`{"error":{"code":"not_implemented","message":"this endpoint is not yet implemented"}}`))
}

// compositeAuth implements middleware.AuthenticationHandler by combining JWT
// verification with an API key lookup function.
type compositeAuth struct {
	JWT            *auth.JWTManager
	APIKeyLookupFn func(ctx context.Context, prefix, hash string) (*auth.AuthenticatedUser, error)
}

func (c *compositeAuth) VerifyToken(tokenString string) (*auth.AuthenticatedUser, error) {
	return c.JWT.VerifyToken(tokenString)
}

func (c *compositeAuth) APIKeyLookup(ctx context.Context, prefix, hash string) (*auth.AuthenticatedUser, error) {
	return c.APIKeyLookupFn(ctx, prefix, hash)
}

// initJWT resolves JWT signing keys through a three-tier fallback chain.
func initJWT(cfg *config.Config) (*auth.JWTManager, error) {
	// Tier 1: Inline PEM from environment variables.
	if cfg.JWTPrivateKey != "" && cfg.JWTPublicKey != "" {
		mgr, err := auth.NewJWTManagerFromPEM(
			[]byte(cfg.JWTPrivateKey),
			[]byte(cfg.JWTPublicKey),
		)
		if err == nil {
			slog.Info("jwt: using inline PEM from environment variables")
			return mgr, nil
		}
		slog.Warn("jwt: inline PEM env vars present but invalid, falling back", "error", err)
	}

	// Tier 2: PEM files on disk.
	if privPEM, err := os.ReadFile(cfg.JWTPrivateKeyPath); err == nil {
		if pubPEM, err := os.ReadFile(cfg.JWTPublicKeyPath); err == nil {
			mgr, err := auth.NewJWTManagerFromPEM(privPEM, pubPEM)
			if err == nil {
				slog.Info("jwt: loaded keys from files",
					"private_key", cfg.JWTPrivateKeyPath,
					"public_key", cfg.JWTPublicKeyPath,
				)
				return mgr, nil
			}
			slog.Warn("jwt: key files present but invalid, regenerating", "error", err)
		}
	}

	// Tier 3: Auto-generate and persist to default paths.
	slog.Info("jwt: no keys found, generating new Ed25519 key pair")
	if err := auth.GenerateKeyPair(cfg.JWTPrivateKeyPath, cfg.JWTPublicKeyPath); err != nil {
		return nil, fmt.Errorf("generate and persist key pair: %w", err)
	}

	privPEM, err := os.ReadFile(cfg.JWTPrivateKeyPath)
	if err != nil {
		return nil, fmt.Errorf("read generated private key: %w", err)
	}
	pubPEM, err := os.ReadFile(cfg.JWTPublicKeyPath)
	if err != nil {
		return nil, fmt.Errorf("read generated public key: %w", err)
	}

	mgr, err := auth.NewJWTManagerFromPEM(privPEM, pubPEM)
	if err != nil {
		return nil, fmt.Errorf("load generated key pair: %w", err)
	}

	slog.Info("jwt: generated and persisted new key pair",
		"private_key", cfg.JWTPrivateKeyPath,
		"public_key", cfg.JWTPublicKeyPath,
	)
	return mgr, nil
}