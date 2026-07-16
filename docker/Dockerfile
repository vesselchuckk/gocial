# syntax=docker/dockerfile:1

FROM golang:1.25.1-alpine AS builder

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


FROM alpine:3.20

WORKDIR /app

RUN apk add --no-cache \
    ca-certificates \
    postgresql-client \
    redis \
    tini \
    wget \
    tar

RUN wget -O /tmp/migrate.tar.gz \
    https://github.com/golang-migrate/migrate/releases/download/v4.18.3/migrate.linux-amd64.tar.gz && \
    tar -xzf /tmp/migrate.tar.gz -C /usr/local/bin && \
    chmod +x /usr/local/bin/migrate && \
    rm /tmp/migrate.tar.gz

RUN addgroup -S app && adduser -S app -G app

COPY --from=builder /out/api ./api
COPY --from=builder /src/cmd/migrations ./migrations
COPY docker/docker-entrypoint.sh .

RUN chmod +x docker-entrypoint.sh && \
    chown -R app:app /app

USER app

EXPOSE 8080

HEALTHCHECK --interval=30s --timeout=5s --start-period=20s \
    CMD wget -q -O /dev/null http://localhost:8080/v1/health || exit 1

ENTRYPOINT ["/sbin/tini", "--", "/app/docker-entrypoint.sh"]
