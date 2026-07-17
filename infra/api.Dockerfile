# syntax=docker/dockerfile:1.7
# Multi-stage build of the Go API. Produces a static binary with the
# migrations embedded (see apps/api/internal/database) — no runtime files to
# copy besides the binary itself.

# ---------- Build ----------
FROM golang:1.26-alpine AS build
RUN apk add --no-cache git
WORKDIR /src
# Separate layer for deps so they're cached across builds as long as go.mod/go.sum don't change.
COPY apps/api/go.mod apps/api/go.sum ./
RUN go mod download
COPY apps/api/ .
RUN CGO_ENABLED=0 GOOS=linux go build -trimpath -ldflags="-s -w" -o /bin/api ./cmd/api

# ---------- Runner ----------
FROM alpine:3.20 AS runner
RUN apk add --no-cache ca-certificates wget
COPY --from=build /bin/api /usr/local/bin/api

ENV API_PORT=8080
EXPOSE 8080

HEALTHCHECK --interval=30s --timeout=5s --start-period=10s --retries=3 \
  CMD wget -qO- http://localhost:${API_PORT}/health || exit 1

ENTRYPOINT ["/usr/local/bin/api"]
