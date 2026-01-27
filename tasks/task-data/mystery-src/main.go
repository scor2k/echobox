package main

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func main() {
	// Always print welcome message
	fmt.Println("Mystery Application v1.0")
	fmt.Println("Initializing...")

	// Add some noise for strace - check various paths
	// This makes investigation more challenging
	_, _ = os.Stat("/etc/hosts")
	_, _ = os.Stat("/proc/version")
	_, _ = os.Stat("/etc/resolv.conf")
	time.Sleep(50 * time.Millisecond)

	// Check for lock file in current directory - required to start
	lockFile := ".mystery.lock"

	if _, err := os.Stat(lockFile); os.IsNotExist(err) {
		// Print locked status but don't explain what's missing
		fmt.Println("Status: locked")

		// Add more noise before exit
		_, _ = os.ReadFile("/etc/os-release")
		_ = os.Getenv("HOME")
		_ = os.Getenv("USER")
		_, _ = os.Getwd()
		time.Sleep(30 * time.Millisecond)

		// Exit without explanation
		os.Exit(1)
	}

	// Success - lock file exists
	// Generate verification code from obfuscated pieces (harder to find with strings)
	// Code is split across multiple variables and concatenated
	part1 := string([]byte{83, 82, 69, 45})        // "SRE-"
	part2 := string([]byte{68, 69, 84, 69, 67})    // "DETEC"
	part3 := string([]byte{84, 73, 86, 69, 45})    // "TIVE-"
	suffix := fmt.Sprintf("%d%d%d", 4, 2, 7)       // "427"

	verificationCode := part1 + part2 + part3 + suffix

	fmt.Println("Status: Running successfully")
	fmt.Println("Lock file found: .mystery.lock")
	fmt.Println("Initialization complete")
	fmt.Println("")
	fmt.Println("═══════════════════════════════════════════════")
	fmt.Println("  VERIFICATION CODE: " + verificationCode)
	fmt.Println("═══════════════════════════════════════════════")
	fmt.Println("")
	fmt.Println("Copy this code to your solution file!")
	fmt.Println("")
	fmt.Println("Service running. Press Ctrl+C to stop.")

	// Set up signal handling to avoid deadlock
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	// Wait for signal instead of infinite select
	<-sigChan
	fmt.Println("\nShutting down...")
}
