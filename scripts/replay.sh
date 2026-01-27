#!/usr/bin/env bash
#
# replay.sh - View recorded terminal session output
#
# Usage: ./replay.sh <session_directory>
#        ./replay.sh <session_directory> > output.txt  # Save to file
#

set -e

# Colors (only if outputting to terminal)
if [ -t 1 ]; then
    RED='\033[0;31m'
    GREEN='\033[0;32m'
    YELLOW='\033[1;33m'
    BLUE='\033[0;34m'
    NC='\033[0m'
else
    # Disable colors if piping to file
    RED=''
    GREEN=''
    YELLOW=''
    BLUE=''
    NC=''
fi

# Check if session directory is provided
if [ $# -eq 0 ]; then
    echo -e "${RED}Error: No session directory specified${NC}" >&2
    echo "" >&2
    echo "Usage: $0 <session_directory>" >&2
    echo "" >&2
    echo "Examples:" >&2
    echo "  $0 sessions/john_doe_2024-01-26_14-30-00_a3f7b9c1" >&2
    echo "  $0 sessions/john_doe_2024-01-26_14-30-00_a3f7b9c1 > output.txt  # Save to file" >&2
    echo "" >&2
    echo "Available sessions:" >&2
    find sessions -maxdepth 1 -type d -name "*_*_*" 2>/dev/null | sort -r | head -10 >&2
    exit 1
fi

SESSION_DIR="$1"

# Check if directory exists
if [ ! -d "$SESSION_DIR" ]; then
    echo -e "${RED}Error: Directory not found: $SESSION_DIR${NC}"
    exit 1
fi

# Check if required file exists
TERMINAL_LOG="$SESSION_DIR/terminal.log"
METADATA="$SESSION_DIR/metadata.json"

if [ ! -f "$TERMINAL_LOG" ]; then
    echo -e "${RED}Error: terminal.log not found in $SESSION_DIR${NC}" >&2
    exit 1
fi

# Show session info to stderr (so it doesn't pollute piped output)
if [ -t 1 ]; then
    # Only show if outputting to terminal (not piping)
    echo -e "${BLUE}╔═══════════════════════════════════════════════════╗${NC}" >&2
    echo -e "${BLUE}║${NC}         ${GREEN}Session Terminal Output${NC}                  ${BLUE}║${NC}" >&2
    echo -e "${BLUE}╚═══════════════════════════════════════════════════╝${NC}" >&2
    echo "" >&2

    if [ -f "$METADATA" ]; then
        echo -e "${YELLOW}Session Information:${NC}" >&2
        if command -v jq &> /dev/null; then
            jq -r '
                "  Candidate:     \(.candidate_name)",
                "  Session ID:    \(.id)",
                "  Duration:      \(.duration_seconds)s",
                "  Status:        \(.status)"
            ' "$METADATA" >&2
        fi
        echo "" >&2
    fi

    echo -e "${GREEN}▶ Terminal Output${NC}" >&2
    echo -e "${BLUE}═══════════════════════════════════════════════════${NC}" >&2
    echo "" >&2
fi

# Dump terminal log (this goes to stdout, can be piped)
cat "$TERMINAL_LOG"

# Show footer to stderr (only if outputting to terminal)
if [ -t 1 ]; then
    echo "" >&2
    echo -e "${BLUE}═══════════════════════════════════════════════════${NC}" >&2
    echo -e "${GREEN}✓ Output complete${NC}" >&2
    echo "" >&2
    echo -e "${YELLOW}Tip: Pipe to file or grep to filter:${NC}" >&2
    echo "  $0 $SESSION_DIR > output.txt" >&2
    echo "  $0 $SESSION_DIR | grep strace" >&2
    echo "" >&2
fi
