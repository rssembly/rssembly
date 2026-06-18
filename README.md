# RSSembly

> Take back control of your RSS reading.

**RSSembly** is an open-source, self-hostable RSS reader platform. One backend, multiple clients — web, browser extension, desktop, and mobile — all built with privacy in mind. No tracking, no telemetry, no AI slop.

## Status

🚧 **Pre-alpha** — under active development. Not ready for production use.

## Features (planned)

- 📡 Feed polling with adaptive scheduling (ETag, Last-Modified, jitter, backoff)
- 🗂️ Per-feed folders and organization
- 🔍 Full-text search via PostgreSQL tsvector
- 🚀 Real-time updates via WebSocket
- 🔑 JWT auth (Ed25519) + API keys for M2M
- 🐳 Single-binary Docker deployment (distroless)
- 📊 Prometheus metrics + OpenTelemetry tracing

## Quick start

```bash
cp .env.example .env       # configure your environment
make docker-up             # start with Docker Compose
```

The server will be available at `http://localhost:8080`.

## Architecture

```
├── cmd/rssembly/       # Server entrypoint
├── cmd/migrate/        # Database migration runner
├── internal/
│   ├── auth/           # JWT + API key authentication
│   ├── config/         # Configuration (env + YAML)
│   ├── database/       # PostgreSQL pool + migrations
│   ├── handler/        # HTTP handlers + router
│   ├── middleware/      # HTTP middleware (CORS, rate limit, auth, logging)
│   ├── models/         # Domain types + UUIDv7
│   ├── poller/         # Feed polling scheduler (future)
│   └── ws/             # WebSocket (future)
├── migrations/         # SQL migrations (future)
├── web/                # Frontend (future)
└── api/                # OpenAPI spec (future)
```

## License

MIT — see [LICENSE](LICENSE).