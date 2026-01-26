# Echobox - SRE Interview Terminal

A secure, web-based terminal environment for conducting technical SRE interviews with comprehensive session recording and anti-cheat measures.

## Features

### Completed

**Phase 1: Core Terminal**
- ✅ Web-based terminal using xterm.js
- ✅ WebSocket-based PTY bridge
- ✅ Real-time terminal interaction
- ✅ Clean, professional UI
- ✅ Connection status indicators
- ✅ Client-side anti-cheat (paste prevention, focus tracking)
- ✅ Health check endpoint
- ✅ Graceful shutdown handling (Ctrl+C responds in ~6s)

**Phase 2: Recording System**
- ✅ Session manager with UUID-based directories
- ✅ Multi-file recording (keystrokes, terminal, timing, websocket, events)
- ✅ scriptreplay-compatible format
- ✅ SHA-256 integrity hashing
- ✅ Metadata collection and JSON export
- ✅ Buffered I/O with periodic flushing (10s)
- ✅ Session replay script with speed control
- ✅ Comprehensive analysis script

### In Progress
- 🚧 Server-side anti-cheat detection
- 🚧 Reconnection support
- 🚧 Comprehensive metrics (Prometheus)
- 🚧 Docker containerization
- 🚧 Security hardening

## Quick Start

### Prerequisites
- Go 1.21 or later
- Linux or macOS (for PTY support)
- Make (optional, but recommended)

### Using Make (Recommended)

**Build:**
```bash
make build
```

**Run:**
```bash
make run
```

**Run in development mode:**
```bash
make run-dev
```

**See all available commands:**
```bash
make help
```

### Without Make

**Build:**
```bash
go build -o echobox ./cmd/server
```

**Run:**
```bash
./echobox
```

The server will start on `http://localhost:8080` by default.

### Configuration

All configuration is done via environment variables:

| Variable | Default | Description |
|----------|---------|-------------|
| `PORT` | 8080 | HTTP server port |
| `SESSION_TIMEOUT` | 7200 | Max session duration (seconds) |
| `CANDIDATE_NAME` | anonymous | Candidate identifier |
| `OUTPUT_DIR` | /output | Session recording directory |
| `SHELL` | /bin/bash | Shell to spawn in PTY |
| `RECONNECT_WINDOW` | 300 | Reconnection time window (seconds) |
| `INPUT_RATE_LIMIT` | 30 | Max chars/second (anti-paste) |
| `FLUSH_INTERVAL` | 10 | Log flush interval (seconds) |
| `NETWORK_ISOLATED` | true | Enforce network isolation |
| `ENABLE_METRICS` | true | Enable /metrics endpoint |
| `LOG_LEVEL` | info | Log level (debug, info, warn, error) |

Example:
```bash
CANDIDATE_NAME="john_doe" SESSION_TIMEOUT=3600 ./echobox
```

## Makefile Targets

The project includes a comprehensive Makefile for common development tasks:

