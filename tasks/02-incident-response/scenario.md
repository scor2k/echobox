# Task 2: Incident Response - High CPU Usage

## Scenario

A production server is experiencing high CPU usage. Multiple processes are running slowly, and users are complaining about performance issues.

## Problem Description

Your monitoring system has alerted that CPU usage is at 95%+ on a production web server. You need to identify the root cause and resolve the issue.

## Your Mission

1. Identify which process is consuming excessive CPU
2. Determine the root cause
3. Stop the problematic process
4. Implement a solution to prevent recurrence
5. Document your findings and remediation steps

## Files

In this directory:
- `check_system.sh` - Script that starts a simulated problematic process
- `app.log` - Application log file with clues

## Starting the Scenario

```bash
cd /tasks/02-incident-response
./check_system.sh &
```

This will simulate the incident. Now investigate!

## Investigation Tools

Use these commands:
```bash
# Check CPU usage
top
htop

# Find processes
ps aux | sort -nrk 3 | head -10

# Check process details
ps -p <PID> -f
pstree -p <PID>

# Kill process
kill <PID>
kill -9 <PID>  # Force kill if needed
```

## Expected Outcome

- Identify the CPU-hogging process
- Successfully terminate it
- CPU usage returns to normal (<20%)
- Document root cause

## Save Your Solution

Create `~/solutions/task2-incident/` with:
1. `investigation_notes.md` - Your investigation process
2. `root_cause.txt` - What caused the issue
3. `remediation.sh` - Script to fix/prevent the issue

## Time Estimate

15-20 minutes

## Success Criteria

- Problematic process identified and stopped
- Root cause documented
- Prevention strategy proposed
