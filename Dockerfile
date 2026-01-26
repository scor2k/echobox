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

# Create users for security separation
# echobox (UID 1001): Runs the application, writes audit logs (no shell access)
RUN addgroup -g 1001 echobox && \
    adduser -D -u 1001 -G echobox -s /bin/false echobox

# candidate (UID 1000): Interactive shell for tasks (cannot modify logs)
RUN addgroup -g 1000 candidate && \
    adduser -D -u 1000 -G candidate candidate

# Create necessary directories with proper ownership
# /output - owned by echobox (audit logs protected from candidate tampering)
RUN mkdir -p /output && \
    chown echobox:echobox /output && \
    chmod 755 /output

# /tasks - owned by root, read-only for all
RUN mkdir -p /tasks && \
    chmod 755 /tasks

# /home/candidate - owned by candidate (for task solutions)
RUN mkdir -p /home/candidate/solutions && \
    chown -R candidate:candidate /home/candidate

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
# /app owned by echobox (application user)
RUN chmod +x /app/echobox && \
    chown -R echobox:echobox /app

# Switch to application user (not candidate!)
# This ensures logs are owned by echobox, not candidate
USER echobox

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
