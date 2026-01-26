# Docker Deployment Guide

## Overview

Echobox is designed to run as a single-use Docker container for each interview session. Each container represents one candidate's isolated environment.

## Quick Start

### Build Image
```bash
make docker-build
```

### Run Container
```bash
# Development mode
make docker-run

# Or with docker-compose
make docker-compose-up
```

## Image Details

### Multi-Stage Build
- **Builder stage**: golang:1.21-alpine (builds binary)
- **Runtime stage**: alpine:3.19 (minimal runtime)
- **Final size**: ~40-50MB

### Installed Tools
- bash, vim, nano
- curl, wget
- jq (JSON processing)
- htop, ps (process monitoring)
- net-tools (netstat, ifconfig)
- util-linux (scriptreplay)

### Security Features
- ✅ Non-root user (UID 1000: `candidate`)
- ✅ Minimal base image (Alpine Linux)
- ✅ No unnecessary packages
- ✅ Health check included
- ✅ Security options: no-new-privileges
- ✅ Capability dropping (CAP_DROP=ALL + selective add)
- ✅ Resource limits (CPU, memory)
- ⚠️ Writable filesystem (required for tasks)

## Production Deployment

### Single Interview Session
```bash
docker run -d \
  --name interview-jane-doe-$(date +%s) \
  -p 8080:8080 \
  -v $(pwd)/sessions:/output \
  -v $(pwd)/tasks:/tasks:ro \
  -e CANDIDATE_NAME="jane_doe" \
  -e SESSION_TIMEOUT=7200 \
  --memory="512m" \
  --memory-reservation="256m" \
  --cpus="0.5" \
  --security-opt=no-new-privileges:true \
  --cap-drop=ALL \
  --cap-add=CHOWN \
  --cap-add=SETUID \
  --cap-add=SETGID \
  --restart=no \
  echobox:latest
```

### Multiple Concurrent Sessions
```bash
# Session 1 on port 8081
docker run -d \
  --name interview-candidate1 \
  -p 8081:8080 \
  -v $(pwd)/sessions:/output \
  -e CANDIDATE_NAME="candidate1" \
  --memory="512m" \
  --cpus="0.5" \
  --restart=no \
  echobox:latest

# Session 2 on port 8082
docker run -d \
  --name interview-candidate2 \
  -p 8082:8080 \
  -v $(pwd)/sessions:/output \
  -e CANDIDATE_NAME="candidate2" \
  --memory="512m" \
  --cpus="0.5" \
  --restart=no \
  echobox:latest
```

## Environment Variables

