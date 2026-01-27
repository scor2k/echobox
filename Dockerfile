# Multi-stage build for Echobox SRE Interview Terminal
# Stage 1: Builder
FROM golang:1.24-alpine AS builder

# Install build dependencies
RUN apk add --no-cache git make

# Set working directory
WORKDIR /build

# Copy go mod files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy source code
COPY cmd/ ./cmd/
COPY internal/ ./internal/
COPY web/ ./web/

# Build the binary
# Use TARGETARCH build arg to support multi-platform builds
ARG TARGETARCH=arm64
RUN CGO_ENABLED=0 GOOS=linux GOARCH=${TARGETARCH} go build \
    -ldflags="-s -w -X main.Version=${VERSION:-dev}" \
    -o echobox \
    ./cmd/server

# Stage 2: Runtime
FROM alpine:3.19

# Install runtime dependencies
RUN apk add --no-cache \
    bash \
    ca-certificates \
    curl \
    vim \
    nano \
    jq \
    htop \
    net-tools \
    procps \
    util-linux \
    && rm -rf /var/cache/apk/*

# Create base directories
# Note: Home directories for random UIDs created dynamically by app
RUN mkdir -p /output /tasks /home && \
    chmod 755 /tasks && \
    chmod 777 /home

# Set working directory
WORKDIR /app

# Copy binary from builder
COPY --from=builder /build/echobox /app/echobox

# Copy web assets
COPY web/ /app/web/

# Verify web assets copied correctly
RUN ls -la /app/web/ && \
    ls -la /app/web/vendor/ && \
    echo "Web assets copied successfully"

# Set permissions
RUN chmod +x /app/echobox

# NOTE: App runs as ROOT to enable UID isolation
# Each shell gets a random UID (10000-60000)
# Logs stay owned by root (UID 0)
# Shell user (random UID) cannot modify root-owned logs
# This provides isolation even with shared volumes

# Expose port
EXPOSE 8080

# Environment variables with defaults
ENV PORT=8080 \
    SESSION_TIMEOUT=7200 \
    CANDIDATE_NAME=anonymous \
    OUTPUT_DIR=/output \
    SHELL=/bin/bash \
    RECONNECT_WINDOW=300 \
    INPUT_RATE_LIMIT=30 \
    FLUSH_INTERVAL=10 \
    LOG_LEVEL=info

# Health check
HEALTHCHECK --interval=30s --timeout=5s --start-period=5s --retries=3 \
    CMD wget --no-verbose --tries=1 --spider http://localhost:8080/health || exit 1

# Run the application
CMD ["/app/echobox"]
