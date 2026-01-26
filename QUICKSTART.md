# Quick Start Guide - Fast Mode

This guide gets you up and running with Echobox in under 2 minutes for testing and development.

## 🚀 Fastest Way to Run

### Prerequisites
- Go 1.21+ installed
- 2 minutes of your time

### 3 Commands to Live Server
```bash
# 1. Build
make build

# 2. Run
make run

# 3. Open browser
open http://localhost:8080
```

**Done!** You now have a fully functional interview terminal.

---

## Fast Development Mode

### Quick Iteration
```bash
# Run with debug logging and short timeout
make run-dev
```

This starts the server with:
- **Candidate**: `dev_user`
- **Timeout**: 1 hour (instead of 2)
- **Log Level**: `debug` (verbose)
- **Output**: `./sessions`

### Watch Mode (Auto-Reload)
```bash
# Install entr first: brew install entr
make watch
```

Auto-rebuilds and restarts on any code change!

---

## Quick Testing

### Test Single Feature
```bash
# Just build
make build

# Just test
make test

# Just check code quality
make check
```

### Test Full System
```bash
# Run comprehensive test suite
./test_all.sh
```

Tests everything in ~30 seconds.

---

## Fast Docker Testing

### One-Command Docker
```bash
# Build and run in one command
make docker-run
```

This will:
1. Build the Docker image
2. Create sessions and tasks directories
3. Start the container
4. Open port 8080

### Docker with Custom Config
```bash
# Quick production test
CANDIDATE_NAME=test_candidate make docker-run-prod
```

---

## Rapid Session Review

### Analyze Last Session
```bash
# Analyze most recent session
./scripts/analyze.sh $(ls -td sessions/*/ | head -1)
```

### Replay Last Session
```bash
# Replay most recent session
./scripts/replay.sh $(ls -td sessions/*/ | head -1)
```

### Quick Session Check
```bash
# List all sessions with basic info
for session in sessions/*/; do
  echo "Session: $(basename $session)"
  jq -r '"  Candidate: \(.candidate_name), Duration: \(.duration_seconds)s, Status: \(.status)"' $session/metadata.json
  echo ""
done
```

---

## Fast Configuration

### Environment Variables (Quick Override)
```bash
# Quick 30-minute interview
SESSION_TIMEOUT=1800 ./echobox

# Custom candidate name
CANDIDATE_NAME="jane_doe" ./echobox

# Debug mode
LOG_LEVEL=debug ./echobox

# Different port
PORT=9090 ./echobox

# All at once
PORT=9090 \
CANDIDATE_NAME="test" \
SESSION_TIMEOUT=600 \
LOG_LEVEL=debug \
./echobox
```

### Multiple Quick Sessions
```bash
# Start 3 interviews on different ports
for i in {1..3}; do
  PORT=$((8080 + i)) \
  CANDIDATE_NAME="candidate$i" \
  OUTPUT_DIR=./sessions \
  ./echobox &
done

# URLs:
# http://localhost:8081 - candidate1
# http://localhost:8082 - candidate2
# http://localhost:8083 - candidate3

# Stop all
killall echobox
```

---

## Fast Cleanup

### Clean Everything
```bash
# Nuclear option - clean all artifacts
make clean-all

# Or selectively
make clean              # Just build artifacts
rm -rf sessions/*       # Just sessions
make docker-clean       # Just Docker resources
```

### Quick Session Cleanup
```bash
# Delete old sessions (older than 7 days)
find sessions -name "*_*" -type d -mtime +7 -exec rm -rf {} \;

# Delete test sessions
rm -rf sessions/test_*
rm -rf sessions/dev_*
rm -rf sessions/anonymous_*
```

---

## Development Shortcuts

### Skip Building
```bash
# If binary already exists, just run
OUTPUT_DIR=./sessions ./echobox
```

### Background Mode
```bash
# Run in background
make run &
SERVER_PID=$!

# Do your testing...

# Stop when done
kill -INT $SERVER_PID
```

### Quick Health Check
```bash
# Start server in background
make run &
sleep 2

# Check if running
make health

# Or manually
curl http://localhost:8080/health | jq

# Stop
killall echobox
```

---

## Fast Debugging

### Check Why It's Not Working

**Server won't start:**
```bash
# Check if port is in use
lsof -i:8080

# Kill existing process
lsof -ti:8080 | xargs kill -9

# Try again
make run
```

