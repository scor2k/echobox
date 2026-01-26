# Troubleshooting Guide

## Common Issues and Solutions

### 1. JavaScript Files Return 404 in Docker

**Symptoms:**
```
Failed to load resource: 404 (Not Found) xterm.min.js
Failed to load resource: 404 (Not Found) xterm-addon-fit.min.js
```

**Causes:**
- Web vendor files not copied to Docker image
- Working directory mismatch
- File permissions issue

**Solutions:**

**A. Verify files are in the image:**
```bash
# Rebuild with verification
docker compose build echobox-prod

# Check build output for:
# "Web assets copied successfully"

# Or inspect running container
docker exec -it <container_name> ls -la /app/web/vendor/

# Should show:
# xterm.min.js
# xterm.min.css
# xterm-addon-fit.min.js
```

**B. Check working directory:**
```bash
docker exec -it <container_name> pwd
# Should be: /app

docker exec -it <container_name> ls -la web/vendor/
# Should list the xterm files
```

**C. Check file permissions:**
```bash
docker exec -it <container_name> ls -l /app/web/vendor/
# All files should be readable (r-- for candidate user)
```

**D. Rebuild image completely:**
```bash
make docker-clean
make docker-build
```

---

### 2. Permission Denied on Session Directory

**Symptoms:**
```
Failed to create session manager: permission denied
mkdir /output/candidate_...: permission denied
```

**Cause:**
Container runs as UID 1000 (candidate), but mounted volume is owned by different user.

**Solution:**
```bash
# Option 1: Make directory world-writable
chmod 777 sessions/

# Option 2: Change ownership to UID 1000
sudo chown -R 1000:1000 sessions/

# Option 3: Use Makefile (does it automatically)
make docker-compose-prod
```

---

### 3. Port Already in Use

**Symptoms:**
```
bind: address already in use
Error starting userland proxy
```

**Solutions:**

**A. Find what's using the port:**
```bash
lsof -i:8080
# Or
netstat -tulpn | grep 8080
```

**B. Kill the process:**
```bash
lsof -ti:8080 | xargs kill -9
# Or
killall echobox
```

**C. Use a different port:**
```bash
DOCKER_PORT=8081 make docker-run-prod
# Or
DOCKER_PORT=8081 docker compose up echobox-prod
```

**D. Stop all echobox containers:**
```bash
make docker-stop
```

---

### 4. Content-Type Script MIME Error

**Symptoms:**
```
Refused to execute .../xterm.min.js as script because
"X-Content-Type-Options: nosniff" was given and its
Content-Type is not a script MIME type
```

**Cause:**
Server not setting correct Content-Type header for .js files.

**Solution:**
This should be fixed in version 1.0+. If still seeing this:

```bash
# Rebuild from latest code
git pull
make build

# For Docker
make docker-clean
make docker-build
```

**Verify fix:**
```bash
# Start server
make run &
sleep 2

# Check headers
curl -I http://localhost:8080/vendor/xterm.min.js | grep Content-Type
# Should show: application/javascript; charset=utf-8

killall echobox
```

---

### 5. Terminal Not Accepting Input

**Symptoms:**
- Terminal displays but typing does nothing
- No response to keystrokes

**Causes:**
- WebSocket not connected
- Mutex deadlock (fixed in v1.0)
- Browser console errors

**Solutions:**

**A. Check browser console (F12):**
```javascript
// Should see:
WebSocket connected

// Should NOT see:
WebSocket closed
Connection refused
```

**B. Check server logs:**
```bash
# Should see:
WebSocket connected: [ip]

# Should NOT see:
PTY write error
WebSocket write error
```

**C. Restart server:**
```bash
# Stop
make docker-stop
killall echobox

# Start fresh
make run
```

**D. Clear browser cache:**
```
Cmd+Shift+R (Mac) or Ctrl+Shift+R (Windows/Linux)
```

---

### 6. Session Not Recording

**Symptoms:**
- Empty log files in sessions/
- No keystrokes.log or terminal.log content

**Causes:**
- Recorder not initialized
- Flush not called on shutdown
- Permission issues

**Solutions:**

**A. Check if session directory exists:**
```bash
ls -la sessions/
find sessions -name "*_*" -type d
```

**B. Check metadata:**
```bash
cat sessions/*/metadata.json | jq '.status'
# Should be: "completed" or "interrupted"
```

**C. Check server logs for recorder:**
```bash
# Should see:
Recorder: Started session recording in sessions/...
Recorder: Closing and flushing all logs...
Recorder: All logs closed
```

**D. Ensure graceful shutdown:**
```bash
# Don't kill -9 (SIGKILL)
kill -9 <pid>  # ❌ BAD - loses data

# Use SIGTERM or SIGINT
kill -INT <pid>  # ✅ GOOD - graceful
kill -TERM <pid> # ✅ GOOD - graceful
Ctrl+C           # ✅ GOOD - graceful
```

---

### 7. Shell Exit Causes "Reconnecting..." Loop

**Symptoms:**
- User types `exit`
- Terminal shows "Reconnecting..."
- Never stops trying to reconnect

**Cause:**
Old version doesn't detect shell exit properly.

**Solution:**
Update to latest version (should be fixed):

```bash
git pull
make build
# Or
make docker-clean
make docker-build
```

