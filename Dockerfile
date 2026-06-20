# ── Build stage ─────────────────────────────────────────────────────
FROM golang:1.26-alpine AS builder

RUN apk add --no-cache ca-certificates

WORKDIR /src

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 go build -ldflags="-s -w" -o /dist/rssembly    ./cmd/rssembly
RUN CGO_ENABLED=0 go build -ldflags="-s -w" -o /dist/healthcheck ./cmd/healthcheck

# ── Runtime stage ──────────────────────────────────────────────────
FROM gcr.io/distroless/static-debian12:nonroot

COPY --from=builder /dist/rssembly    /rssembly
COPY --from=builder /dist/healthcheck /healthcheck

EXPOSE 8080

ENTRYPOINT ["/rssembly"]
