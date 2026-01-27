# Running Multiple Interview Instances

This guide shows how to run 5 simultaneous interview sessions using docker-compose.

## Quick Start

### Start All 5 Instances

```bash
# Set candidate names (optional)
export CANDIDATE_1="alice_smith"
export CANDIDATE_2="bob_jones"
export CANDIDATE_3="charlie_brown"
export CANDIDATE_4="diana_prince"
export CANDIDATE_5="eve_wilson"

# Start all instances
docker compose up -d

# Or with Makefile
make docker-compose-up
```

### URLs for Candidates

```
Instance 1: http://your-server:6780 (alice_smith)
Instance 2: http://your-server:6781 (bob_jones)
Instance 3: http://your-server:6782 (charlie_brown)
Instance 4: http://your-server:6783 (diana_prince)
Instance 5: http://your-server:6784 (eve_wilson)
```

---

## Managing Instances

### Start Single Instance

```bash
# Start instance 1
INSTANCE=1 make docker-compose-one

# Start instance 3
INSTANCE=3 make docker-compose-one

# Or directly
docker compose up -d echobox-prod-1
docker compose up -d echobox-prod-3
```

### Check Status

```bash
# Show status of all instances
make docker-status

# Output:
# ✓ Instance 1 (port 6780): alice_smith - Up 5 minutes
# ✓ Instance 2 (port 6781): bob_jones - Up 5 minutes
# ○ Instance 3 (port 6782): Not running
# ○ Instance 4 (port 6783): Not running
# ○ Instance 5 (port 6784): Not running
```

### View Logs

```bash
# View logs for instance 1
INSTANCE=1 make docker-logs

# View logs for instance 3
INSTANCE=3 make docker-logs

# Or directly
docker logs -f echobox-prod-1
docker logs -f echobox-prod-3
```

### Shell Access (Debug)

```bash
# Shell into instance 1
INSTANCE=1 make docker-exec

# Shell into instance 4
INSTANCE=4 make docker-exec

# Or directly
docker exec -it echobox-prod-1 /bin/bash
```

### Stop Instances

```bash
# Stop all instances
docker compose down

# Stop specific instance
docker compose stop echobox-prod-1

# Or
make docker-stop  # Stops all echobox containers
```

---

## Configuration

### Set Candidate Names

**Option 1: Environment Variables**
```bash
CANDIDATE_1="alice" \
CANDIDATE_2="bob" \
CANDIDATE_3="charlie" \
docker compose up -d
```

**Option 2: .env File**
Create `.env` file in project root:
```bash
CANDIDATE_1=alice_smith
CANDIDATE_2=bob_jones
CANDIDATE_3=charlie_brown
CANDIDATE_4=diana_prince
CANDIDATE_5=eve_wilson
SESSION_TIMEOUT=7200
```

Then just run:
```bash
docker compose up -d
```

**Option 3: Modify docker-compose.yml**
```yaml
environment:
  CANDIDATE_NAME: alice_smith  # Change default
```

---

## Port Assignment

Fixed ports for each instance:
- Instance 1: **6780**
- Instance 2: **6781**
- Instance 3: **6782**
- Instance 4: **6783**
- Instance 5: **6784**

### Firewall Configuration

```bash
# Allow all 5 ports
sudo ufw allow 6780:6784/tcp

# Or individually
sudo ufw allow 6780/tcp
sudo ufw allow 6781/tcp
sudo ufw allow 6782/tcp
sudo ufw allow 6783/tcp
sudo ufw allow 6784/tcp
```

---

## Session Organization

All instances share the same `/sessions` volume:

```
sessions/
├── alice_smith_2026-01-27_14-30-00_abc123/
├── bob_jones_2026-01-27_14-31-00_def456/
├── charlie_brown_2026-01-27_14-32-00_ghi789/
├── diana_prince_2026-01-27_14-33-00_jkl012/
└── eve_wilson_2026-01-27_14-34-00_mno345/
```

**Benefits:**
- All sessions in one place
- Easy to review all candidates
- Random UIDs prevent cross-contamination

**Security:**
- Each container gets random UID (10000-60000)
- Logs owned by root (UID 0)
- Even with shared volume, complete isolation

---

## Workflow Examples

### Example 1: 5 Simultaneous Interviews

```bash
# 1. Prepare
git pull origin master
docker compose build
chmod 777 sessions/

# 2. Set names
export CANDIDATE_1="alice"
export CANDIDATE_2="bob"
export CANDIDATE_3="charlie"
export CANDIDATE_4="diana"
export CANDIDATE_5="eve"

# 3. Start all
docker compose up -d

# 4. Send URLs to candidates
echo "Alice: http://server:6780"
echo "Bob: http://server:6781"
echo "Charlie: http://server:6782"
echo "Diana: http://server:6783"
echo "Eve: http://server:6784"

# 5. Monitor
make docker-status
watch -n 10 'make docker-status'

# 6. After interviews, review sessions
./scripts/analyze.sh sessions/alice_*
./scripts/analyze.sh sessions/bob_*
# ...

# 7. Cleanup
docker compose down
```

