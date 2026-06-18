# ── Build stage ─────────────────────────────────────────────────────
FROM golang:1.26-alpine AS builder

RUN apk add --no-cache gcc musl-dev

WORKDIR /src

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 go build -ldflags="-s -w" -o /dist/rssembly ./cmd/rssembly
RUN CGO_ENABLED=0 go build -ldflags="-s -w" -o /dist/migrate   ./cmd/migrate

# ── Runtime stage ──────────────────────────────────────────────────
FROM gcr.io/distroless/static-debian12:nonroot

COPY --from=builder /dist/rssembly /rssembly
COPY --from=builder /dist/migrate   /migrate

EXPOSE 8080

ENTRYPOINT ["/rssembly"]