#!/bin/bash
# Setup candidate environment
# This script is called when PTY is created to prepare the candidate's home directory

CANDIDATE_HOME="$1"
TASK_DIR="/tasks/task-data"

# Create directories
mkdir -p "$CANDIDATE_HOME/solutions"
mkdir -p "$CANDIDATE_HOME/deploy"
mkdir -p "/var/log/nginx"
mkdir -p "/var/run"

# Copy task files
cp -f /tasks/INTERVIEW_TASKS.txt "$CANDIDATE_HOME/" 2>/dev/null || true
cp -f "$TASK_DIR/frankenstein.txt" "$CANDIDATE_HOME/" 2>/dev/null || true
cp -f "$TASK_DIR/nginx-access.log" "/var/log/nginx/access.log" 2>/dev/null || true
cp -f "$TASK_DIR/deploy/"*.sh "$CANDIDATE_HOME/deploy/" 2>/dev/null || true

# Copy mystery binary (architecture-specific)
ARCH=$(uname -m)
if [ "$ARCH" = "aarch64" ] || [ "$ARCH" = "arm64" ]; then
    cp -f "$TASK_DIR/mystery" "$CANDIDATE_HOME/" 2>/dev/null || true
elif [ "$ARCH" = "x86_64" ]; then
    cp -f "$TASK_DIR/mystery-amd64" "$CANDIDATE_HOME/mystery" 2>/dev/null || true
else
    # Fallback to ARM64
    cp -f "$TASK_DIR/mystery" "$CANDIDATE_HOME/" 2>/dev/null || true
fi
chmod +x "$CANDIDATE_HOME/mystery" 2>/dev/null || true

# Set ownership to the candidate UID
CANDIDATE_UID=$(stat -c %u "$CANDIDATE_HOME" 2>/dev/null || stat -f %u "$CANDIDATE_HOME" 2>/dev/null || echo "1000")
chown -R "$CANDIDATE_UID:$CANDIDATE_UID" "$CANDIDATE_HOME" 2>/dev/null || true
chown -R "$CANDIDATE_UID:$CANDIDATE_UID" "/var/log/nginx" 2>/dev/null || true

# Make solutions directory easily accessible
chmod 755 "$CANDIDATE_HOME/solutions"

echo "Candidate environment setup complete for $CANDIDATE_HOME"
