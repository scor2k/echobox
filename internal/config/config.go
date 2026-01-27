package config

import (
	"crypto/rand"
	"encoding/binary"
	"fmt"
	"os"
	"strconv"
	"time"
)

// Config holds all application configuration
type Config struct {
	// Server
	Port int

	// Session
	CandidateName   string
	SessionTimeout  time.Duration
	ReconnectWindow time.Duration

	// Paths
	OutputDir string
	Shell     string

	// Recording
	FlushInterval time.Duration

	// Anti-cheat
	InputRateLimit int // chars per second

	// Security
	NetworkIsolated bool
	ShellUID        uint32 // Random UID for shell user (generated at startup)

	// Observability
	EnableMetrics bool
	LogLevel      string

	// Message of the day
	MOTD string
}

// Load reads configuration from environment variables with sensible defaults
func Load() (*Config, error) {
	// Generate random UID for shell user (10000-60000 range)
	// This UID is permanent for the container lifetime
	// Provides isolation: different containers = different UIDs
	shellUID := generateShellUID()

	cfg := &Config{
		Port:            getEnvInt("PORT", 8080),
		CandidateName:   getEnv("CANDIDATE_NAME", "anonymous"),
		SessionTimeout:  time.Duration(getEnvInt("SESSION_TIMEOUT", 7200)) * time.Second,
		ReconnectWindow: time.Duration(getEnvInt("RECONNECT_WINDOW", 300)) * time.Second,
		OutputDir:       getEnv("OUTPUT_DIR", "./sessions"),
		Shell:           getEnv("SHELL", "/bin/bash"),
		FlushInterval:   time.Duration(getEnvInt("FLUSH_INTERVAL", 10)) * time.Second,
		InputRateLimit:  getEnvInt("INPUT_RATE_LIMIT", 30),
		NetworkIsolated: getEnvBool("NETWORK_ISOLATED", true),
		ShellUID:        shellUID,
		EnableMetrics:   getEnvBool("ENABLE_METRICS", true),
		LogLevel:        getEnv("LOG_LEVEL", "info"),
		MOTD:            getEnv("MOTD", defaultMOTD()),
	}

	// Validation
	if cfg.Port < 1 || cfg.Port > 65535 {
		return nil, fmt.Errorf("PORT must be between 1 and 65535, got %d", cfg.Port)
	}

	if cfg.CandidateName == "" {
		return nil, fmt.Errorf("CANDIDATE_NAME cannot be empty")
	}

	if cfg.SessionTimeout < time.Minute {
		return nil, fmt.Errorf("SESSION_TIMEOUT must be at least 60 seconds")
	}

	if cfg.OutputDir == "" {
		return nil, fmt.Errorf("OUTPUT_DIR cannot be empty")
	}

	return cfg, nil
}

// Helper functions
func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getEnvInt(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if intVal, err := strconv.Atoi(value); err == nil {
			return intVal
		}
	}
	return defaultValue
}

func getEnvBool(key string, defaultValue bool) bool {
	if value := os.Getenv(key); value != "" {
		if boolVal, err := strconv.ParseBool(value); err == nil {
			return boolVal
		}
	}
	return defaultValue
}

func defaultMOTD() string {
	return `
╔══════════════════════════════════════════════════════════════╗
║                  SRE TECHNICAL INTERVIEW                     ║
╚══════════════════════════════════════════════════════════════╝

Welcome! You have been connected to an isolated interview environment.

INSTRUCTIONS:
• Complete the tasks in /tasks/ directory
• Read /tasks/README.md for detailed instructions
• Save your solutions in ~/solutions/
• Your session is being recorded for evaluation
• Use the "Finish" button when you're done

NOTES:
• Copy-paste is disabled for assessment integrity
• If you lose connection, refresh to reconnect
• All commands and keystrokes are logged

Good luck! 🚀
`
}

// generateShellUID generates a random UID for shell isolation
// Range: 10000-60000 (avoids system UIDs and gives 50k unique values)
// Each container gets a unique UID, preventing cross-session tampering
func generateShellUID() uint32 {
	var b [4]byte
	if _, err := rand.Read(b[:]); err != nil {
		// Fallback to timestamp-based if crypto/rand fails
		return uint32(10000 + (time.Now().UnixNano() % 50000))
	}

	// Convert to uint32 and constrain to range [10000, 60000]
	randomValue := binary.BigEndian.Uint32(b[:])
	return 10000 + (randomValue % 50000)
}
