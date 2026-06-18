# rssembly

> Take back control of your RSS reading.

![rssembly logo](assets/rssembly-logo.svg)

**rssembly** is an open-source, self-hostable RSS reader platform. One backend, multiple clients — web, browser extension, desktop, and mobile — all built with privacy in mind. No tracking, no telemetry, no AI slop.

## Status

🚧 **Pre-alpha** — under active development. Not ready for production use.

## Features

- ✅ User registration & authentication (JWT Ed25519, argon2id)
- ✅ API key authentication with granular path-based scoping
- ✅ OpenAPI 3.1 spec (single source of truth for the API contract)
- ✅ PostgreSQL-backed with auto-migrations on startup
- ✅ Configuration via YAML > `.env` > env vars (priority chain)
- ✅ Prometheus `/metrics` endpoint + OpenTelemetry tracing
- ✅ Rate limiting, CORS, structured logging (slog)
- 🔜 Feed polling with adaptive scheduling (ETag, Last-Modified, jitter, backoff)
- 🔜 Per-feed folders and organization
- 🔜 Full-text search via PostgreSQL tsvector
- 🔜 Real-time updates via WebSocket
- 🔜 Web, browser extension, desktop, and mobile clients

## Quick start

```bash
cp .env.example .env       # configure your environment
make docker-up             # start with Docker Compose
```

The server will be available at `http://localhost:8080`.

JWT keys are auto-generated on first startup — no manual setup required.

## Architecture

```
├── cmd/
│   ├── rssembly/          # Server entrypoint
│   └── migrate/           # Database migration runner
├── internal/
│   ├── auth/              # JWT, API keys, argon2id, AES-GCM, scope matching
│   ├── config/            # Configuration (YAML > .env > env vars > defaults)
│   ├── database/          # PostgreSQL connection pool + SQL migrations
│   ├── handler/           # HTTP handlers, response helpers, route registration
│   ├── middleware/        # Auth, CORS, rate limiting, logging, recovery
│   ├── models/            # Domain types (User, Feed, Article, APIKey, Setting)
│   ├── repo/              # Database query layer (UserRepo, etc.)
│   ├── poller/            # Feed polling scheduler (future)
│   ├── settings/          # Global & per-user settings resolution
│   ├── telemetry/         # OpenTelemetry tracing + metrics
│   └── ws/                # WebSocket (future)
├── api/
│   └── openapi.yaml       # OpenAPI 3.1 spec (single source of truth)
├── assets/                # Project logos and branding
└── web/                   # Frontend (future)
```

## Configuration

Priority order: YAML config file (`-config`) > `.env` file > environment variables > defaults.

Essential env vars:

| Variable | Default | Description |
|---|---|---|
| `DATABASE_URL` | — | PostgreSQL connection string (required) |
| `SERVER_PORT` | `8080` | HTTP listen port |
| `JWT_PRIVATE_KEY` | — | Inline Ed25519 private key PEM (beats file paths) |
| `JWT_PUBLIC_KEY` | — | Inline Ed25519 public key PEM |
| `ENCRYPTION_KEY` | — | 64-char hex key for AES-256-GCM feed password encryption |
| `LOG_LEVEL` | `info` | One of: debug, info, warn, error |

## Development

```bash
make build      # build the server binary
make test       # run all tests with race detector
make lint       # golangci-lint
make dev        # build and run
```

## License

MIT — see [LICENSE](LICENSE).