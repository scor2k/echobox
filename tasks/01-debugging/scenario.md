# Task 1: Debug the Broken Web Application

## Scenario

A simple web server application is failing to start. Your job is to identify and fix the issue(s).

## Problem Description

The development team deployed a Python web server, but it's not responding. The application should be listening on port 8000 and serving a simple "Hello World" page.

## Your Mission

1. Identify why the application is not working
2. Fix the issue(s)
3. Verify the application runs correctly
4. Document what was wrong and how you fixed it

## Files

The broken application is in this directory:
- `app.py` - Main application file
- `config.json` - Configuration file
- `start.sh` - Startup script

## Expected Outcome

After fixing, you should be able to:
```bash
./start.sh
# Application should start successfully

curl http://localhost:8000
# Should return: Hello, SRE Candidate!
```

## Save Your Solution

Create a directory in `~/solutions/task1-debugging/` with:
1. The fixed files
2. A text file explaining what was broken and how you fixed it

## Hints

<details>
<summary>Click to reveal hints (try solving first!)</summary>

- Check file permissions
- Verify port availability
- Check configuration syntax
- Look at file paths
- Consider dependencies
</details>

## Time Estimate

15-20 minutes
