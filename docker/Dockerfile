# syntax=docker/dockerfile:1

FROM golang:1.24-alpine AS builder
WORKDIR /src

RUN apk add --no-cache git ca-certificates

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -trimpath -ldflags='-s -w' -o /out/api ./cmd/api

FROM alpine:3.20
WORKDIR /app

RUN apk add --no-cache postgresql-client redis

COPY --from=builder /out/api /app/api
COPY --from=builder /src/cmd/migrations /app/migrations
COPY docker-entrypoint.sh /app/docker-entrypoint.sh

RUN chmod +x /app/docker-entrypoint.sh

ENV SERVER_HOST=0.0.0.0 \
    SERVER_PORT=8080 \
    DB_HOST=db \
    DB_PORT=5432 \
    DB_USER=admin \
    DB_PASSWORD=qwerty \
    DB_NAME=gosocial \
    DB_URL=postgres://admin:qwerty@db:5432/gosocial?sslmode=disable \
    REDIS_ENABLED=true \
    REDIS_ADDR=redis:6379 \
    REDIS_PASSWORD=my-password \
    REDIS_DB=0 \
    ENV=production

EXPOSE 8080

ENTRYPOINT ["/app/docker-entrypoint.sh"]
