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
	fmt.Println("Mystery Application v1.0")
	fmt.Println("Status: Running successfully")
	fmt.Println("Lock file found: .mystery.lock")
	fmt.Println("Service started on port 9999")
	fmt.Println("")
	fmt.Println("Press Ctrl+C to stop")

	// Simulate a running service
	select {}
}
