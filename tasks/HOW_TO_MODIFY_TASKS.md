# How to Modify Interview Tasks

Tasks are **mounted as a volume** - you can modify them without rebuilding the Docker image!

## Quick Task Modification

### 1. Edit Task Files (No Rebuild Needed!)

```bash
# Stop container
docker compose down

# Edit tasks - just change the text files!
vim tasks/INTERVIEW_TASKS.txt
vim tasks/task-data/nginx-access.log
vim tasks/task-data/frankenstein.txt

# Start container - new tasks loaded immediately!
docker compose up -d echobox-prod
```

**That's it!** No docker build required. Tasks are live-updated via volume mount.

---

## Task File Structure

```
tasks/
├── INTERVIEW_TASKS.txt           # Main task instructions (seen by candidate)
├── HOW_TO_MODIFY_TASKS.md        # This file (for admins)
├── setup-candidate-env.sh        # Environment setup script
└── task-data/                    # Task data files
    ├── nginx-access.log          # Task 1: Log file for analysis
    ├── frankenstein.txt          # Task 2: Text for word frequency
    └── deploy/                   # Task 3: Broken scripts
        ├── deploy.sh
        ├── backup.sh
        └── restart.sh
```

---

## Modifying Existing Tasks

### Task 1: Saskatoon (Log Analysis)

**File**: `tasks/task-data/nginx-access.log`

**To change**:
```bash
# Edit the log file
vim tasks/task-data/nginx-access.log

# Add/remove/modify log entries
# Change IP addresses, URLs, status codes, etc.
```

**Answer**: The IP with most requests (count them manually or with script)

### Task 2: Marrakech (Word Frequency)

**File**: `tasks/task-data/frankenstein.txt`

**To change**:
```bash
# Edit the text file
vim tasks/task-data/frankenstein.txt

# Change words to adjust frequency
# Be careful to know what the 2nd most frequent word is!
```

**Current answer**: MONSTER (appears 12 times, second after "the")

### Task 3: Kampala (Broken Scripts)

**Files**: `tasks/task-data/deploy/*.sh`

**Issue**: Scripts have DOS line endings (CRLF instead of LF)

**To change**:
```bash
# Create different broken scripts
vim tasks/task-data/deploy/deploy.sh

# Or add more scripts
cp tasks/task-data/deploy/deploy.sh tasks/task-data/deploy/monitor.sh

# Re-add DOS line endings if needed:
unix2dos tasks/task-data/deploy/*.sh
```

**Current issue**: CRLF line endings
**Fix**: dos2unix, sed 's/\r$//', or tr -d '\r'

### Task 4: Berlin (Binary Detective)

**Files**: `tasks/task-data/mystery` (ARM64), `tasks/task-data/mystery-amd64` (x86_64)

**Issue**: Binary exits immediately (looks for .mystery.lock file)

**To change**:
```bash
# Rebuild with different requirement
cd tasks/task-data/mystery-src
vim main.go
# Change lockFile := ".mystery.lock" to something else
# e.g., lockFile := ".config/app.conf"

# Rebuild for both architectures
GOOS=linux GOARCH=arm64 go build -ldflags="-s -w" -o ../mystery .
GOOS=linux GOARCH=amd64 go build -ldflags="-s -w" -o ../mystery-amd64 .
cd ../../..

# Restart container
docker compose down && docker compose up -d
```

**Current requirement**: `.mystery.lock` file in current directory
**Fix**: `touch .mystery.lock`

---

## Creating New Tasks

### Add Task 4

1. **Add to INTERVIEW_TASKS.txt**:
   ```bash
   vim tasks/INTERVIEW_TASKS.txt
   # Add new section at the end
   ```

2. **Create data files**:
   ```bash
   # Create data in task-data/
   echo "data" > tasks/task-data/task4-data.txt
   ```

3. **Update setup script**:
   ```bash
   vim tasks/setup-candidate-env.sh
   # Add line to copy new task file
   ```

4. **Restart containers**:
   ```bash
   docker compose down
   docker compose up -d echobox-prod
   ```

---

## Testing Tasks Locally

### Test Setup Script

```bash
# Run setup script manually
mkdir -p /tmp/test-home
sudo bash tasks/setup-candidate-env.sh /tmp/test-home

# Check what got copied
ls -la /tmp/test-home/
cat /tmp/test-home/INTERVIEW_TASKS.txt

# Cleanup
rm -rf /tmp/test-home
```

### Test in Container

```bash
# Start container
docker compose up -d echobox-dev

# Check task files
docker exec -it echobox-dev ls -la /home/candidate-*/

# Check specific files
docker exec -it echobox-dev cat /tasks/INTERVIEW_TASKS.txt

# Stop
docker compose down
```

---

## Volume Mount Configuration

**In docker-compose.yml**:
```yaml
volumes:
  - ./tasks:/tasks:ro   # Read-only mount
```

**This means:**
- ✅ Edit files in `./tasks/` on host
- ✅ Changes immediately visible in container
- ✅ No rebuild needed
- ✅ No restart needed (for file content changes)
- ⚠️ If you change setup-candidate-env.sh, restart container

---

## Task Answers (For Reference)

**Keep these secret from candidates!**

### Task 1: Saskatoon
```bash
# Answer: 192.168.1.42
# Solution:
awk '{print $1}' /var/log/nginx/access.log | sort | uniq -c | sort -rn | head -1 | awk '{print $2}'
```

### Task 2: Marrakech
```bash
# Answer: MONSTER
# Solution:
cat frankenstein.txt | tr '[:upper:]' '[:lower:]' | tr -d '[:punct:]' | \
  tr ' ' '\n' | grep -v '^$' | sort | uniq -c | sort -rn | head -2 | tail -1 | \
  awk '{print toupper($2)}'
```

