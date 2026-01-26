# SRE Interview Terminal - Technical Specification

## Project Overview

Build a self-contained Docker-based web terminal solution for async SRE technical interviews. The solution must provide a web-based terminal interface that connects to a sandboxed Linux environment where candidates complete tasks. All keystrokes and command history must be recorded for later review.

## Core Requirements

### 1. Single Docker Container
- One container that includes both the web server and the terminal environment
- Container runs on a configurable port (default: 8080)
- Container self-terminates after session ends (no auto-restart)
- Clean shutdown with all logs flushed before exit

### 2. Go Web Application
- Single Go binary that serves everything
- Embedded static files (HTML/JS/CSS) - no external dependencies
- WebSocket-based terminal communication
- Uses `github.com/creack/pty` for PTY allocation
- Uses `github.com/gorilla/websocket` for WebSocket handling
- Serves xterm.js terminal emulator (embed minified version)

### 3. Web Terminal Interface
- Clean, minimal UI with xterm.js
- **Copy-paste prevention:**
  - Disable paste events (Ctrl+V, right-click paste, browser paste)
  - Disable copy from terminal (optional, for reviewing commands)
  - Disable context menu
  - Detect and log paste attempts
- Session timeout configuration (e.g., 2 hours max)
- "Finish" button to end session and trigger container shutdown
- Display remaining time

### 4. Keystroke and Session Recording
- Record ALL keystrokes with precise timestamps (millisecond precision)
- Record terminal output (what candidate sees)
- Use `script` command format for compatibility with `scriptreplay`
- Save to mounted volume: `/output/session_<timestamp>/`
  - `keystrokes.log` - raw keystrokes with timestamps
  - `terminal.log` - full terminal session (script format)
  - `timing.log` - timing file for scriptreplay
  - `commands.log` - extracted command history
  - `metadata.json` - session info (start time, end time, duration, etc.)

### 5. Interview Tasks Environment
- Pre-configured Linux environment (Alpine or Debian-slim based)
- Common SRE tools installed: vim, nano, curl, wget, netstat, ss, ps, top, htop, systemctl (if applicable), journalctl, grep, awk, sed, jq, dig, nslookup, tcpdump, strace
- Tasks placed in `/tasks/` directory with README.md
- Candidate home directory: `/home/candidate/`
- Solution submission directory: `/home/candidate/solutions/`

### 6. Container Lifecycle
```
Start container
    ↓
Web server starts on specified port
    ↓
Candidate opens browser, session begins
    ↓
Recording starts automatically
    ↓
Candidate works on tasks
    ↓
Session ends (button click OR timeout OR terminal exit)
    ↓
Flush all logs to mounted volume
    ↓
Container exits with code 0
```

## Technical Architecture

### Directory Structure
```
sre-interview-terminal/
├── cmd/
│   └── server/
│       └── main.go           # Entry point
├── internal/
│   ├── terminal/
│   │   ├── pty.go            # PTY handling
│   │   ├── recorder.go       # Keystroke recording
│   │   └── session.go        # Session management
│   ├── web/
│   │   ├── handler.go        # HTTP handlers
│   │   ├── websocket.go      # WebSocket handling
│   │   └── static.go         # Embedded static files
│   └── config/
│       └── config.go         # Configuration
├── web/
│   ├── index.html            # Main page
│   ├── terminal.js           # Terminal logic + anti-paste
│   └── style.css             # Styling
├── tasks/                    # Sample interview tasks
│   ├── README.md
│   ├── task1/
│   └── task2/
├── Dockerfile
├── docker-compose.yml
├── go.mod
├── go.sum
└── README.md
```

### Go Dependencies
```go
require (
    github.com/creack/pty v1.1.21
    github.com/gorilla/websocket v1.5.1
)
```

### Key Implementation Details

#### PTY + Recording (internal/terminal/recorder.go)
```
WebSocket ←→ Go Server ←→ PTY ←→ bash
                ↓
         Record to files:
         - Input (keystrokes + timestamps)
         - Output (terminal display)
```

#### Anti-Paste Detection (terminal.js)
- Intercept all paste events
- Detect rapid input (>50 chars in <100ms = likely paste)
- Log all paste attempts with timestamp and attempted content
- Show warning to candidate: "Paste is disabled for this assessment"

#### Session Timeout
- Configurable via environment variable: `SESSION_TIMEOUT=7200` (seconds)
- Warning at 10 minutes remaining
- Auto-terminate and save logs when timeout reached

#### Graceful Shutdown
- SIGTERM handler in Go
- Flush all buffers
- Close PTY cleanly
- Write final metadata
- Exit with code 0

## Configuration (Environment Variables)

| Variable | Default | Description |
|----------|---------|-------------|
| `PORT` | 8080 | Web server port |
| `SESSION_TIMEOUT` | 7200 | Max session duration (seconds) |
| `CANDIDATE_NAME` | anonymous | Candidate identifier for logs |
| `OUTPUT_DIR` | /output | Where to save session logs |
| `SHELL` | /bin/bash | Shell to spawn |
| `MOTD` | (see below) | Message shown at session start |

## Docker Usage

### Build
```bash
docker build -t sre-interview-terminal .
```

### Run
```bash
docker run -d \
  --name interview-candidate-john \
  -p 8080:8080 \
  -v $(pwd)/sessions:/output \
  -v $(pwd)/my-tasks:/tasks:ro \
  -e CANDIDATE_NAME="john_doe" \
  -e SESSION_TIMEOUT=7200 \
  --restart=no \
  sre-interview-terminal
```

### Output Structure (after session)
```
sessions/
└── john_doe_2024-01-15_14-30-00/
    ├── metadata.json
    ├── keystrokes.log
    ├── terminal.log
    ├── timing.log
    ├── commands.log
    └── paste_attempts.log
```

## Deliverables

1. Complete Go source code with all packages
2. Dockerfile (multi-stage build for small image)
3. docker-compose.yml for easy testing
4. Sample interview tasks (2-3 SRE scenarios)
5. README.md with usage instructions
6. Script to replay sessions: `replay.sh`

## Anti-Cheat Measures Summary

1. **Paste disabled** - JS event interception
2. **Paste detection** - Timing analysis for rapid input
3. **Full keystroke log** - Can analyze typing patterns
4. **Timing data** - Can replay session in real-time to review
5. **Isolated environment** - Cannot access external resources (optional: block network)

## Optional Enhancements (if time permits)

- Basic auth for web interface
- Multiple concurrent sessions (different ports)
- Network isolation (no external access)
- Resource limits (CPU/memory)
- Automatic task validation scripts

