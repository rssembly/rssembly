.PHONY: build build-all test lint vet clean dev dev-db migrate-up migrate-down \
        docker-build docker-up docker-down docker-logs fmt help

APP_NAME    := rssembly
BUILD_DIR   := dist
LDFLAGS     := -ldflags="-s -w"

# Detect OS for binary naming
ifeq ($(OS),Windows_NT)
	BINARY      := $(BUILD_DIR)/$(APP_NAME).exe
	MIGRATE_BIN := $(BUILD_DIR)/migrate.exe
else
	BINARY      := $(BUILD_DIR)/$(APP_NAME)
	MIGRATE_BIN := $(BUILD_DIR)/migrate
endif

# ── Build ──────────────────────────────────────────────────────────

build: go.sum
	go build $(LDFLAGS) -o $(BINARY) ./cmd/$(APP_NAME)

build-all: go.sum
	GOOS=linux   GOARCH=amd64 go build $(LDFLAGS) -o $(BUILD_DIR)/$(APP_NAME)-linux-amd64   ./cmd/$(APP_NAME)
	GOOS=linux   GOARCH=arm64 go build $(LDFLAGS) -o $(BUILD_DIR)/$(APP_NAME)-linux-arm64   ./cmd/$(APP_NAME)
	GOOS=darwin  GOARCH=amd64 go build $(LDFLAGS) -o $(BUILD_DIR)/$(APP_NAME)-darwin-amd64  ./cmd/$(APP_NAME)
	GOOS=darwin  GOARCH=arm64 go build $(LDFLAGS) -o $(BUILD_DIR)/$(APP_NAME)-darwin-arm64  ./cmd/$(APP_NAME)
	GOOS=windows GOARCH=amd64 go build $(LDFLAGS) -o $(BUILD_DIR)/$(APP_NAME)-windows-amd64.exe ./cmd/$(APP_NAME)

build-migrate: go.sum
	go build $(LDFLAGS) -o $(MIGRATE_BIN) ./cmd/migrate

# ── Dev / Test ─────────────────────────────────────────────────────

dev: build
	$(BINARY)

test: go.sum
	go test -race -shuffle=on -count=1 -coverprofile=coverage.out ./...
	@echo "---"
	@echo "Coverage:"
	@go tool cover -func=coverage.out | tail -1

lint:
	golangci-lint run ./...

vet:
	go vet ./...

fmt:
	go fmt ./...

clean:
	rm -rf $(BUILD_DIR) coverage.out coverage.html tmp

# ── Database ───────────────────────────────────────────────────────

MIGRATIONS_DIR := internal/database/migrations
DB_URL         ?= postgres://rssembly:rssembly@localhost:5432/rssembly?sslmode=disable

migrate-up: build-migrate
	$(MIGRATE_BIN) -dir $(MIGRATIONS_DIR) -db "$(DB_URL)" up

migrate-down: build-migrate
	$(MIGRATE_BIN) -dir $(MIGRATIONS_DIR) -db "$(DB_URL)" down 1

migrate-create:
	@read -p "Migration name: " name; \
	touch "$(MIGRATIONS_DIR)/$$(date -u +%Y%m%d%H%M%S)_$${name}.up.sql" \
	      "$(MIGRATIONS_DIR)/$$(date -u +%Y%m%d%H%M%S)_$${name}.down.sql"

# ── Docker ─────────────────────────────────────────────────────────

docker-build:
	docker build -t rssembly:latest .

docker-up:
	docker compose up -d --build

docker-down:
	docker compose down

docker-logs:
	docker compose logs -f

# ── Development ────────────────────────────────────────────────────

# Start Postgres via Docker (for local development with make run)
dev-db:
	@echo "Starting Postgres via Docker Compose..."
	docker compose up -d postgres
	@echo ""
	@echo "Postgres is running. Start the application with:"
	@echo "  make run"

# ── Help ───────────────────────────────────────────────────────────

help:
	@echo "RSSembly — Development Makefile"
	@echo ""
	@echo "Build:"
	@echo "  make build           Build binary for current OS"
	@echo "  make build-all       Cross-compile for all platforms"
	@echo "  make build-migrate   Build migration binary"
	@echo ""
	@echo "Test / Lint:"
	@echo "  make test            Run tests with race detector and coverage"
	@echo "  make lint            Run golangci-lint"
	@echo "  make vet             Run go vet"
	@echo "  make fmt             Format source code"
	@echo ""
	@echo "Run:"
	@echo "  make dev             Build and run locally"
	@echo "  make dev-db          Start Postgres in Docker (for local dev)"
	@echo ""
	@echo "Docker:"
	@echo "  make docker-build    Build Docker image"
	@echo "  make docker-up       Build and start all services"
	@echo "  make docker-down     Stop all services"
	@echo "  make docker-logs     Tail compose logs"
	@echo ""
	@echo "Database:"
	@echo "  make migrate-up      Run pending migrations"
	@echo "  make migrate-down    Rollback last migration"
	@echo "  make migrate-create  Create a new migration"
	@echo ""
	@echo "Other:"
	@echo "  make clean           Remove build artifacts"

# ── Misc ───────────────────────────────────────────────────────────

go.sum: go.mod
	go mod tidy