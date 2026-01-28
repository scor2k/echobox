#!/bin/bash
# Setup candidate environment
# This script is called when PTY is created to prepare the candidate's home directory

CANDIDATE_HOME="$1"
TASK_DIR="/tasks/task-data"

# Create directories
mkdir -p "$CANDIDATE_HOME/solutions"

# Copy interview tasks file
cp -f /tasks/INTERVIEW_TASKS.txt "$CANDIDATE_HOME/" 2>/dev/null || true

# Copy all task-data files to candidate home
cp -rf "$TASK_DIR"/* "$CANDIDATE_HOME/" 2>/dev/null || true

# Handle architecture-specific mystery binary
ARCH=$(uname -m)
if [ "$ARCH" = "x86_64" ]; then
    # On x86_64, use the amd64 version
    mv -f "$CANDIDATE_HOME/mystery-amd64" "$CANDIDATE_HOME/mystery" 2>/dev/null || true
fi
# Remove the architecture we don't need
rm -f "$CANDIDATE_HOME/mystery-amd64" 2>/dev/null || true
chmod +x "$CANDIDATE_HOME/mystery" 2>/dev/null || true

# Set ownership to the candidate UID
CANDIDATE_UID=$(stat -c %u "$CANDIDATE_HOME" 2>/dev/null || stat -f %u "$CANDIDATE_HOME" 2>/dev/null || echo "1000")
chown -R "$CANDIDATE_UID:$CANDIDATE_UID" "$CANDIDATE_HOME" 2>/dev/null || true

# Make solutions directory easily accessible
chmod 755 "$CANDIDATE_HOME/solutions"

echo "Candidate environment setup complete for $CANDIDATE_HOME"
