# Session Recording Documentation

## Overview

Echobox records all interview sessions with comprehensive logging for later review and analysis.

## Recorded Files

Each session creates a directory: `sessions/<candidate>_<timestamp>_<session_id>/`

### Files Created

| File | Description | Format |
|------|-------------|--------|
| `metadata.json` | Session info, hashes, status | JSON |
| `keystrokes.log` | Raw keystroke input with timestamps | `<ms> "<chars>"\n` |
| `terminal.log` | Complete terminal output | Script format (binary) |
| `timing.log` | Timing data for scriptreplay | `<seconds> <bytes>\n` |
| `websocket.log` | All WebSocket messages | `<ms> <direction> <type> <len> "<sample>"\n` |
| `events.log` | Anti-cheat events | `<ms> <event> <data>\n` |
| `commands.log` | Extracted shell commands | Text |

## Metadata Structure

```json
{
  "id": "a3f7b9c1",
  "candidate_name": "john_doe",
  "start_time": "2026-01-26T14:30:00Z",
  "end_time": "2026-01-26T15:45:00Z",
  "duration_seconds": 4500,
  "output_dir": "sessions/john_doe_2026-01-26_14-30-00_a3f7b9c1",
  "status": "completed",
  "file_hashes": {
    "keystrokes.log": "sha256_hash_here",
    "terminal.log": "sha256_hash_here",
    ...
  },
  "metadata": {
    "interrupted": false
  }
}
```

### Status Values
- `active` - Session in progress
- `completed` - Session finished normally
- `interrupted` - Session terminated by signal

## Recording Features

### Keystroke Logging
- Every character with millisecond-precision timestamp
- Format: `<timestamp_ms> "<character>"\n`
- Example: `1234567 "l"`

### Terminal Output (Script Format)
- Compatible with `scriptreplay` command
- Binary format capturing all output
- Paired with `timing.log` for replay

### Timing Log
- Scriptreplay-compatible timing data
- Format: `<seconds_since_last> <bytes>\n`
- Example: `0.123456 42`

### WebSocket Messages
- All client↔server messages
- Includes resize, finish, anticheat events
- Format: `<timestamp> <direction> <type> <length> "<sample>"`

### Events Log
- Anti-cheat events (paste attempts, focus loss, rapid input)
- Connection events
- Format: `<timestamp> <event_type> <data>`

### SHA-256 Integrity
- All log files hashed on session completion
- Hashes stored in metadata.json
- Verify with: `sha256sum <file>` vs metadata

## Progressive Persistence

- **Flush interval**: 10 seconds (configurable)
- **Buffer size**: 4KB per file
- **On crash**: Maximum 10 seconds of data loss
- **On shutdown**: All buffers flushed before exit

## Scripts

### Replay Session
```bash
./scripts/replay.sh sessions/john_doe_2026-01-26_14-30-00_a3f7b9c1
```

Features:
- Shows session metadata
- Verifies file integrity (SHA-256)
- Interactive speed selection (1x, 2x, 5x)
- Uses `scriptreplay` for authentic terminal replay

### Analyze Session
```bash
./scripts/analyze.sh sessions/john_doe_2026-01-26_14-30-00_a3f7b9c1
```

Reports:
- Session information
- File sizes and line counts
- File integrity verification
- Anti-cheat event summary
- Typing statistics (WPM estimate)
- Paste attempts and rapid input detection

## Usage Examples

### Start Interview
```bash
docker run -d \
  -v $(pwd)/sessions:/output \
  -e CANDIDATE_NAME="jane_doe" \
  -e SESSION_TIMEOUT=7200 \
  echobox:latest
```

### Review Session
```bash
# List sessions
ls sessions/

# Analyze
./scripts/analyze.sh sessions/jane_doe_*/

# Replay
./scripts/replay.sh sessions/jane_doe_*/
```

### Verify Integrity
```bash
cd sessions/jane_doe_*/

# Verify all files
jq -r '.file_hashes | to_entries[] | "\(.value)  \(.key)"' metadata.json | shasum -a 256 -c
```

## Anti-Cheat Detection

The recording system captures:
1. **Paste attempts** - Logged to events.log
2. **Rapid input** - >30 chars in <100ms
3. **Focus loss** - Tab switches, window blur
4. **Typing patterns** - Can calculate WPM and detect anomalies

### Example Events Log
```
1234 anticheat {"event":"paste_attempt","timestamp":1234567,"source":"paste_event"}
5678 anticheat {"event":"rapid_input","timestamp":5678901,"chars":45,"time_ms":80}
9012 anticheat {"event":"window_focus","timestamp":9012345,"gained":false}
```

## File Formats

### Keystrokes Log
```
0 "l"
52 "s"
105 " "
157 "-"
209 "l"
261 "\r"
```

### Timing Log (scriptreplay format)
```
0.123456 42
0.056789 18
0.234567 156
```

### WebSocket Log
```
1234 client->server text 1 "l"
1235 server->client text 42 "total 8\r\ndrwxr-xr-x 2 user..."
1456 client->server text 45 "{\"type\":\"resize\",\"data\":{\"cols\":80,\"rows\":24}}"
```

## Storage Considerations

### Typical Session Sizes
- 1 hour session: ~2-5 MB total
- Keystrokes: ~10-50 KB
- Terminal output: ~1-3 MB
- Events: <10 KB
- Metadata: <5 KB

### Retention
- No automatic cleanup
- Manual cleanup with: `rm -rf sessions/old_*`
- Or use Docker volume mounts to store elsewhere

## Security

- All files created with mode 0644 (read/write for owner)
- Session directory: 0755 (readable by others)
- No sensitive data in logs (except terminal contents)
- Hashes prevent tampering

## Troubleshooting

### Empty Log Files
- Check if WebSocket connected: `grep "WebSocket connected" <logs>`
- Verify candidate typed commands
- Check flush interval (default 10s)

### Missing Hashes
- Hashes only calculated on clean shutdown
- SIGKILL prevents hash calculation
- Use proper shutdown (SIGTERM, finish button)

### Replay Doesn't Work
- Requires `scriptreplay` command
- Install: `brew install util-linux` (macOS)
- Check terminal.log and timing.log exist and not empty

### Commands Log Empty
- Command extraction is basic placeholder
- Use `scriptreplay` for full session review
- Enhanced extraction coming in future updates
