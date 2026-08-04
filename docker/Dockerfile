# syntax=docker/dockerfile:1

# =========================
# Stage 1 - Build
# =========================
FROM golang:1.25.7-alpine@sha256:f6751d823c26342f9506c03797d2527668d095b0a15f1862cddb4d927a7a4ced AS builder

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
FROM golang:1.25.7-alpine AS migrate-builder

RUN apk add --no-cache git

RUN git clone --depth 1 --branch v4.19.1 \
      https://github.com/golang-migrate/migrate.git /src/migrate

WORKDIR /src/migrate

RUN go get github.com/jackc/pgx/v5@v5.9.0 google.golang.org/grpc@v1.79.3 && \
    go mod tidy && \
    CGO_ENABLED=0 GOOS=linux GOARCH=amd64 \
    go build -tags 'postgres' -trimpath -ldflags="-s -w" -o /out/migrate ./cmd/migrate

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
COPY --from=migrate-builder /out/migrate /usr/local/bin/migrate

COPY docker/docker-entrypoint.sh .

RUN chmod +x docker-entrypoint.sh && \
    chown -R app:app /app

USER app

EXPOSE 8080

HEALTHCHECK --interval=30s --timeout=5s --start-period=20s --retries=3 \
    CMD wget -q -O /dev/null http://127.0.0.1:8080/v1/health || exit 1

ENTRYPOINT ["/sbin/tini", "--", "/app/docker-entrypoint.sh"]