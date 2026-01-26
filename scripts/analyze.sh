#!/usr/bin/env bash
#
# analyze.sh - Analyze session for anti-cheat patterns
#
# Usage: ./analyze.sh <session_directory>
#

set -e

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

if [ $# -eq 0 ]; then
    echo -e "${RED}Error: No session directory specified${NC}"
    echo ""
    echo "Usage: $0 <session_directory>"
    exit 1
fi

SESSION_DIR="$1"

if [ ! -d "$SESSION_DIR" ]; then
    echo -e "${RED}Error: Directory not found: $SESSION_DIR${NC}"
    exit 1
fi

echo -e "${BLUE}╔═══════════════════════════════════════════════════╗${NC}"
echo -e "${BLUE}║${NC}      ${GREEN}Session Analysis Report${NC}                     ${BLUE}║${NC}"
echo -e "${BLUE}╚═══════════════════════════════════════════════════╝${NC}"
echo ""

# Session info
METADATA="$SESSION_DIR/metadata.json"
if [ -f "$METADATA" ]; then
    echo -e "${YELLOW}Session Information:${NC}"
    if command -v jq &> /dev/null; then
        jq -r '
            "  Candidate:     \(.candidate_name)",
            "  Session ID:    \(.id)",
            "  Duration:      \(.duration_seconds)s",
            "  Status:        \(.status)"
        ' "$METADATA"
    fi
    echo ""
fi

# File sizes
echo -e "${YELLOW}Recorded Files:${NC}"
for file in keystrokes.log terminal.log timing.log websocket.log events.log commands.log; do
    if [ -f "$SESSION_DIR/$file" ]; then
        SIZE=$(wc -c < "$SESSION_DIR/$file" | tr -d ' ')
        LINES=$(wc -l < "$SESSION_DIR/$file" | tr -d ' ')
        printf "  %-18s %8s bytes, %6s lines\n" "$file" "$SIZE" "$LINES"
    fi
done
echo ""

# Verify file hashes
if [ -f "$METADATA" ] && command -v jq &> /dev/null; then
    echo -e "${YELLOW}File Integrity Check:${NC}"

    while IFS= read -r filename; do
        RECORDED_HASH=$(jq -r ".file_hashes[\"$filename\"]" "$METADATA")
        if [ "$RECORDED_HASH" != "null" ] && [ -f "$SESSION_DIR/$filename" ]; then
            ACTUAL_HASH=$(shasum -a 256 "$SESSION_DIR/$filename" | awk '{print $1}')
            if [ "$RECORDED_HASH" == "$ACTUAL_HASH" ]; then
                echo -e "  ${GREEN}✓${NC} $filename"
            else
                echo -e "  ${RED}✗${NC} $filename (hash mismatch!)"
            fi
        fi
    done < <(jq -r '.file_hashes | keys[]' "$METADATA")
    echo ""
fi

# Anti-cheat analysis
EVENTS_LOG="$SESSION_DIR/events.log"
if [ -f "$EVENTS_LOG" ]; then
    EVENTS_COUNT=$(wc -l < "$EVENTS_LOG" | tr -d ' ')
    echo -e "${YELLOW}Anti-Cheat Events: ${NC}$EVENTS_COUNT total"

    # Count paste attempts
    PASTE_COUNT=$(grep -c "paste_attempt" "$EVENTS_LOG" 2>/dev/null || echo "0")
    if [ "$PASTE_COUNT" -gt 0 ]; then
        echo -e "  ${RED}⚠${NC}  Paste attempts: $PASTE_COUNT"
    else
        echo -e "  ${GREEN}✓${NC}  No paste attempts detected"
    fi

    # Count rapid input
    RAPID_COUNT=$(grep -c "rapid_input" "$EVENTS_LOG" 2>/dev/null || echo "0")
    if [ "$RAPID_COUNT" -gt 0 ]; then
        echo -e "  ${YELLOW}⚠${NC}  Rapid input events: $RAPID_COUNT"
    else
        echo -e "  ${GREEN}✓${NC}  No rapid input detected"
    fi

    # Count focus loss
    BLUR_COUNT=$(grep -c '"gained":false' "$EVENTS_LOG" 2>/dev/null || echo "0")
    if [ "$BLUR_COUNT" -gt 0 ]; then
        echo -e "  ${YELLOW}ℹ${NC}  Focus loss events: $BLUR_COUNT"
    fi

    echo ""
fi

# Keystroke statistics
KEYSTROKES_LOG="$SESSION_DIR/keystrokes.log"
if [ -f "$KEYSTROKES_LOG" ]; then
    KEYSTROKE_COUNT=$(wc -l < "$KEYSTROKES_LOG" | tr -d ' ')
    echo -e "${YELLOW}Typing Statistics:${NC}"
    echo "  Total keystrokes: $KEYSTROKE_COUNT"

    # Calculate session duration in seconds
    if [ -f "$METADATA" ] && command -v jq &> /dev/null; then
        DURATION=$(jq -r '.duration_seconds // 0' "$METADATA")
        if [ "$DURATION" != "0" ]; then
            WPM=$(echo "scale=2; ($KEYSTROKE_COUNT / $DURATION) * 12" | bc 2>/dev/null || echo "N/A")
            echo "  Average WPM: ~$WPM"
        fi
    fi
    echo ""
fi

# Summary
echo -e "${BLUE}═══════════════════════════════════════════════════${NC}"
echo -e "${GREEN}Analysis complete!${NC}"
echo ""
echo "View logs:"
echo "  Keystrokes:   less $SESSION_DIR/keystrokes.log"
echo "  Events:       less $SESSION_DIR/events.log"
echo "  WebSocket:    less $SESSION_DIR/websocket.log"
echo "  Commands:     cat $SESSION_DIR/commands.log"
echo ""
echo "Replay session:"
echo "  ./scripts/replay.sh $SESSION_DIR"
echo ""
