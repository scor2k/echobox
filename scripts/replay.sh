#!/usr/bin/env bash
#
# replay.sh - Replay recorded terminal sessions
#
# Usage: ./replay.sh <session_directory>
#

set -e

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Check if session directory is provided
if [ $# -eq 0 ]; then
    echo -e "${RED}Error: No session directory specified${NC}"
    echo ""
    echo "Usage: $0 <session_directory>"
    echo ""
    echo "Example:"
    echo "  $0 sessions/john_doe_2024-01-26_14-30-00_a3f7b9c1"
    echo ""
    echo "Available sessions:"
    find sessions -maxdepth 1 -type d -name "*_*_*" 2>/dev/null | sort -r | head -10
    exit 1
fi

SESSION_DIR="$1"

# Check if directory exists
if [ ! -d "$SESSION_DIR" ]; then
    echo -e "${RED}Error: Directory not found: $SESSION_DIR${NC}"
    exit 1
fi

# Check if required files exist
TERMINAL_LOG="$SESSION_DIR/terminal.log"
TIMING_LOG="$SESSION_DIR/timing.log"
METADATA="$SESSION_DIR/metadata.json"

if [ ! -f "$TERMINAL_LOG" ]; then
    echo -e "${RED}Error: terminal.log not found in $SESSION_DIR${NC}"
    exit 1
fi

if [ ! -f "$TIMING_LOG" ]; then
    echo -e "${RED}Error: timing.log not found in $SESSION_DIR${NC}"
    exit 1
fi

# Display session metadata
echo -e "${BLUE}╔═══════════════════════════════════════════════════╗${NC}"
echo -e "${BLUE}║${NC}         ${GREEN}Session Replay${NC}                           ${BLUE}║${NC}"
echo -e "${BLUE}╚═══════════════════════════════════════════════════╝${NC}"
echo ""

if [ -f "$METADATA" ]; then
    echo -e "${YELLOW}Session Information:${NC}"
    if command -v jq &> /dev/null; then
        jq -r '
            "  Candidate:     \(.candidate_name)",
            "  Session ID:    \(.id)",
            "  Start Time:    \(.start_time)",
            "  End Time:      \(.end_time)",
            "  Duration:      \(.duration_seconds)s",
            "  Status:        \(.status)"
        ' "$METADATA"
    else
        cat "$METADATA"
    fi
    echo ""
fi

# Check for scriptreplay command
if ! command -v scriptreplay &> /dev/null; then
    echo -e "${RED}Error: 'scriptreplay' command not found${NC}"
    echo ""
    echo "Install it with:"
    echo "  macOS: brew install util-linux"
    echo "  Ubuntu/Debian: sudo apt-get install bsdutils"
    echo ""
    exit 1
fi

# Show file hashes if available
if [ -f "$METADATA" ] && command -v jq &> /dev/null; then
    HASH_COUNT=$(jq -r '.file_hashes | length' "$METADATA" 2>/dev/null || echo "0")
    if [ "$HASH_COUNT" -gt 0 ]; then
        echo -e "${YELLOW}File Integrity (SHA-256):${NC}"
        jq -r '.file_hashes | to_entries[] | "  \(.key): \(.value)"' "$METADATA"
        echo ""
    fi
fi

# Ask for playback speed
echo -e "${YELLOW}Playback Options:${NC}"
echo "  1) Real-time (1x speed)"
echo "  2) Fast (2x speed)"
echo "  3) Very fast (5x speed)"
echo ""
read -p "Select option [1-3, default: 1]: " SPEED_OPTION

case "$SPEED_OPTION" in
    2)
        SPEED_DIVISOR=2
        echo -e "${GREEN}▶ Playing at 2x speed${NC}"
        ;;
    3)
        SPEED_DIVISOR=5
        echo -e "${GREEN}▶ Playing at 5x speed${NC}"
        ;;
    *)
        SPEED_DIVISOR=1
        echo -e "${GREEN}▶ Playing at real-time speed${NC}"
        ;;
esac

echo ""
echo -e "${BLUE}═══════════════════════════════════════════════════${NC}"
echo ""
sleep 1

# Replay the session
scriptreplay -t "$TIMING_LOG" -s "$TERMINAL_LOG" -d "$SPEED_DIVISOR"

echo ""
echo -e "${BLUE}═══════════════════════════════════════════════════${NC}"
echo -e "${GREEN}✓ Replay complete${NC}"
echo ""
echo -e "${YELLOW}Other available logs:${NC}"
ls -lh "$SESSION_DIR"/*.log 2>/dev/null | awk '{print "  " $9 " (" $5 ")"}'
echo ""
