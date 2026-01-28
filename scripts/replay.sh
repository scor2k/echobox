#!/usr/bin/env bash
#
# replay.sh - View recorded session keystrokes and commands
#
# Usage: ./replay.sh <session_directory>
#        ./replay.sh <session_directory> -c  # Show commands only
#

set -e

# Colors (only if outputting to terminal)
if [ -t 1 ]; then
    RED='\033[0;31m'
    GREEN='\033[0;32m'
    YELLOW='\033[1;33m'
    BLUE='\033[0;34m'
    CYAN='\033[0;36m'
    NC='\033[0m'
else
    RED=''
    GREEN=''
    YELLOW=''
    BLUE=''
    CYAN=''
    NC=''
fi

# Check if session directory is provided
if [ $# -eq 0 ]; then
    echo -e "${RED}Error: No session directory specified${NC}" >&2
    echo "" >&2
    echo "Usage: $0 <session_directory> [-c]" >&2
    echo "" >&2
    echo "Options:" >&2
    echo "  -c    Show extracted commands only (cleaner view)" >&2
    echo "" >&2
    echo "Examples:" >&2
    echo "  $0 sessions/john_doe_2024-01-26_14-30-00_a3f7b9c1" >&2
    echo "  $0 sessions/john_doe_2024-01-26_14-30-00_a3f7b9c1 -c" >&2
    echo "" >&2
    echo "Available sessions:" >&2
    find sessions -maxdepth 1 -type d -name "*_*_*" 2>/dev/null | sort -r | head -10 >&2
    exit 1
fi

SESSION_DIR="$1"
SHOW_COMMANDS=false

# Parse options
shift
while [ $# -gt 0 ]; do
    case "$1" in
        -c|--commands)
            SHOW_COMMANDS=true
            shift
            ;;
        *)
            echo -e "${RED}Unknown option: $1${NC}" >&2
            exit 1
            ;;
    esac
done

# Check if directory exists
if [ ! -d "$SESSION_DIR" ]; then
    echo -e "${RED}Error: Directory not found: $SESSION_DIR${NC}" >&2
    exit 1
fi

KEYSTROKES_LOG="$SESSION_DIR/keystrokes.log"
COMMANDS_LOG="$SESSION_DIR/commands.log"
METADATA="$SESSION_DIR/metadata.json"

# Show session info to stderr
if [ -t 1 ]; then
    echo -e "${BLUE}======================================${NC}" >&2
    echo -e "${GREEN}       Session Replay${NC}" >&2
    echo -e "${BLUE}======================================${NC}" >&2
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
fi

if [ "$SHOW_COMMANDS" = true ]; then
    # Show commands only
    if [ ! -f "$COMMANDS_LOG" ]; then
        echo -e "${RED}Error: commands.log not found in $SESSION_DIR${NC}" >&2
        echo "Commands are extracted when session ends." >&2
        exit 1
    fi

    if [ -t 1 ]; then
        echo -e "${GREEN}Extracted Commands:${NC}" >&2
        echo -e "${BLUE}--------------------------------------${NC}" >&2
    fi

    # Format: timestamp command
    while IFS= read -r line; do
        # Skip comments
        [[ "$line" =~ ^# ]] && continue
        [[ -z "$line" ]] && continue

        # Parse timestamp and command
        timestamp=$(echo "$line" | cut -d' ' -f1)
        command=$(echo "$line" | cut -d' ' -f2-)

        if [ -t 1 ]; then
            # Format timestamp as seconds
            secs=$((timestamp / 1000))
            ms=$((timestamp % 1000))
            printf "${CYAN}[%d.%03ds]${NC} %s\n" "$secs" "$ms" "$command"
        else
            echo "$command"
        fi
    done < "$COMMANDS_LOG"
else
    # Show keystrokes
    if [ ! -f "$KEYSTROKES_LOG" ]; then
        echo -e "${RED}Error: keystrokes.log not found in $SESSION_DIR${NC}" >&2
        exit 1
    fi

    if [ -t 1 ]; then
        echo -e "${GREEN}Keystrokes:${NC}" >&2
        echo -e "${BLUE}--------------------------------------${NC}" >&2
    fi

    # Format: timestamp_ms "keystroke"
    while IFS= read -r line; do
        [[ -z "$line" ]] && continue

        # Parse timestamp and keystroke
        timestamp=$(echo "$line" | cut -d' ' -f1)
        keystroke=$(echo "$line" | cut -d' ' -f2-)

        if [ -t 1 ]; then
            # Format timestamp as seconds
            secs=$((timestamp / 1000))
            ms=$((timestamp % 1000))
            printf "${CYAN}[%d.%03ds]${NC} %s\n" "$secs" "$ms" "$keystroke"
        else
            echo "$line"
        fi
    done < "$KEYSTROKES_LOG"
fi

if [ -t 1 ]; then
    echo "" >&2
    echo -e "${BLUE}--------------------------------------${NC}" >&2
    echo -e "${GREEN}Replay complete${NC}" >&2
    echo "" >&2
    echo -e "${YELLOW}Tips:${NC}" >&2
    echo "  $0 $SESSION_DIR -c     # Show commands only" >&2
    echo "  $0 $SESSION_DIR > out.txt  # Save to file" >&2
fi