### Example 2: Rolling Interviews

```bash
# Start instance 1 for first candidate
CANDIDATE_1="alice" INSTANCE=1 make docker-compose-one

# When alice finishes, start instance 1 for next candidate
docker compose stop echobox-prod-1
CANDIDATE_1="frank" docker compose up -d echobox-prod-1

# Or use instance 2
CANDIDATE_2="bob" INSTANCE=2 make docker-compose-one
```

### Example 3: Staggered Start

```bash
# 10:00 AM - Start alice
CANDIDATE_1="alice" docker compose up -d echobox-prod-1

# 10:30 AM - Start bob
CANDIDATE_2="bob" docker compose up -d echobox-prod-2

# 11:00 AM - Start charlie
CANDIDATE_3="charlie" docker compose up -d echobox-prod-3

# Monitor active sessions
make docker-status
```

---

## Resource Considerations

### Memory Usage

Each instance uses:
- **Memory**: 512MB limit, 256MB reservation
- **CPU**: 0.5 core (50% of one CPU)

**5 instances total:**
- Memory: ~2.5GB (512MB × 5)
- CPU: ~2.5 cores (0.5 × 5)

**Recommendation:** Host with 4GB+ RAM, 4+ CPU cores

### Concurrent Interviews

**Light load (2-3 active):**
- Any modern server works
- 4GB RAM, 2 CPUs sufficient

**Full load (all 5 active):**
- Recommended: 8GB RAM, 4 CPUs
- Ensures smooth performance

### Scaling Beyond 5

To run more than 5 instances:

1. **Add more services** in docker-compose.yml:
   ```yaml
   echobox-prod-6:
     # ... same config ...
     ports:
       - "6785:8080"
     environment:
       CANDIDATE_NAME: ${CANDIDATE_6:-candidate-6}
   ```

2. **Or use docker run**:
   ```bash
   for i in {6..10}; do
     DOCKER_PORT=$((6780 + i - 1)) \
     CANDIDATE_NAME="candidate-$i" \
     make docker-run-prod &
   done
   ```

---

## Monitoring

### Real-Time Status

```bash
# Watch all instances
watch -n 5 'make docker-status'

# Check resource usage
docker stats
```

### Health Checks

```bash
# Check all instances
for port in {6780..6784}; do
  echo "Port $port:"
  curl -s http://localhost:$port/health | jq
done
```

### Session Progress

```bash
# List active sessions
ls -lt sessions/ | head -10

# Check latest logs
tail -f sessions/*/terminal.log
```

---

## Troubleshooting

### Port Conflicts

```bash
# Find what's using a port
lsof -i:6780

# Use different ports in docker-compose.yml
ports:
  - "7780:8080"  # Changed from 6780
```

### Instance Won't Start

```bash
# Check logs
docker compose logs echobox-prod-1

# Check if port is available
nc -zv localhost 6780

# Restart specific instance
docker compose restart echobox-prod-1
```

### Out of Memory

```bash
# Check memory usage
docker stats

# Stop unused instances
docker compose stop echobox-prod-4 echobox-prod-5

# Or increase host RAM
```

---

## Best Practices

### Naming

✅ **Good names** (descriptive):
```bash
CANDIDATE_1="alice_smith_backend"
CANDIDATE_2="bob_jones_devops"
```

❌ **Bad names** (hard to track):
```bash
CANDIDATE_1="test1"
CANDIDATE_2="abc"
```

### Session Management

1. **Start instances just before interviews**
2. **Monitor progress** with `docker-status`
3. **Stop after completion** to free resources
4. **Review sessions** before next batch

### Security

1. **Different candidate names** for each instance
2. **Unique URLs** (different ports)
3. **Firewall rules** for allowed ports only
4. **Shared volume safe** (random UID isolation)

---

## Quick Reference

```bash
# Start all 5
make docker-compose-up

# Check status
make docker-status

# View logs (instance 1)
INSTANCE=1 make docker-logs

# Shell access (instance 2)
INSTANCE=2 make docker-exec

# Stop all
docker compose down

# Review sessions
./scripts/analyze.sh sessions/*
```

---

## Summary

✅ **5 production instances** ready simultaneously
✅ **Ports**: 6780-6784 (one per instance)
✅ **Isolated**: Random UIDs, shared volume safe
✅ **Easy naming**: Set CANDIDATE_1-5 env vars
✅ **Simple management**: Makefile targets for everything
✅ **Monitored**: Status, logs, health checks

Perfect for conducting multiple SRE interviews in parallel! 🚀
