# Multi-stage build for Prism (SvelteKit static + Go server)
# 1. Frontend build stage
FROM node:25-alpine AS frontend
WORKDIR /app

# Only copy package manifests first for better layer caching
COPY web/package.json web/package-lock.json ./web/
WORKDIR /app/web
RUN npm ci --include=dev

# Copy rest of frontend source
COPY web/ .
# Build static site (outputs to web/build via adapter-static config)
RUN npm run build

# 2. Go build stage
FROM golang:1.25.0-alpine AS gobuilder
# Cross-compilation setup
ARG TARGETOS=linux
ARG TARGETARCH=amd64

WORKDIR /go/api/

# Install build dependencies
RUN apk add --update gcc musl-dev ca-certificates --no-cache

# Enable Go modules and caching
COPY api/go.mod api/go.sum ./

# Download dependencies with verify
RUN go mod download && go mod verify

# Copy Go source
COPY api/ .
# Copy built frontend from previous stage into expected path
COPY --from=frontend /app/web/build ./web/build

# Build a static binary with CGO (required for SQLite)
# Note: Building for native platform to avoid cross-compilation complexities with CGO
RUN CGO_ENABLED=1 go build \
    -a \
    -ldflags '-linkmode external -extldflags "-static"  -w -s' \
    -tags 'sqlite_omit_load_extension sqlite_fts5' \
    -o /go/bin/prism .

# 3. Final runtime stage (alpine for CGO binary support)
FROM scratch AS runtime
LABEL org.opencontainers.image.source="https://github.com/NorskHelsenett/prism" \
      org.opencontainers.image.title="Prism" \
      org.opencontainers.image.description="PRISM - Pentest Report Information Security Management" \
      maintainer="NorskHelsenett <https://github.com/NorskHelsenett/prism>"

ENV GIN_MODE=release
ENV CONFIG_PATH=/config
ENV ROLES_PATH=/app/roles.yaml

WORKDIR /app

# Copy in SSL root certificates so outbound HTTPS works
COPY --from=gobuilder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/ca-certificates.crt

# Copy statically linked binary and prerendered assets
COPY --from=gobuilder /go/bin/prism /app/prism
COPY --from=gobuilder /go/api/roles.yaml /app/roles.yaml
COPY --from=gobuilder /go/api/web/build /app/web/build

EXPOSE 8080
USER 1000

# No healthcheck (scratch has no shell/tools); rely on external monitoring.
ENTRYPOINT ["/app/prism"]
