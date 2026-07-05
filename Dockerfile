# syntax=docker/dockerfile:1

# ---- Stage 1: build the static Next.js frontend ----
FROM node:22-alpine AS web-build
WORKDIR /app/web
# Install deps first for layer caching.
COPY web/package.json web/package-lock.json ./
RUN npm ci
COPY web/ ./
# Produces /app/web/out (output: "export" in next.config.ts).
RUN npm run build

# ---- Stage 2: build the Go server ----
FROM golang:1.25-alpine AS server-build
WORKDIR /src
# Download modules first for layer caching.
COPY server/go.mod server/go.sum ./
RUN go mod download
COPY server/ ./
# Pure-Go deps, so static binary with cgo disabled.
RUN CGO_ENABLED=0 GOOS=linux go build -trimpath -ldflags="-s -w" -o /server ./cmd/server

# ---- Stage 3: minimal runtime ----
FROM alpine:3.20
# Non-root user.
RUN addgroup -S app && adduser -S app -G app
WORKDIR /app
COPY --from=server-build /server /app/server
COPY --from=web-build /app/web/out /app/web

ENV MEDIA_MERGE_ADDR=:8080 \
    MEDIA_MERGE_STATIC_DIR=/app/web \
    MEDIA_MERGE_WORKDIR=/tmp/media-merge

USER app
EXPOSE 8080
CMD ["/app/server"]
