# Task 3: Automation - Log Analysis Script

## Scenario

You need to create a bash script to parse application logs and generate a summary report. This is a common SRE task for incident analysis and monitoring.

## Problem Description

The application generates logs in `/tasks/03-automation/access.log`. You need to write a script that analyzes these logs and produces a summary report.

## Your Mission

Write a bash script that:

1. **Counts total requests** by HTTP method (GET, POST, PUT, DELETE)
2. **Identifies the top 5 most accessed endpoints**
3. **Finds all 4xx and 5xx errors** with their counts
4. **Calculates average response time**
5. **Identifies the slowest requests** (response time > 1000ms)
6. **Generates a summary report** in a readable format

## Input File

`access.log` - Standard web server access log format:
```
<timestamp> <method> <endpoint> <status> <response_time_ms> <ip_address>
```

Example:
```
2026-01-26T10:00:01Z GET /api/users 200 45 192.168.1.100
```

## Expected Output

Your script should generate a report like:
```
=== Log Analysis Report ===
Total Requests: 1000

Requests by Method:
  GET:    750
  POST:   200
  PUT:     30
  DELETE:  20

Top 5 Endpoints:
  1. /api/users (350 requests)
  2. /api/products (200 requests)
  ...

Error Summary:
  4xx errors: 25
  5xx errors: 5
  Total errors: 30

Performance:
  Average response time: 125ms
  Slowest requests: 3

Slow Requests (>1000ms):
  2026-01-26T10:05:23Z POST /api/import 1543ms
  ...
```

## Requirements

- Pure bash script (use: grep, awk, sed, sort, uniq, etc.)
- Handle edge cases (empty logs, malformed lines)
- Script should be executable: `chmod +x analyze_logs.sh`
- Include comments explaining your approach
- Make output human-readable

## Files Provided

- `access.log` - Sample log file with ~100 entries
- `large_access.log` - Larger log file for performance testing (optional)

## Save Your Solution

Create `~/solutions/task3-automation/` with:
1. `analyze_logs.sh` - Your analysis script
2. `output_sample.txt` - Sample output from your script
3. `explanation.md` - How your script works

## Testing Your Script

```bash
cd ~/solutions/task3-automation
chmod +x analyze_logs.sh
./analyze_logs.sh /tasks/03-automation/access.log
```

## Bonus Points

- Handle multiple log files
- Add command-line options (e.g., `--errors-only`, `--top N`)
- Color-coded output
- JSON output option

## Time Estimate

20-30 minutes

## Success Criteria

- Script processes access.log correctly
- All required metrics calculated
- Output is clear and well-formatted
- Script handles errors gracefully
