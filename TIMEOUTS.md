# Timeout Configuration

This document describes the various timeouts in the application.

## Graceful Shutdown Timeouts

### SIGTERM/SIGINT (Ctrl+C) Shutdown
**Location**: `cmd/server/main.go:101-109`
```go
ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
```
- **Total Timeout**: ~6 seconds (500ms PTY close + 5s server shutdown)
- **Trigger**: SIGTERM or SIGINT signal (Ctrl+C, `kill`, `docker stop`)
- **Behavior**:
  1. Closes PTY (kills shell process immediately)
  2. Waits 500ms for WebSocket connections to close
  3. Stops accepting new HTTP connections
  4. Waits up to 5s for active HTTP connections to complete
  5. Exits

### Finish Button Shutdown
**Location**: `cmd/server/main.go:71-84`
```go
ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
```
- **Total Timeout**: ~4 seconds (500ms PTY close + 500ms wait + 3s server shutdown)
- **Trigger**: User clicks "Finish Session" button
- **Behavior**:
  1. Closes PTY (kills shell process immediately)
  2. Waits 500ms for WebSocket connections to close
  3. Stops accepting new HTTP connections
  4. Waits up to 3s for active HTTP connections to complete
  5. Exits with code 0

## HTTP Server Timeouts

**Location**: `internal/web/server.go:42-47`
```go
httpServer := &http.Server{
    ReadTimeout:  15 * time.Second,
    WriteTimeout: 15 * time.Second,
    IdleTimeout:  60 * time.Second,
}
```

- **ReadTimeout**: 15 seconds - Max time to read entire request
- **WriteTimeout**: 15 seconds - Max time to write response
- **IdleTimeout**: 60 seconds - Max time keep-alive connection stays open

## Session Timeout (Future)

Currently configured via environment variable but not enforced:
- **Variable**: `SESSION_TIMEOUT`
- **Default**: 7200 seconds (2 hours)
- **Status**: Configuration exists, enforcement pending (Phase 2)

## Customizing Timeouts

To change shutdown timeouts, edit:
- `cmd/server/main.go:101` - SIGTERM/SIGINT HTTP shutdown timeout (currently 5s)
- `cmd/server/main.go:71` - Finish button HTTP shutdown timeout (currently 3s)
- `cmd/server/main.go:98` and `cmd/server/main.go:71` - WebSocket close wait (currently 500ms both)

To change HTTP timeouts, edit:
- `internal/web/server.go:42-47`

Example for faster shutdown (useful for development):
```go
// Very fast shutdown (1 second)
ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
```

## Troubleshooting

If shutdown hangs:
1. Check logs for "PTY: Close complete" - confirms PTY closed successfully
2. Check for "HTTP server shutdown complete" - confirms server closed
3. If stuck at "Closing PTY", shell process may be unkillable (rare)
4. If stuck at "Shutting down HTTP server", active connections not closing

The current implementation should never hang because:
- PTY close is immediate (SIGKILL)
- WebSocket reads unblock when PTY closes
- HTTP shutdown has a timeout (force-closes after timeout)

## PTY Close Behavior

**Location**: `internal/terminal/pty.go:145-178`

When PTY closes:
1. Closes file descriptor (non-blocking)
2. Sends SIGKILL to shell process immediately
3. Waits for process in background goroutine (non-blocking)

This ensures shutdown never hangs waiting for shell to exit.

## Docker Considerations

When running in Docker:
- `docker stop` sends SIGTERM, waits 10s (default), then SIGKILL
- Our shutdown completes in ~6 seconds, well within Docker's default
- No need to customize `--stop-timeout` unless you want faster SIGKILL

Example (works with defaults):
```bash
docker run echobox:latest
docker stop <container_id>  # Completes in ~6 seconds
```