**Expected behavior:**
- User types `exit`
- Terminal shows "Session Ended"
- Server shuts down automatically
- No reconnection attempts

---

### 8. Docker Build Fails

**Symptoms:**
```
go.mod requires go >= 1.24.2 (running go 1.21.13)
```

**Solution:**
Dockerfile should use `golang:1.24-alpine`. Update and rebuild:

```bash
# Pull latest
git pull

# Clean and rebuild
make docker-clean
make docker-build
```

---

### 9. Container Exits Immediately

**Symptoms:**
```
docker ps  # Shows no containers running
docker ps -a  # Shows exited with code 1
```

**Causes:**
- Configuration error
- Missing environment variables
- Permission issues

**Solutions:**

**A. Check container logs:**
```bash
docker logs <container_name>
# Or
make docker-logs
```

**B. Check for errors:**
```bash
docker logs <container_name> 2>&1 | grep -i error
docker logs <container_name> 2>&1 | grep -i fail
```

**C. Run interactively for debugging:**
```bash
# Override CMD to get shell
docker run -it --rm \
  -v $(pwd)/sessions:/output \
  echobox:latest \
  /bin/bash

# Then manually run
/app/echobox
```

---

### 10. Replay Script Fails

**Symptoms:**
```
scriptreplay: command not found
scriptreplay: file not found
```

**Solutions:**

**A. Install scriptreplay:**
```bash
# macOS
brew install util-linux

# Ubuntu/Debian
sudo apt-get install bsdutils

# Alpine (in Docker)
apk add util-linux
```

**B. Check files exist:**
```bash
ls -la sessions/*/terminal.log
ls -la sessions/*/timing.log

# Both files must exist and be non-empty
```

**C. Use analyze instead:**
```bash
# If scriptreplay not available
./scripts/analyze.sh sessions/*
cat sessions/*/keystrokes.log
```

---

### 11. High Memory Usage

**Symptoms:**
- Container killed by OOM
- System slow after running

**Solutions:**

**A. Check container stats:**
```bash
docker stats echobox-prod
```

**B. Increase memory limit:**
```bash
docker run --memory="1g" echobox:latest
# Or in docker-compose.yml:
mem_limit: 1g
```

**C. Check for memory leaks:**
```bash
# Monitor over time
watch -n 1 docker stats --no-stream echobox-prod
```

---

### 12. Can't Stop Container

**Symptoms:**
- `docker stop` hangs
- Container won't respond to signals

**Solutions:**

**A. Wait for graceful shutdown (up to 10s):**
```bash
docker stop --time=15 <container_name>
```

**B. Force kill if needed:**
```bash
docker kill <container_name>
```

**C. Stop all:**
```bash
make docker-stop
# Or
docker ps -q -f name=echobox | xargs docker stop
```

---

### 13. Network Issues in Container

**Symptoms:**
- Can't access external sites
- DNS not working

**Solutions:**

**A. Check network mode:**
```bash
docker inspect <container_name> | jq '.[0].HostConfig.NetworkMode'
```

**B. If isolated (`none`):**
```yaml
# In docker-compose.yml, comment out:
# network_mode: none
```

**C. Test network:**
```bash
docker exec -it <container_name> ping -c 1 8.8.8.8
docker exec -it <container_name> curl -I https://google.com
```

---

### 14. Anti-Cheat False Positives

**Symptoms:**
- Legitimate fast typing flagged
- High WPM causing "SUSPICIOUS" verdict

**Solutions:**

**A. Adjust rate limit:**
```bash
# Increase from 30 to 50 chars/sec
INPUT_RATE_LIMIT=50 ./echobox
```

**B. Review analysis report:**
```bash
cat sessions/*/analysis.json | jq '.flags, .recommendations'
```

**C. Manual review:**
```bash
# Watch the replay to verify
./scripts/replay.sh sessions/*
```

---

## Quick Diagnostic Script

```bash
#!/bin/bash
# diagnose.sh - Quick diagnostics

echo "=== System Check ==="
echo "Go version: $(go version)"
echo "Docker: $(docker --version 2>/dev/null || echo 'Not installed')"
echo ""

echo "=== Port Check ==="
lsof -i:8080 || echo "Port 8080 free"
echo ""

echo "=== Directory Check ==="
ls -ld sessions/ tasks/
echo ""

echo "=== Build Check ==="
[ -f ./echobox ] && echo "✓ Binary exists" || echo "✗ Binary missing"
[ -d web/vendor ] && echo "✓ Vendor directory exists" || echo "✗ Vendor missing"
ls web/vendor/*.js 2>/dev/null && echo "✓ JS files present" || echo "✗ JS files missing"
echo ""

echo "=== Recent Sessions ==="
ls -lt sessions/ | head -5
```

---

## Getting Help

1. **Check logs first:**
   ```bash
   # Local
   cat latest.log

   # Docker
   make docker-logs
   ```

2. **Run diagnostics:**
   ```bash
   ./test_all.sh
   ```

3. **Check documentation:**
   - README.md - Overview
   - QUICKSTART.md - Fast setup
   - DOCKER.md - Deployment
   - SECURITY.md - Security
   - This file - Troubleshooting

4. **File an issue:**
   https://github.com/scor2k/echobox/issues

Include:
- Error message
- Steps to reproduce
- Server logs
- System info (OS, Docker version, Go version)
