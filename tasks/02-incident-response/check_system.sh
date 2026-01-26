#!/bin/bash
# Simulated problematic process
# This script creates a CPU-intensive background process

echo "[$(date)] Starting system check..."
echo "[$(date)] Initializing background monitoring..."

# Create a CPU-intensive process
# This simulates a runaway process
(
  while true; do
    # Burn CPU
    dd if=/dev/zero of=/dev/null bs=1M count=100 2>/dev/null
  done
) &

CPU_HOG_PID=$!
echo "[$(date)] Background monitor started (PID: $CPU_HOG_PID)"
echo ""
echo "System check running. Monitor CPU usage with: top or htop"
echo "Something seems wrong... CPU usage is very high!"
echo ""
echo "Your task: Find and stop the problematic process"
