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

# Create non-root user
RUN addgroup -g 1000 candidate && \
    adduser -D -u 1000 -G candidate candidate

# Create necessary directories
RUN mkdir -p /output /tasks /home/candidate/solutions && \
    chown -R candidate:candidate /output /home/candidate

# Set working directory
WORKDIR /app

# Copy binary from builder
COPY --from=builder /build/echobox /app/echobox

# Copy web assets
COPY web/ /app/web/

# Set permissions
RUN chmod +x /app/echobox && \
    chown -R candidate:candidate /app

# Switch to non-root user
USER candidate

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
