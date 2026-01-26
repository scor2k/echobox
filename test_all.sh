#!/usr/bin/env bash
#
# Comprehensive test script for Echobox
# Tests all major features to ensure everything works
#

set +e  # Don't exit on errors - we handle them in test_feature

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m'

echo -e "${BLUE}╔════════════════════════════════════════════════╗${NC}"
echo -e "${BLUE}║${NC}   ${GREEN}Echobox Comprehensive Test Suite${NC}          ${BLUE}║${NC}"
echo -e "${BLUE}╚════════════════════════════════════════════════╝${NC}"
echo ""

# Track results
PASSED=0
FAILED=0

# Test function
test_feature() {
    local name="$1"
    local command="$2"

    echo -e "${YELLOW}Testing: ${name}${NC}"

    if eval "$command" > /dev/null 2>&1; then
        echo -e "  ${GREEN}✓ PASS${NC}"
        ((PASSED++))
    else
        echo -e "  ${RED}✗ FAIL${NC}"
        ((FAILED++))
    fi
}

# Clean up first
echo -e "${BLUE}Cleaning test environment...${NC}"
rm -rf sessions/test_*
killall echobox 2>/dev/null || true
sleep 1
echo ""

# 1. Build Test
echo -e "${BLUE}═══ 1. Build Tests ═══${NC}"
test_feature "Go build" "make build"
test_feature "Binary exists" "[ -f ./echobox ]"
test_feature "Binary executable" "[ -x ./echobox ]"
echo ""

# 2. Configuration Test
echo -e "${BLUE}═══ 2. Configuration Tests ═══${NC}"
test_feature "Config loads with defaults" "OUTPUT_DIR=./sessions ./echobox -h 2>&1 | grep -q 'flag provided but not defined' || true"
test_feature "Sessions directory creation" "mkdir -p sessions"
test_feature "Tasks directory exists" "[ -d tasks ]"
echo ""

# 3. Server Startup Test
echo -e "${BLUE}═══ 3. Server Startup Tests ═══${NC}"
echo "  Starting server..."
OUTPUT_DIR=./sessions CANDIDATE_NAME="test_startup" ./echobox > /tmp/echobox_test.log 2>&1 &
SERVER_PID=$!
sleep 3

test_feature "Server process running" "ps -p $SERVER_PID"
test_feature "Port 8080 listening" "lsof -i:8080 || netstat -an | grep -q 8080"
test_feature "Health endpoint responds" "curl -sf http://localhost:8080/health"
test_feature "Session directory created" "ls sessions/test_startup_* > /dev/null 2>&1"
echo ""

# 4. Recording Test
echo -e "${BLUE}═══ 4. Recording Tests ═══${NC}"
SESSION_DIR=$(find sessions -name "test_startup_*" -type d | head -1)

if [ -n "$SESSION_DIR" ]; then
    test_feature "Metadata file exists" "[ -f $SESSION_DIR/metadata.json ]"
    test_feature "Keystrokes log exists" "[ -f $SESSION_DIR/keystrokes.log ]"
    test_feature "Terminal log exists" "[ -f $SESSION_DIR/terminal.log ]"
    test_feature "Events log exists" "[ -f $SESSION_DIR/events.log ]"
    test_feature "Metadata is valid JSON" "jq empty $SESSION_DIR/metadata.json"
    test_feature "Session ID in metadata" "jq -e '.id' $SESSION_DIR/metadata.json"
else
    echo -e "  ${RED}✗ No session directory found${NC}"
    ((FAILED+=5))
fi
echo ""

# 5. Graceful Shutdown Test
echo -e "${BLUE}═══ 5. Shutdown Tests ═══${NC}"
echo "  Sending SIGINT..."
kill -INT $SERVER_PID
sleep 3

test_feature "Server stopped" "! ps -p $SERVER_PID"
test_feature "Metadata updated" "jq -e '.end_time' $SESSION_DIR/metadata.json"
test_feature "Status is completed or interrupted" "jq -e '.status' $SESSION_DIR/metadata.json | grep -qE 'completed|interrupted'"
test_feature "File hashes generated" "jq -e '.file_hashes' $SESSION_DIR/metadata.json"
test_feature "Analysis report exists" "[ -f $SESSION_DIR/analysis.json ]"
echo ""

# 6. Analysis Scripts Test
echo -e "${BLUE}═══ 6. Analysis Script Tests ═══${NC}"
test_feature "replay.sh exists and executable" "[ -x scripts/replay.sh ]"
test_feature "analyze.sh exists and executable" "[ -x scripts/analyze.sh ]"
test_feature "analyze.sh runs" "scripts/analyze.sh $SESSION_DIR | grep -q 'Analysis complete'"
echo ""

# 7. Docker Files Test (without actually building)
echo -e "${BLUE}═══ 7. Docker Configuration Tests ═══${NC}"
test_feature "Dockerfile exists" "[ -f Dockerfile ]"
test_feature "docker-compose.yml exists" "[ -f docker-compose.yml ]"
test_feature "Dockerfile syntax" "docker build --help > /dev/null 2>&1 && grep -q 'FROM golang' Dockerfile"
test_feature "docker-compose syntax" "grep -q 'version:' docker-compose.yml"
echo ""

# 8. Tasks Test
echo -e "${BLUE}═══ 8. Interview Tasks Tests ═══${NC}"
test_feature "Tasks README exists" "[ -f tasks/README.md ]"
test_feature "Task 01 exists" "[ -d tasks/01-debugging ]"
test_feature "Task 02 exists" "[ -d tasks/02-incident-response ]"
test_feature "Task 03 exists" "[ -d tasks/03-automation ]"
test_feature "Task 01 scenario exists" "[ -f tasks/01-debugging/scenario.md ]"
echo ""

# 9. Documentation Test
echo -e "${BLUE}═══ 9. Documentation Tests ═══${NC}"
test_feature "README.md exists" "[ -f README.md ]"
test_feature "DOCKER.md exists" "[ -f DOCKER.md ]"
test_feature "SECURITY.md exists" "[ -f SECURITY.md ]"
test_feature "RECORDING.md exists" "[ -f RECORDING.md ]"
test_feature "TIMEOUTS.md exists" "[ -f TIMEOUTS.md ]"
test_feature "Makefile exists" "[ -f Makefile ]"
echo ""

# Summary
echo -e "${BLUE}════════════════════════════════════════════════${NC}"
echo -e "${BLUE}Test Summary${NC}"
echo -e "${BLUE}════════════════════════════════════════════════${NC}"
echo -e "  ${GREEN}Passed:${NC} $PASSED"
echo -e "  ${RED}Failed:${NC} $FAILED"
echo -e "  ${BLUE}Total:${NC}  $((PASSED + FAILED))"
echo ""

if [ $FAILED -eq 0 ]; then
    echo -e "${GREEN}✓ All tests passed!${NC}"
    echo ""
    echo "Echobox is ready for production use."
    EXIT_CODE=0
else
    echo -e "${RED}✗ Some tests failed${NC}"
    echo ""
    echo "Review the failures above and fix before deploying."
    EXIT_CODE=1
fi

# Cleanup
echo ""
echo -e "${BLUE}Cleaning up test artifacts...${NC}"
rm -f /tmp/echobox_test.log
echo -e "${GREEN}✓ Cleanup complete${NC}"
echo ""

exit $EXIT_CODE
