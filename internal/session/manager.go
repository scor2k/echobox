package session

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"time"

	"github.com/google/uuid"
)

// Session represents an interview session
type Session struct {
	ID           string    `json:"id"`
	CandidateName string   `json:"candidate_name"`
	StartTime    time.Time `json:"start_time"`
	EndTime      time.Time `json:"end_time,omitempty"`
	Duration     float64   `json:"duration_seconds,omitempty"`
	OutputDir    string    `json:"output_dir"`
	Status       string    `json:"status"` // active, completed, error
	FileHashes   map[string]string `json:"file_hashes,omitempty"`
	Metadata     map[string]interface{} `json:"metadata,omitempty"`
}

// Manager manages session lifecycle and recording
type Manager struct {
	session    *Session
	state      *SessionState
	baseDir    string
	sessionDir string
}

// NewManager creates a new session manager
func NewManager(baseDir, candidateName string, reconnectWindow time.Duration) (*Manager, error) {
	// Create base output directory if it doesn't exist
	if err := os.MkdirAll(baseDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create base directory: %w", err)
	}

	// Generate session ID and directory name
	sessionID := uuid.New().String()[:8] // Short UUID for readability
	timestamp := time.Now().Format("2006-01-02_15-04-05")
	sessionDirName := fmt.Sprintf("%s_%s_%s", candidateName, timestamp, sessionID)
	sessionDir := filepath.Join(baseDir, sessionDirName)

	// Create session directory with restricted permissions (root only)
	// Mode 0700 = only owner (root) can read/write/execute
	// Prevents candidate from reading logs
	if err := os.MkdirAll(sessionDir, 0700); err != nil {
		return nil, fmt.Errorf("failed to create session directory: %w", err)
	}

	session := &Session{
		ID:            sessionID,
		CandidateName: candidateName,
		StartTime:     time.Now(),
		OutputDir:     sessionDir,
		Status:        "active",
		Metadata:      make(map[string]interface{}),
	}

	// Create session state for reconnection support
	state := NewSessionState(reconnectWindow)

	m := &Manager{
		session:    session,
		state:      state,
		baseDir:    baseDir,
		sessionDir: sessionDir,
	}

	// Store reconnect token in metadata
	m.SetMetadata("reconnect_token", state.GetReconnectToken())
	m.SetMetadata("reconnect_window_seconds", reconnectWindow.Seconds())

	// Write initial metadata
	if err := m.SaveMetadata(); err != nil {
		return nil, fmt.Errorf("failed to save initial metadata: %w", err)
	}

	return m, nil
}

// GetState returns the session state
func (m *Manager) GetState() *SessionState {
	return m.state
}

// GetSession returns the current session
func (m *Manager) GetSession() *Session {
	return m.session
}

// GetSessionDir returns the session directory path
func (m *Manager) GetSessionDir() string {
	return m.sessionDir
}

// GetFilePath returns the full path for a session file
func (m *Manager) GetFilePath(filename string) string {
	return filepath.Join(m.sessionDir, filename)
}

// SetMetadata sets a metadata key-value pair
func (m *Manager) SetMetadata(key string, value interface{}) {
	m.session.Metadata[key] = value
}

// SaveMetadata writes the current session metadata to metadata.json
func (m *Manager) SaveMetadata() error {
	m.session.EndTime = time.Now()
	m.session.Duration = m.session.EndTime.Sub(m.session.StartTime).Seconds()

	metadataPath := m.GetFilePath("metadata.json")
	data, err := json.MarshalIndent(m.session, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal metadata: %w", err)
	}

	// Write with restricted permissions (root only)
	if err := os.WriteFile(metadataPath, data, 0600); err != nil {
		return fmt.Errorf("failed to write metadata: %w", err)
	}

	return nil
}

// Complete marks the session as completed and calculates file hashes
func (m *Manager) Complete() error {
	m.session.Status = "completed"
	m.session.EndTime = time.Now()
	m.session.Duration = m.session.EndTime.Sub(m.session.StartTime).Seconds()

	// Calculate hashes for all recorded files
	if err := m.calculateFileHashes(); err != nil {
		return fmt.Errorf("failed to calculate file hashes: %w", err)
	}

	// Save final metadata
	if err := m.SaveMetadata(); err != nil {
		return fmt.Errorf("failed to save final metadata: %w", err)
	}

	// Protect all session files from tampering
	// Make metadata, analysis, and commands read-only
	protectedFiles := []string{
		"metadata.json",
		"analysis.json",
		"commands.log",
	}

	for _, filename := range protectedFiles {
		filePath := m.GetFilePath(filename)
		if _, err := os.Stat(filePath); err == nil {
			// File exists, make it read-only (owner only)
			if err := os.Chmod(filePath, 0400); err != nil {
				log.Printf("Warning: Could not protect %s: %v", filename, err)
			}
		}
	}

	log.Println("Session: All files protected (read-only, root access only)")
	return nil
}

// Error marks the session as errored
func (m *Manager) Error(err error) error {
	m.session.Status = "error"
	m.session.Metadata["error"] = err.Error()
	return m.SaveMetadata()
}

// calculateFileHashes calculates SHA-256 hashes for all session files
func (m *Manager) calculateFileHashes() error {
	m.session.FileHashes = make(map[string]string)

	// List of files to hash
	files := []string{
		"keystrokes.log",
		"terminal.log",
		"timing.log",
		"websocket.log",
		"events.log",
		"commands.log",
	}

	for _, filename := range files {
		filePath := m.GetFilePath(filename)

		// Check if file exists
		if _, err := os.Stat(filePath); os.IsNotExist(err) {
			continue // Skip non-existent files
		}

		hash, err := hashFile(filePath)
		if err != nil {
			return fmt.Errorf("failed to hash %s: %w", filename, err)
		}

		m.session.FileHashes[filename] = hash
	}

	return nil
}

// hashFile calculates SHA-256 hash of a file
func hashFile(filePath string) (string, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return "", err
	}
	defer file.Close()

	hash := sha256.New()
	if _, err := io.Copy(hash, file); err != nil {
		return "", err
	}

	return hex.EncodeToString(hash.Sum(nil)), nil
}

// VerifyFileHash verifies a file's hash against recorded hash
func (m *Manager) VerifyFileHash(filename string) (bool, error) {
	expectedHash, exists := m.session.FileHashes[filename]
	if !exists {
		return false, fmt.Errorf("no hash recorded for %s", filename)
	}

	filePath := m.GetFilePath(filename)
	actualHash, err := hashFile(filePath)
	if err != nil {
		return false, err
	}

	return actualHash == expectedHash, nil
}
