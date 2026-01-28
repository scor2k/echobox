#!/bin/bash
# Setup candidate environment
# This script is called when PTY is created to prepare the candidate's home directory

CANDIDATE_HOME="$1"
TASK_DIR="/tasks/task-data"

# Create directories
mkdir -p "$CANDIDATE_HOME/solutions"

# Copy interview tasks file
cp -f /tasks/INTERVIEW_TASKS.txt "$CANDIDATE_HOME/" 2>/dev/null || true

# Copy task-data files to candidate home (excluding mystery-src)
for item in "$TASK_DIR"/*; do
    [ "$(basename "$item")" = "mystery-src" ] && continue
    cp -rf "$item" "$CANDIDATE_HOME/" 2>/dev/null || true
done

# Handle architecture-specific mystery binary
ARCH=$(uname -m)
if [ "$ARCH" = "x86_64" ]; then
    # On x86_64, use the amd64 version
    mv -f "$CANDIDATE_HOME/mystery-amd64" "$CANDIDATE_HOME/mystery" 2>/dev/null || true
fi
# Remove the architecture we don't need
rm -f "$CANDIDATE_HOME/mystery-amd64" 2>/dev/null || true
chmod +x "$CANDIDATE_HOME/mystery" 2>/dev/null || true

# Create .bash_profile to source .bashrc
cat > "$CANDIDATE_HOME/.bash_profile" << 'PROFILE'
# Source .bashrc if it exists
if [ -f ~/.bashrc ]; then
    . ~/.bashrc
fi
PROFILE

# Create .bashrc with welcome message
cat > "$CANDIDATE_HOME/.bashrc" << 'BASHRC'
# Candidate .bashrc

# Show welcome message on login
cat << 'MOTD'

╔══════════════════════════════════════════════════════════════╗
║                  SRE TECHNICAL INTERVIEW                     ║
╚══════════════════════════════════════════════════════════════╝

Welcome! You have been connected to an isolated interview environment.

INSTRUCTIONS:
  - Read ~/INTERVIEW_TASKS.txt for your assignments
  - Save your solutions in ~/solutions/
  - Your session is being recorded for evaluation
  - Use the "Finish" button when you're done

NOTES:
  - Copy-paste is disabled for assessment integrity
  - If you lose connection, refresh to reconnect
  - All commands and keystrokes are logged

Good luck!

MOTD

# Standard bash settings
export PS1='\[\033[01;32m\]candidate@interview\[\033[00m\]:\[\033[01;34m\]\w\[\033[00m\]\$ '
export EDITOR=vim
alias ll='ls -la'
alias l='ls -CF'

cd ~
BASHRC

# Set ownership to the candidate UID
CANDIDATE_UID=$(stat -c %u "$CANDIDATE_HOME" 2>/dev/null || stat -f %u "$CANDIDATE_HOME" 2>/dev/null || echo "1000")
chown -R "$CANDIDATE_UID:$CANDIDATE_UID" "$CANDIDATE_HOME" 2>/dev/null || true

# Make solutions directory easily accessible
chmod 755 "$CANDIDATE_HOME/solutions"

echo "Candidate environment setup complete for $CANDIDATE_HOME"
