package middleware

import (
	"log/slog"
	"net/http"
	"strings"
	"sync"
	"time"
)

// responseWriter wraps http.ResponseWriter to capture the status code.
type responseWriter struct {
	http.ResponseWriter
	status int
}

func (rw *responseWriter) WriteHeader(code int) {
	rw.status = code
	rw.ResponseWriter.WriteHeader(code)
}

// Logging returns middleware that logs each request via slog.
func Logging(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		rw := &responseWriter{ResponseWriter: w, status: http.StatusOK}

		next.ServeHTTP(rw, r)

		slog.Info("request",
			"method", r.Method,
			"path", r.URL.Path,
			"status", rw.status,
			"duration", time.Since(start).String(),
			"remote", r.RemoteAddr,
			"user_agent", r.UserAgent(),
		)
	})
}

// Recoverer recovers from panics and returns a 500 so the server stays up.
func Recoverer(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if rec := recover(); rec != nil {
				slog.Error("panic recovered",
					"path", r.URL.Path,
					"recover", rec,
				)
				http.Error(w, `{"error":{"code":"internal","message":"internal server error"}}`, http.StatusInternalServerError)
			}
		}()
		next.ServeHTTP(w, r)
	})
}

// CORSMiddleware returns middleware that handles CORS based on the allowed origins list.
// If allowedOrigins is empty or contains "*", all origins are allowed (self-hosted default).
func CORSMiddleware(allowedOrigins []string) func(http.Handler) http.Handler {
	allowAll := len(allowedOrigins) == 0
	for _, o := range allowedOrigins {
		if o == "*" {
			allowAll = true
			break
		}
	}

	originSet := make(map[string]bool, len(allowedOrigins))
	for _, o := range allowedOrigins {
		originSet[o] = true
	}

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			origin := r.Header.Get("Origin")

			if allowAll || originSet[origin] {
				w.Header().Set("Access-Control-Allow-Origin", origin)
				w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, PATCH, DELETE, OPTIONS")
				w.Header().Set("Access-Control-Allow-Headers", "Authorization, Content-Type, X-API-Key")
				w.Header().Set("Access-Control-Allow-Credentials", "true")
				w.Header().Set("Access-Control-Max-Age", "86400")
			}

			if r.Method == http.MethodOptions {
				w.WriteHeader(http.StatusNoContent)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

// RateLimiter enforces a per-IP rate limit using an in-memory counter.
type RateLimiter struct {
	mu       sync.Mutex
	requests int
	window   time.Duration
	ips      map[string]*bucket
}

type bucket struct {
	count   int
	resetAt time.Time
}

func NewRateLimiter(requests int, window time.Duration) *RateLimiter {
	return &RateLimiter{
		requests: requests,
		window:   window,
		ips:      make(map[string]*bucket),
	}
}

func (rl *RateLimiter) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ip := r.RemoteAddr
		// Strip port if present.
		if idx := strings.LastIndex(ip, ":"); idx != -1 {
			ip = ip[:idx]
		}

		rl.mu.Lock()
		b, exists := rl.ips[ip]
		if !exists || time.Now().After(b.resetAt) {
			rl.ips[ip] = &bucket{count: 1, resetAt: time.Now().Add(rl.window)}
			rl.mu.Unlock()
			next.ServeHTTP(w, r)
			return
		}

		b.count++
		if b.count > rl.requests {
			rl.mu.Unlock()
			w.Header().Set("Retry-After", rl.window.String())
			http.Error(w, `{"error":{"code":"rate_limited","message":"too many requests"}}`, http.StatusTooManyRequests)
			return
		}
		rl.mu.Unlock()
		next.ServeHTTP(w, r)
	})
}