### Task 3: Kampala
```bash
# Problem: DOS line endings (CRLF)
# Diagnosis: file deploy.sh shows "CRLF line terminators"
# Fix: dos2unix deploy/*.sh
# Or: sed -i 's/\r$//' deploy/*.sh
# Or: tr -d '\r' < deploy.sh > deploy.sh.fixed && mv deploy.sh.fixed deploy.sh
```

### Task 4: Berlin (Binary Detective)
```bash
# Investigation:
strace ./mystery 2>&1 | grep -E "open|stat|access"
# Shows: stat(".mystery.lock", ...) = -1 ENOENT (No such file or directory)

strings mystery | grep -i lock
# Shows: ".mystery.lock" (among other strings)

# Solution:
touch .mystery.lock
./mystery
# Success! Binary runs and prints:
# Mystery Application v1.0
# Status: Running successfully
# Lock file found: .mystery.lock
# Initialization complete
#
# ═══════════════════════════════════════════════
#   VERIFICATION CODE: SRE-DETECTIVE-427
# ═══════════════════════════════════════════════
#
# Copy this code to your solution file!

# Answer file format (~/solutions/berlin.txt):
# Line 1: SRE-DETECTIVE-427
# Line 2+: Explanation

# Example solution:
SRE-DETECTIVE-427

Used strace ./mystery 2>&1 | grep stat to trace system calls.
Found: stat(".mystery.lock") = -1 ENOENT
Created file: touch .mystery.lock
Binary ran successfully and printed verification code.

# Verification Code Details:
# The code is obfuscated in the binary:
# - part1: byte array (not visible as plain string)
# - part2: byte array (not visible as plain string)
# - part3: byte array (not visible as plain string)
# - suffix: generated from numbers (not plain string)
# This makes it harder to find with strings command
# Candidate must actually run the binary to get the code!
```

---

## Best Practices

### Task Design

1. **Clear objective** - One specific thing to find/fix
2. **Measurable output** - Exact format for solution file
3. **Realistic scenario** - Real-world SRE problems
4. **Time-appropriate** - 10-20 minutes per task
5. **Multiple approaches** - Different valid solutions

### File Formats

1. **Keep task data simple** - Plain text, standard formats
2. **Test answers** - Verify there's only ONE correct answer
3. **Avoid ambiguity** - Clear requirements, exact output format
4. **Document answers** - Keep reference solutions

### Difficulty Levels

- **Easy**: Single command, direct approach (Saskatoon)
- **Medium**: Multi-step, requires thinking (Marrakech)
- **Hard**: Debugging, investigation required (Kampala)

---

## Common Modifications

### Change IP in Task 1
```bash
# Edit nginx log
vim tasks/task-data/nginx-access.log

# Change all instances of 192.168.1.42 to different IP
:%s/192.168.1.42/10.10.10.100/g
```

### Change Word Frequency in Task 2
```bash
# Edit text
vim tasks/task-data/frankenstein.txt

# Add/remove words to change frequency
# Count words to verify: cat frankenstein.txt | tr ' ' '\n' | sort | uniq -c | sort -rn
```

### Create Different Script Error in Task 3
```bash
# Wrong shebang
echo '#!/usr/bin/python' > tasks/task-data/deploy/deploy.sh
echo 'print("test")' >> tasks/task-data/deploy/deploy.sh
# Error: python not at that path

# Missing execute permission
chmod 644 tasks/task-data/deploy/deploy.sh
# Error: Permission denied

# Wrong interpreter
echo '#!/bin/zsh' > tasks/task-data/deploy/deploy.sh
# Error: zsh not installed
```

---

## Validation

### After Candidate Finishes

Solutions are automatically saved to:
```
sessions/<candidate>_<timestamp>_<id>/solutions/
├── saskatoon.txt
├── marrakech.txt
└── kampala.txt
```

**Check answers**:
```bash
SESSION=$(ls -td sessions/*/ | head -1)

# Task 1
cat $SESSION/solutions/saskatoon.txt
# Expected: 192.168.1.42

# Task 2
cat $SESSION/solutions/marrakech.txt
# Expected: MONSTER

# Task 3
cat $SESSION/solutions/kampala.txt
# Should explain: DOS line endings, fix with dos2unix/sed/tr

# Task 4
cat $SESSION/solutions/berlin.txt
# Should explain: Used strace, found .mystery.lock missing, created it with touch
```

---

## Quick Task Template

```bash
# Create new task file
cat > tasks/task-data/my-task-data.txt << 'EOF'
[task data here]
EOF

# Update INTERVIEW_TASKS.txt
vim tasks/INTERVIEW_TASKS.txt
# Add:
# ═══════════════════════════════════════════════════════════════
# TASK X: "CityName" — Task Title
# ═══════════════════════════════════════════════════════════════
#
# [task description]
# Save to: ~/solutions/taskname.txt

# Update setup script
vim tasks/setup-candidate-env.sh
# Add: cp -f "$TASK_DIR/my-task-data.txt" "$CANDIDATE_HOME/" 2>/dev/null || true

# Restart (no rebuild!)
docker compose down
docker compose up -d echobox-prod
```

---

## Summary

✅ **No rebuild needed** - Tasks are mounted via volume
✅ **Easy to modify** - Just edit text files
✅ **Solutions captured** - Automatically saved to session output
✅ **Immediate updates** - Edit and restart (no docker build)
✅ **Version control** - Tasks in git, track changes

**Modify tasks as often as you like - it's just text files!** 📝
