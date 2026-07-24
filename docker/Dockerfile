# syntax=docker/dockerfile:1

# =========================
# Stage 1 - Build
# =========================
FROM golang:1.25.1-alpine@sha256:b6ed3fd0452c0e9bcdef5597f29cc1418f61672e9d3a2f55bf02e7222c014abd AS builder

WORKDIR /src

RUN apk add --no-cache git ca-certificates

COPY go.mod go.sum ./

RUN --mount=type=cache,target=/go/pkg/mod \
    go mod download

COPY . .

RUN --mount=type=cache,target=/go/pkg/mod \
    --mount=type=cache,target=/root/.cache/go-build \
    CGO_ENABLED=0 \
    GOOS=linux \
    GOARCH=amd64 \
    go build \
        -trimpath \
        -ldflags="-s -w" \
        -o /out/api \
        ./cmd/api


# =========================
# Stage 2 - golang-migrate
# =========================
FROM migrate/migrate:v4.19.1 AS migrate


# =========================
# Stage 3 - Runtime
# =========================
FROM alpine:3.22@sha256:14358309a308569c32bdc37e2e0e9694be33a9d99e68afb0f5ff33cc1f695dce

WORKDIR /app

RUN apk add --no-cache \
    ca-certificates \
    postgresql-client \
    redis \
    tini

RUN addgroup -S app && adduser -S app -G app

COPY --from=builder /out/api ./api
COPY --from=builder /src/cmd/migrations ./migrations

# Копируем migrate из официального образа
COPY --from=migrate /usr/local/bin/migrate /usr/local/bin/migrate

COPY docker/docker-entrypoint.sh .

RUN chmod +x docker-entrypoint.sh && \
    chown -R app:app /app

USER app

EXPOSE 8080

HEALTHCHECK --interval=30s --timeout=5s --start-period=20s --retries=3 \
    CMD wget -q -O /dev/null http://127.0.0.1:8080/v1/health || exit 1

ENTRYPOINT ["/sbin/tini", "--", "/app/docker-entrypoint.sh"]