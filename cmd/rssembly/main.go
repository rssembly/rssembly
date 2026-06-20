// Package main is the RSSembly server entrypoint. It wires config, database,
// JWT auth, telemetry, all HTTP handlers, and starts the HTTP server.
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
	"github.com/rssembly/rssembly/internal/poller"
	"github.com/rssembly/rssembly/internal/repo"
	"github.com/rssembly/rssembly/internal/telemetry"
	"github.com/rssembly/rssembly/internal/ws"
)

func main() {
	slog.SetDefault(slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo})))

	cfg, err := config.Load()
	if err != nil {
		slog.Error("failed to load config", "error", err)
		os.Exit(1)
	}

	var logLevel slog.Level
	if err := logLevel.UnmarshalText([]byte(cfg.LogLevel)); err != nil {
		logLevel = slog.LevelInfo
	}
	slog.SetDefault(slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: logLevel})))

	// Database
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	db, err := database.Connect(ctx, cfg.DatabaseURL)
	cancel()
	if err != nil {
		slog.Error("failed to connect to database", "error", err)
		os.Exit(1)
	}
	defer db.Close()
	slog.Info("connected to database")

	if err := database.RunMigrations(cfg.DatabaseURL); err != nil {
		slog.Error("failed to run migrations", "error", err)
		os.Exit(1)
	}

	// Telemetry
	shutdownTelemetry, err := telemetry.Init("rssembly", "0.1.0")
	if err != nil {
		slog.Error("failed to init telemetry", "error", err)
		os.Exit(1)
	}
	defer shutdownTelemetry()
	slog.Info("telemetry initialized")

	// Auth: JWT Key Resolution
	jwtManager, err := initJWT(cfg)
	if err != nil {
		slog.Error("failed to initialize JWT manager", "error", err)
		os.Exit(1)
	}
	slog.Info("JWT manager initialized")

	// Auth Middleware
	apiKeyRepo := repo.NewAPIKeyRepo(db)
	authHandler := &compositeAuth{
		JWT: jwtManager,
		APIKeyLookupFn: func(ctx context.Context, prefix, hash string) (*auth.AuthenticatedUser, error) {
			key, err := apiKeyRepo.GetAPIKeyByPrefix(ctx, prefix)
			if err != nil {
				return nil, fmt.Errorf("api key lookup: %w", err)
			}
			return &auth.AuthenticatedUser{
				UserID:   key.CreatedBy,
				Scopes:   key.Scopes,
				IsAPIKey: true,
			}, nil
		},
	}
	authMiddleware := middleware.NewAuth(authHandler)

	// WebSocket Hub
	wsHub := ws.NewHub()
	wsHub.SetAuthenticator(ws.NewJWTAdapter(jwtManager))

	// Router
	r := chi.NewRouter()
	r.Use(middleware.Recoverer)
	r.Use(middleware.Logging)
	r.Use(chiMiddleware.RequestID)
	r.Use(middleware.CORSMiddleware(cfg.CORSAllowedOrigins))
	rl := middleware.NewRateLimiter(cfg.RateLimitRequests, cfg.RateLimitWindow)
	r.Use(rl.Middleware)

	handler.RegisterRoutes(r, &handler.Handlers{
		Auth:     handler.NewAuthHandler(db, jwtManager),
		Feeds:    handler.NewFeedHandler(repo.NewFeedRepo(db)),
		Articles: handler.NewArticleHandler(repo.NewArticleRepo(db)),
		Folders:  handler.NewFolderHandler(repo.NewFolderRepo(db)),
		Users:    handler.NewUserHandler(repo.NewUserRepo(db)),
		Health:   handler.NewHealthHandler(db),
		WSHub:    wsHub,
	}, authMiddleware, telemetry.MetricsHandler().ServeHTTP, wsHub)

	// Feed Poller
	pollInterval := cfg.DefaultPollInterval
	if pollInterval <= 0 {
		pollInterval = 30 * time.Second
	}
	feedPoller := poller.New(repo.NewFeedRepo(db), repo.NewArticleRepo(db))
	go func() {
		slog.Info("starting feed poller", "check_interval", pollInterval)
		ticker := time.NewTicker(pollInterval)
		defer ticker.Stop()

		feeds, articles, err := feedPoller.PollOnce(context.Background())
		if err != nil {
			slog.Warn("initial poll failed", "error", err)
		} else {
			slog.Info("initial poll complete", "feeds", feeds, "articles", articles)
		}

		for range ticker.C {
			feeds, articles, err := feedPoller.PollOnce(context.Background())
			if err != nil {
				slog.Warn("poll cycle failed", "error", err)
				continue
			}
			if feeds > 0 || articles > 0 {
				slog.Info("poll cycle complete", "feeds", feeds, "articles", articles)
			}
		}
	}()

	// Server
	srv := &http.Server{
		Addr:         cfg.Addr,
		Handler:      r,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 30 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

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

func initJWT(cfg *config.Config) (*auth.JWTManager, error) {
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
