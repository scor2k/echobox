package main

import (
	"fmt"
	"os"
)

func main() {
	// Check for lock file in current directory - required to start
	lockFile := ".mystery.lock"

	if _, err := os.Stat(lockFile); os.IsNotExist(err) {
		// Exit silently without explanation - this is the mystery!
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

	fmt.Println("Mystery Application v1.0")
	fmt.Println("Status: Running successfully")
	fmt.Println("Lock file found: .mystery.lock")
	fmt.Println("Initialization complete")
	fmt.Println("")
	fmt.Println("═══════════════════════════════════════════════")
	fmt.Println("  VERIFICATION CODE: " + verificationCode)
	fmt.Println("═══════════════════════════════════════════════")
	fmt.Println("")
	fmt.Println("Copy this code to your solution file!")
	fmt.Println("Press Ctrl+C to stop")

	// Simulate a running service
	select {}
}