All variables from [README.md](README.md#configuration) are supported:

```bash
docker run -d \
  -e CANDIDATE_NAME="john_doe" \
  -e SESSION_TIMEOUT=3600 \
  -e OUTPUT_DIR=/output \
  -e INPUT_RATE_LIMIT=30 \
  -e RECONNECT_WINDOW=300 \
  -e LOG_LEVEL=info \
  echobox:latest
```

## Volume Mounts

### Required Volumes

**Output Directory** (recording storage):
```bash
-v $(pwd)/sessions:/output
```
- Must be writable
- Sessions saved here: `/output/<candidate>_<timestamp>_<id>/`

### Optional Volumes

**Tasks Directory** (interview tasks):
```bash
-v $(pwd)/tasks:/tasks:ro
```
- Read-only mount
- Tasks visible at `/tasks/` in container

**Custom Shell Config**:
```bash
-v $(pwd)/bashrc:/home/candidate/.bashrc:ro
```

## Resource Limits

### Memory
```bash
--memory="512m"              # Hard limit
--memory-reservation="256m"   # Soft limit
--memory-swap="512m"          # Swap limit (same as memory = no swap)
```

### CPU
```bash
--cpus="0.5"                 # 50% of one CPU core
--cpu-shares=512             # Relative weight (default: 1024)
```

### Disk I/O (if needed)
```bash
--device-write-bps /dev/sda:10mb   # Limit write speed
--device-read-bps /dev/sda:50mb    # Limit read speed
```

## Network Configuration

### Default (Bridge Network)
```bash
# Allows external network access (default)
docker run echobox:latest
```

Use this when:
- Tasks require external package installation
- Tasks need to curl external APIs
- Debugging requires network access

### Network Isolation
```bash
# Complete network isolation
docker run --network none echobox:latest
```

Use this when:
- Maximum security required
- All tasks are self-contained
- Want to prevent external resource access

### Custom Network
```bash
# Create isolated network with DNS
docker network create --internal interview-network
docker run --network interview-network echobox:latest
```

## Security Hardening

### Recommended Production Settings
```bash
docker run -d \
  --security-opt=no-new-privileges:true \
  --cap-drop=ALL \
  --cap-add=CHOWN \
  --cap-add=SETUID \
  --cap-add=SETGID \
  --pids-limit=100 \
  --ulimit nofile=1024:2048 \
  --ulimit nproc=512:1024 \
  echobox:latest
```

### Why Not Read-Only Filesystem?
Interview tasks often require:
- Creating shell scripts
- Modifying configuration files
- Writing solution files
- Creating temporary files

Therefore, the filesystem must be writable. Security is maintained through:
- Non-root user execution
- Resource limits
- Capability dropping
- Process limits

## Container Lifecycle

### Single-Use Pattern
```bash
# Start container
docker run -d --name interview-candidate1 echobox:latest

# Candidate completes interview
# - Clicks "Finish" button OR
# - Types "exit" in terminal

# Container automatically exits (code 0)
# Session saved to mounted volume

# Remove container
docker rm interview-candidate1
```

### Auto-Removal
```bash
# Container removes itself after exit
docker run --rm echobox:latest
```

## Monitoring

### Container Logs
```bash
# Follow logs
make docker-logs

# Or directly
docker logs -f echobox-dev

# Show last 100 lines
docker logs --tail 100 echobox-dev
```

### Health Check
```bash
# Docker built-in health check
docker ps --format "table {{.Names}}\t{{.Status}}"

# Manual health check
curl http://localhost:8080/health
```

### Resource Usage
```bash
# Real-time stats
docker stats echobox-dev

# One-time stats
docker stats --no-stream echobox-dev
```

## Docker Compose

### Development
```bash
# Start in foreground
docker-compose up echobox-dev

# Start in background
docker-compose up -d echobox-dev

# Stop
docker-compose down
```

### Production
```bash
# Start production config
CANDIDATE_NAME=john_doe SESSION_TIMEOUT=7200 docker-compose up -d echobox-prod

# Check logs
docker-compose logs -f echobox-prod

# Stop
docker-compose down
```

## Makefile Commands

### Build
- `make docker-build` - Build latest image
- `make docker-build-prod` - Build production image with version tags

### Run
- `make docker-run` - Build and run development container
- `make docker-run-prod` - Run production container
- `make docker-compose-up` - Start with docker-compose (dev)
- `make docker-compose-prod` - Start with docker-compose (prod)

### Manage
- `make docker-stop` - Stop all echobox containers
- `make docker-logs` - Show container logs
- `make docker-exec` - Shell into running container
- `make docker-clean` - Remove all echobox images and containers
- `make docker-inspect` - Inspect container configuration

## Troubleshooting

### Container Won't Start

**Check logs:**
```bash
docker logs echobox-dev
```

**Common issues:**
- Port 8080 already in use: Change `-p 8081:8080`
- Volume mount permission denied: Check directory ownership
- Out of memory: Increase `--memory` limit

### Session Not Recorded

**Check volume mount:**
```bash
docker inspect echobox-dev | jq '.[0].Mounts'
```

**Verify output directory:**
```bash
docker exec echobox-dev ls -la /output
```

**Check permissions:**
```bash
ls -la sessions/
```

### Cannot Type in Terminal

**Check WebSocket connection:**
```bash
# Browser console
# Should show: WebSocket connected

docker logs echobox-dev | grep WebSocket
```

### Container Doesn't Exit

**Shell not closed:**
- User must type `exit` or click "Finish" button
- Or send SIGTERM: `docker stop echobox-dev`

**Check if stuck:**
```bash
docker logs --tail 50 echobox-dev
```

## Best Practices

### 1. Unique Container Names
```bash
# Include candidate name and timestamp
--name interview-${CANDIDATE_NAME}-$(date +%s)
```

### 2. Persistent Sessions
```bash
# Always mount sessions directory
-v $(pwd)/sessions:/output
```

### 3. Resource Limits
```bash
# Always set limits to prevent resource exhaustion
--memory="512m" --cpus="0.5"
```

### 4. Clean Up
```bash
# Remove container after interview
docker rm interview-candidate1

# Or use --rm for auto-removal
docker run --rm echobox:latest
```

### 5. Security
```bash
# Always use security options in production
--security-opt=no-new-privileges:true
--cap-drop=ALL
```

## Advanced Usage

### Custom Tasks
```bash
# Mount custom task directory
-v $(pwd)/my-custom-tasks:/tasks:ro
```

### Custom Shell Configuration
```bash
# Provide custom .bashrc
-v $(pwd)/interview-bashrc:/home/candidate/.bashrc:ro
```

### Logging to External System
```bash
# Use Docker logging driver
docker run \
  --log-driver=syslog \
  --log-opt syslog-address=tcp://logserver:514 \
  echobox:latest
```

### Multiple Sessions on One Host
```bash
# Use different ports
for i in {1..5}; do
  docker run -d \
    -p $((8080 + i)):8080 \
    -e CANDIDATE_NAME="candidate$i" \
    --name echobox-candidate$i \
    echobox:latest
done
```

## Dual-User Architecture (Audit Log Protection)

### Security Model

Echobox uses **two separate users** for security:

1. **Application User**: `echobox` (UID 1001)
   - Runs the Go application
   - Writes audit logs to `/output`
   - No shell access (`/bin/false`)
   - Owns all log files

2. **Interactive User**: `candidate` (UID 1000)
   - Used for shell access during interviews
   - Works in `/home/candidate` for tasks
   - Can READ logs but cannot MODIFY them
   - Cannot tamper with audit trail

### Why This Matters

**Without separation:**
```bash
# Bad: Candidate can tamper with logs
docker exec -it container bash
rm /output/session/keystrokes.log  # ✅ Works - BAD!
```

**With dual-user architecture:**
```bash
# Good: Logs are protected
docker exec -it --user candidate container bash
rm /output/session/keystrokes.log  # ❌ Permission denied - GOOD!
```

### Shell Access for Debugging

**Correct way to exec into container:**
```bash
# As candidate user (for tasks/debugging)
docker exec -it --user candidate echobox-prod-123 /bin/bash

# Or use Makefile (does this automatically)
make docker-exec
```

**What candidate CAN do:**
```bash
# Work on tasks
cd ~/solutions
vi task1.sh

# View logs (read-only)
cat /output/session_*/keystrokes.log

# Do interview tasks
cd /tasks
./check_system.sh
```

**What candidate CANNOT do:**
```bash
# Modify logs (permission denied)
echo "fake" >> /output/session_*/keystrokes.log

# Delete logs (permission denied)
rm /output/session_*/terminal.log

# Change ownership (permission denied)
chown candidate /output/session_*/
```

### File Ownership in Container

```bash
# Check who owns what
docker exec -it container-name ls -la /

# Application and logs owned by echobox (UID 1001)
/app/echobox              -> echobox:echobox
/output/session_*/        -> echobox:echobox

# Home directory owned by candidate (UID 1000)
/home/candidate/          -> candidate:candidate
```

## Security Considerations

### What's Protected
✅ **Tamper-proof audit logs** (echobox UID 999, candidate cannot modify)
✅ Non-root execution (app: UID 999, shell: UID 1000)
✅ Resource limits prevent DoS
✅ Capability dropping limits syscalls
✅ Single-use containers (no state persistence)
✅ Isolated sessions per candidate

### What's NOT Protected
⚠️ Writable filesystem (by design - needed for tasks)
⚠️ Network access (optional - may be needed)
⚠️ Candidate can READ logs (acceptable for debugging tasks)
⚠️ No AppArmor/SELinux profile (optional enhancement)

### Threat Model
- Candidates cannot escalate to root
- **Candidates cannot tamper with audit logs** (dual-user protection)
- Resource exhaustion prevented by limits
- Network isolation optional (task-dependent)
- Session recordings tamper-evident (SHA-256 + OS permissions)
- Anti-cheat detects paste/automation

## Production Checklist

- [ ] Build production image: `make docker-build-prod`
- [ ] Test with sample candidate: `make docker-run`
- [ ] Verify sessions recorded correctly
- [ ] Test reconnection after browser refresh
- [ ] Test shell exit (`exit` command)
- [ ] Verify resource limits enforced
- [ ] Check container stops after session
- [ ] Review security settings
- [ ] Test analysis scripts work: `./scripts/analyze.sh sessions/*`
- [ ] Test replay works: `./scripts/replay.sh sessions/*`

## Example Workflow

```bash
# 1. Build image
make docker-build

# 2. Prepare tasks
mkdir -p tasks
echo "Task description" > tasks/README.md

# 3. Start interview
docker run -d \
  --name interview-jane-$(date +%s) \
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

# 4. Send URL to candidate
echo "Interview URL: http://your-server:8080"

# 5. Monitor (optional)
make docker-logs

# 6. After interview, review session
./scripts/analyze.sh sessions/jane_doe_*
./scripts/replay.sh sessions/jane_doe_*

# 7. Cleanup
docker rm interview-jane-*
```