### Development
- `make build` - Build the binary
- `make build-all` - Build for all platforms (Linux, macOS, Windows)
- `make run` - Build and run the server
- `make run-dev` - Run with development settings
- `make watch` - Watch for changes and rebuild (requires [entr](https://github.com/eradman/entr))

### Testing
- `make test` - Run tests
- `make test-coverage` - Run tests with coverage report
- `make bench` - Run benchmarks

### Code Quality
- `make fmt` - Format code
- `make vet` - Run go vet
- `make lint` - Run golangci-lint (requires [golangci-lint](https://golangci-lint.run/))
- `make check` - Run all checks (fmt, vet)
- `make pre-commit` - Run pre-commit checks (fmt, vet, test)

### Dependencies
- `make deps` - Download dependencies
- `make deps-update` - Update dependencies
- `make tidy` - Tidy go.mod

### Docker
- `make docker-build` - Build Docker image
- `make docker-run` - Run Docker container
- `make docker-stop` - Stop Docker container

### Utilities
- `make clean` - Clean build artifacts
- `make clean-all` - Clean everything including sessions
- `make health` - Check server health (requires [jq](https://stedolan.github.io/jq/))
- `make install` - Install binary to $GOPATH/bin
- `make info` - Show project information
- `make dev-setup` - Setup development environment
- `make help` - Show all available targets

## Architecture

```
┌─────────────────────────────────────────────┐
│  Browser (xterm.js)                         │
│  - Terminal rendering                       │
│  - WebSocket client                         │
│  - Anti-cheat (client-side)                 │
└─────────────────┬───────────────────────────┘
                  │ WebSocket (bidirectional)
┌─────────────────▼───────────────────────────┐
│  Go HTTP Server                             │
│  ├─ WebSocket Handler                       │
│  ├─ PTY Bridge                              │
│  ├─ Session Manager                         │
│  └─ Recording Layer                         │
└─────────────────┬───────────────────────────┘
                  │ PTY I/O
┌─────────────────▼───────────────────────────┐
│  PTY (Pseudo-Terminal)                      │
│  └─ /bin/bash or /bin/zsh                   │
└─────────────────────────────────────────────┘
```

## API Endpoints

### `GET /`
Serves the web terminal interface.

### `GET /ws`
WebSocket endpoint for terminal communication.

**Messages:**
- Text data: Sent to/from PTY as input/output
- JSON messages:
  - `{"type":"resize","data":{"cols":80,"rows":24}}` - Resize terminal
  - `{"type":"anticheat","data":{...}}` - Anti-cheat events
  - `{"type":"finish","data":{...}}` - End session

### `GET /health`
Health check endpoint.

**Response:**
```json
{
  "status": "ok",
  "candidate": "john_doe"
}
```

## Development

### Project Structure
```
echobox/
├── cmd/
│   └── server/
│       └── main.go              # Application entry point
├── internal/
│   ├── config/
│   │   └── config.go            # Configuration management
│   ├── session/
│   │   └── manager.go           # Session lifecycle and metadata
│   ├── terminal/
│   │   ├── pty.go               # PTY allocation and management
│   │   └── recorder.go          # Multi-file session recording
│   ├── web/
│   │   ├── server.go            # HTTP server
│   │   └── websocket.go         # WebSocket handler with recording
│   ├── anticheat/
│   │   ├── detector.go          # [TODO] Server-side detection
│   │   └── analyzer.go          # [TODO] Pattern analysis
│   └── security/
│       └── ratelimit.go         # [TODO] Rate limiting
├── web/
│   ├── index.html               # Main UI
│   ├── terminal.js              # xterm.js integration + anti-cheat
│   ├── style.css                # Styling
│   └── vendor/                  # Third-party libraries (xterm.js)
├── tasks/                       # [TODO] Interview tasks
├── scripts/
│   ├── replay.sh                # Replay recorded sessions
│   └── analyze.sh               # Analyze session data
├── sessions/                    # Recorded sessions (created at runtime)
├── Makefile                     # Build automation
├── TIMEOUTS.md                  # Timeout configuration docs
├── RECORDING.md                 # Recording system docs
├── go.mod
├── go.sum
└── README.md
```

### Building from Source
```bash
# Clone the repository
git clone https://github.com/akonyukov/echobox.git
cd echobox

# Setup development environment (optional)
make dev-setup

# Install dependencies
make deps

# Build
make build

# Run
make run
```

Or without Make:
```bash
go mod download
go build -o echobox ./cmd/server
./echobox
```

### Testing
```bash
# Run tests
make test

# Run tests with coverage
make test-coverage

# Start the server
make run

# In another terminal, check health
make health

# Or manually:
curl http://localhost:8080/health

# Open browser to http://localhost:8080 to test the terminal
```

## Anti-Cheat Measures

### Client-Side (Implemented)
- ✅ Paste events blocked (Ctrl+V, right-click)
- ✅ Context menu disabled
- ✅ Rapid input detection (>30 chars in <100ms)
- ✅ Window focus tracking
- ✅ Tab visibility monitoring
- ✅ All events logged and sent to server

### Server-Side (Implemented & Planned)
- ✅ Anti-cheat event logging (paste attempts, focus loss)
- ✅ WebSocket message logging
- ✅ Session integrity (SHA-256 hashing)
- 🚧 Input rate limiting (enforce max chars/sec)
- 🚧 Typing pattern analysis (WPM, anomaly detection)
- 🚧 Command similarity detection

## Security Features

### Current
- Security headers (X-Frame-Options, CSP, etc.)
- HTTPS-ready (WebSocket upgrade supports WSS)
- Graceful shutdown (no data loss)

### Planned (Phase 5)
- Docker container isolation
- Network isolation (no external access)
- Read-only filesystem
- Resource limits (CPU, memory)
- Non-root user execution
- Capability dropping
- AppArmor/SELinux profiles

## Session Recording

All sessions are automatically recorded to `sessions/` directory (or `OUTPUT_DIR` if configured).

### Recorded Files

Each session creates:
- `keystrokes.log` - Raw input with millisecond timestamps
- `terminal.log` - Complete terminal output (scriptreplay format)
- `timing.log` - Timing data for scriptreplay
- `websocket.log` - All WebSocket messages
- `events.log` - Anti-cheat events (paste, focus loss, etc.)
- `commands.log` - Extracted commands (basic)
- `metadata.json` - Session info + SHA-256 file hashes

### Replay Session
```bash
./scripts/replay.sh sessions/candidate_2026-01-26_14-30-00_a3f7b9c1/
```

Supports real-time, 2x, and 5x playback speeds.

### Analyze Session
```bash
./scripts/analyze.sh sessions/candidate_2026-01-26_14-30-00_a3f7b9c1/
```

Shows:
- Session statistics
- File integrity verification
- Anti-cheat event summary
- Typing speed (WPM estimate)

See [RECORDING.md](RECORDING.md) for complete documentation.

## Roadmap

### Phase 3: Anti-Cheat Enhancement (Next)
- [ ] Server-side paste detection
- [ ] Typing speed analysis
- [ ] Pattern anomaly detection
- [ ] Comprehensive event logging

### Phase 4: Reconnection & Resilience
- [ ] Session state persistence
- [ ] WebSocket reconnection
- [ ] Terminal buffer restoration
- [ ] Connection health monitoring

### Phase 5: Security Hardening
- [ ] Docker multi-stage build
- [ ] Network isolation
- [ ] Resource limits
- [ ] Security testing

### Phase 6: Session Management & UX
- [ ] Session timeout with warnings
- [ ] Finish button with confirmation
- [ ] Instructions panel
- [ ] Metrics endpoint (Prometheus format)

### Phase 7: Tasks & Documentation
- [ ] Sample SRE tasks
- [ ] Replay scripts
- [ ] Analysis scripts
- [ ] Complete documentation

## License

MIT License - see LICENSE file for details

## Contributing

This is currently a personal project for conducting SRE technical interviews. Contributions, issues, and feature requests are welcome!

## Acknowledgments

- [xterm.js](https://xtermjs.org/) - Terminal emulator for the browser
- [gorilla/websocket](https://github.com/gorilla/websocket) - WebSocket implementation
- [creack/pty](https://github.com/creack/pty) - PTY interface for Go
