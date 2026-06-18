# rssembly — Project Conventions

## Architecture

- **Layout:** `cmd/` for entrypoints, `internal/` for private packages, `api/` for OpenAPI specs.
- **Monorepo:** Backend at root, frontend in `web/`, extension in `ext/` (future).
- **Config priority:** YAML config file > `.env` > environment variables > defaults.
- **API versioning:** All HTTP routes under `/api/v1/`. OpenAPI spec in `api/openapi.yaml` is the single source of truth — update it before adding or changing endpoints.

## Code style

- **Go version:** 1.26+, module `github.com/rssembly/rssembly`.
- **Tests:** Use stdlib `testing` + `github.com/stretchr/testify/require`. Test names: `TestPackage_Behavior`.
- **Error format:** `{"error":{"code":"snake_case_code","message":"human readable"}}`.
- **UUIDs:** `github.com/google/uuid` everywhere. All IDs are UUIDv7.
- **Logging:** `log/slog` with JSON handler. Structured keys only, no formatted strings.
- **Middleware pattern:** Chi-compatible `func(http.Handler) http.Handler`.

## Testing requirements

- Database tests: use `testcontainers-go` with a real PostgreSQL.
- API tests: `httptest.Server` with a Chi router.
- Race detector must pass: `go test -race -count=1 -shuffle=on ./...`.

## Making changes

1. Before adding a new route, update `api/openapi.yaml` first.
2. Before adding a new config field, add the env tag + yaml tag + `envDefault` if applicable.
3. Before adding a new dependency, run `go mod tidy` and ensure `go vet ./...` passes.

## Telemetry

- OTel tracer provider + stdout exporter wired in `internal/telemetry/`.
- `/metrics` endpoint serves Prometheus-scrapable metrics.
- All HTTP handlers should accept a `context.Context` for trace propagation.

## Handler conventions

- Use `internal/handler/` for HTTP handlers, `internal/repo/` for database queries.
- Handler structs receive dependencies via constructor injection (DB pool, JWT manager).
- All handlers use `Respond`/`RespondError` from `internal/handler/response.go` — never write raw JSON.
- Route registration in `internal/handler/routes.go` with per-route `RequireScope` checks.
- Tests use `httptest.Server` with a Chi router and real middleware stack.

## Repository conventions

- `internal/repo/` packages own all SQL queries, never in handlers.
- Methods accept `context.Context` as the first parameter.
- Return `(T, error)` or `(nil, ErrNotFound)` for missing rows, never `nil, nil`.

## Auth conventions

- Passwords: argon2id hashing via `auth.HashPassword` / `auth.VerifyPassword`.
- JWT: Ed25519 signing via `auth.JWTManager`.
- New users get scopes `["feeds:read", "feeds:write", "articles:read", "articles:write", "folders:read", "folders:write"]` by default.