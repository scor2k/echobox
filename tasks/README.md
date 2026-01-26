# Interview Tasks

Welcome to your SRE technical interview!

## Instructions

This directory contains the tasks you need to complete during your interview session.

### Environment

You are working in an isolated Linux environment with the following tools pre-installed:
- **Editors**: vim, nano
- **Network**: curl, wget, netstat, ss
- **Process**: ps, top, htop
- **System**: systemctl, journalctl
- **Text Processing**: grep, awk, sed, jq
- **Debugging**: strace, tcpdump
- **Scripting**: bash, sh

### Guidelines

1. **Read each task carefully** before starting
2. **Save your solutions** in `~/solutions/` directory
3. **Document your approach** - comments in scripts are helpful
4. **Test your solutions** before moving to the next task
5. **Ask questions** if task requirements are unclear (in real interview)

### Task Structure

Each task is in its own directory:
```
tasks/
├── 01-debugging/
│   ├── scenario.md      # Problem description
│   ├── broken-app/      # Files to debug
│   └── hints.md         # Optional hints
├── 02-incident-response/
│   └── ...
└── 03-automation/
    └── ...
```

### Submission

Save all your work in `~/solutions/`:
```bash
mkdir -p ~/solutions
cd ~/solutions

# Example structure:
# task1-debugging/
#   ├── fix.sh
#   ├── explanation.txt
#   └── fixed-config.conf
# task2-incident/
#   └── analysis.md
```

### Time Management

- You have a limited time to complete all tasks
- Prioritize based on your strengths
- It's okay if you don't finish everything
- Quality over quantity

### Notes

- Your entire session is being recorded
- All commands and keystrokes are logged
- Copy-paste is disabled for assessment integrity
- This is a learning experience - do your best!

## Available Tasks

Tasks will be added to subdirectories. Check each directory for specific instructions.

### Placeholder Tasks

*Note: Actual interview tasks will be provided by the interviewer. Below are placeholders.*

1. **01-debugging**: Debug a broken web application
2. **02-incident-response**: Investigate high CPU usage
3. **03-automation**: Write a log parsing script

Good luck! 🚀
