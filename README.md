# Echobox - SRE Interview Terminal

A secure, web-based terminal environment for conducting technical SRE interviews with comprehensive session recording and anti-cheat measures.

## 🚀 Quick Start

```bash
# Clone the repository
git clone https://github.com/scor2k/echobox.git
cd echobox

# Build and run (local)
make run

# Or with Docker
make docker-build
make docker-run

# Open browser to http://localhost:8080
# Type commands in the terminal - it works!
# Click "Finish Session" or type 'exit' when done

# Review the session
./scripts/analyze.sh sessions/anonymous_*
./scripts/replay.sh sessions/anonymous_*
```

**That's it!** You now have a fully functional interview terminal with recording, anti-cheat, and reconnection support.

**Want more details?** See [QUICKSTART.md](QUICKSTART.md) for fast mode guide, shortcuts, and development workflows.

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

**Phase 3: Anti-Cheat Detection**
- ✅ Server-side rate limiting (30 chars/sec threshold)
- ✅ Burst detection for paste attempts
- ✅ Real-time typing pattern analysis
- ✅ Anti-cheat event logging (info/warning/critical severity)
- ✅ Post-session analysis with verdict generation
- ✅ WPM calculation and anomaly detection
- ✅ Automated recommendations based on patterns

**Phase 4: Reconnection & Resilience**
- ✅ Session state management with connection tracking
- ✅ Reconnection token system (5-minute window)
- ✅ Terminal buffer persistence (100KB rolling buffer)
- ✅ GET /reconnect endpoint for session resumption
- ✅ Client-side auto-reconnect with exponential backoff
- ✅ Terminal state restoration on reconnect
- ✅ Connection statistics tracking
- ✅ Shell exit detection (exit command ends session cleanly)

**Phase 5: Docker & Security**
- ✅ Multi-stage Dockerfile (<50MB image)
- ✅ Docker Compose (dev + prod configurations)
- ✅ Non-root user execution (candidate:1000)
- ✅ Resource limits (512MB memory, 0.5 CPU)
- ✅ Security hardening (cap_drop, no-new-privileges)
- ✅ Optional network isolation
- ✅ Health checks and monitoring
- ✅ Enhanced Makefile with 15+ Docker targets
- ✅ Complete deployment documentation (DOCKER.md)
- ✅ Sample interview task (debugging exercise)

### Remaining
- 🚧 Comprehensive metrics (Prometheus format) - Phase 6
- 🚧 Additional interview tasks - Phase 7

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
| `OUTPUT_DIR` | ./sessions | Session recording directory |
| `SHELL` | /bin/bash | Shell to spawn in PTY |
| `RECONNECT_WINDOW` | 300 | Reconnection time window (seconds) |
| `INPUT_RATE_LIMIT` | 30 | Max chars/second (anti-paste) |
| `FLUSH_INTERVAL` | 10 | Log flush interval (seconds) |
| `NETWORK_ISOLATED` | true | Enforce network isolation |
| `ENABLE_METRICS` | true | Enable /metrics endpoint |
| `LOG_LEVEL` | info | Log level (debug, info, warn, error) |

**Note**: Use `/output` for Docker containers (mounted volume), `./sessions` for local development.

Example:
```bash
CANDIDATE_NAME="john_doe" SESSION_TIMEOUT=3600 ./echobox

# Or for Docker:
docker run -v $(pwd)/sessions:/output -e OUTPUT_DIR=/output echobox:latest
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
- Network isolation (optional - may need external access for some tasks)
- **Writable filesystem** - Candidates need to create/modify files during tasks
- Resource limits (CPU, memory, disk I/O)
- Non-root user execution
- Capability dropping (CAP_DROP=ALL except necessary)
- AppArmor/SELinux profiles (optional)

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

## Interview Tasks

Three sample SRE tasks included in `tasks/` directory:

1. **Debugging** (`tasks/01-debugging/`): Fix a broken web application
2. **Incident Response** (`tasks/02-incident-response/`): Investigate high CPU usage
3. **Automation** (`tasks/03-automation/`): Write log analysis script

See [tasks/README.md](tasks/README.md) for candidate instructions.

## Documentation

- **[README.md](README.md)** - This file (overview and quick start)
- **[QUICKSTART.md](QUICKSTART.md)** - Fast mode guide (2-minute setup, shortcuts, workflows)
- **[DOCKER.md](DOCKER.md)** - Complete Docker deployment guide (511 lines)
- **[SECURITY.md](SECURITY.md)** - Security model and best practices (401 lines)
- **[RECORDING.md](RECORDING.md)** - Session recording details
- **[TIMEOUTS.md](TIMEOUTS.md)** - Timeout configuration reference

## Testing

Run the comprehensive test suite:
```bash
./test_all.sh
```

Tests:
- Build process
- Server startup
- Recording system
- Session management
- Analysis scripts
- Docker configuration
- Interview tasks

## Production Deployment

### Option 1: Docker (Recommended)
```bash
# Build image
make docker-build

# Run production container
docker run -d \
  --name interview-jane-doe \
  -p 8080:8080 \
  -v $(pwd)/sessions:/output \
  -v $(pwd)/tasks:/tasks:ro \
  -e CANDIDATE_NAME="jane_doe" \
  -e SESSION_TIMEOUT=7200 \
  --memory="512m" \
  --cpus="0.5" \
  --security-opt=no-new-privileges:true \
  --restart=no \
  echobox:latest

# Send URL to candidate
echo "Interview: http://your-server:8080"
```

See [DOCKER.md](DOCKER.md) for complete deployment guide.

### Option 2: Direct Binary
```bash
# Build
make build

# Run with custom config
OUTPUT_DIR=./sessions \
CANDIDATE_NAME="john_doe" \
SESSION_TIMEOUT=3600 \
./echobox
```

## Project Status

**Current Version**: 1.0 (Production Ready)

All core phases completed:
- ✅ Phase 1: Core Terminal
- ✅ Phase 2: Recording System
- ✅ Phase 3: Anti-Cheat Detection
- ✅ Phase 4: Reconnection & Resilience
- ✅ Phase 5: Docker & Security
- ✅ Phase 7: Tasks & Documentation

**Not Implemented:**
- ⏭️ Phase 6: Prometheus Metrics (optional)

## License

MIT License - see LICENSE file for details

## Contributing

This is currently a personal project for conducting SRE technical interviews. Contributions, issues, and feature requests are welcome!

## Acknowledgments

- [xterm.js](https://xtermjs.org/) - Terminal emulator for the browser
- [gorilla/websocket](https://github.com/gorilla/websocket) - WebSocket implementation
- [creack/pty](https://github.com/creack/pty) - PTY interface for Go