**Can't type in terminal:**
```bash
# Check browser console (F12)
# Should see: "WebSocket connected"

# Check server logs
# Should see: "WebSocket connected: [ip]"
```

**Session not recording:**
```bash
# Check if sessions directory exists
ls -la sessions/

# Check if session was created
find sessions -name "*_*" -type d

# Check logs
cat sessions/*/metadata.json | jq '.status'
```

---

## Fast Mode Aliases

Add these to your `.bashrc` or `.zshrc`:

```bash
# Quick start echobox
alias ebox='cd ~/echobox && make run'

# Quick build
alias ebox-build='cd ~/echobox && make build'

# Quick analyze last session
alias ebox-analyze='cd ~/echobox && ./scripts/analyze.sh $(ls -td sessions/*/ | head -1)'

# Quick replay last session
alias ebox-replay='cd ~/echobox && ./scripts/replay.sh $(ls -td sessions/*/ | head -1)'

# Quick cleanup
alias ebox-clean='cd ~/echobox && rm -rf sessions/test_* sessions/dev_* sessions/anonymous_*'

# Quick Docker
alias ebox-docker='cd ~/echobox && make docker-run'
```

Then just type:
```bash
ebox           # Start server
ebox-analyze   # Analyze latest session
ebox-clean     # Clean test sessions
```

---

## Performance Tips

### Faster Builds
```bash
# Use cached dependencies
go build -o echobox ./cmd/server  # Skips 'make clean'

# Build without cleanup
go build -o echobox ./cmd/server
```

### Faster Tests
```bash
# Run short tests only
make test-short

# Skip tests during development
make build  # Just builds, no tests
```

### Faster Session Analysis
```bash
# Skip integrity checks
jq '.verdict, .confidence_score' sessions/*/analysis.json

# Quick event count
wc -l sessions/*/events.log
```

---

## Common Fast Workflows

### Quick Local Test
```bash
make run &
sleep 2
open http://localhost:8080
# Test manually
# Press Ctrl+C when done
```

### Quick Docker Test
```bash
make docker-build
make docker-run
# Opens in terminal, Ctrl+C to stop
```

### Quick Session Review
```bash
# Start
make run
# Do interview
# Press Ctrl+C

# Review
SESSION=$(ls -td sessions/*/ | head -1)
./scripts/analyze.sh $SESSION | grep -E "Verdict|Confidence|WPM"
```

### Quick Multiple Candidates
```bash
# Terminal 1
PORT=8081 CANDIDATE_NAME=alice OUTPUT_DIR=./sessions ./echobox

# Terminal 2
PORT=8082 CANDIDATE_NAME=bob OUTPUT_DIR=./sessions ./echobox

# Terminal 3
PORT=8083 CANDIDATE_NAME=charlie OUTPUT_DIR=./sessions ./echobox

# URLs:
# Alice:   http://localhost:8081
# Bob:     http://localhost:8082
# Charlie: http://localhost:8083
```

---

## Troubleshooting (Fast Fixes)

### Port Already in Use
```bash
lsof -ti:8080 | xargs kill -9 && make run
```

### Session Directory Doesn't Exist
```bash
mkdir -p sessions && make run
```

### Can't Build
```bash
go mod tidy && make build
```

### Docker Won't Start
```bash
# Start Docker daemon
# macOS: open -a Docker
# Linux: sudo systemctl start docker
```

---

## Summary: The Absolute Fastest Path

```bash
# Clone (first time only)
git clone https://github.com/scor2k/echobox.git
cd echobox

# Run (every time)
make run

# Review (after interview)
./scripts/analyze.sh sessions/*
```

**That's it!** Three commands to a fully functional interview system.

---

## What's Next?

After testing in fast mode:

1. **Customize tasks**: Edit `tasks/` directory
2. **Deploy to production**: See [DOCKER.md](DOCKER.md)
3. **Harden security**: Review [SECURITY.md](SECURITY.md)
4. **Scale up**: Run multiple containers

---

**Fast mode is perfect for:**
- ✅ Testing and development
- ✅ Trying out features
- ✅ Learning how it works
- ✅ Quick demonstrations

**For production interviews:**
- Use Docker deployment
- Enable all security features
- See full documentation in README.md

Happy interviewing! 🚀
