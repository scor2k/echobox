package session

import (
	"sync"
	"time"

	"github.com/google/uuid"
)

// ConnectionState represents the state of a session connection
type ConnectionState string

const (
	StateActive       ConnectionState = "active"
	StateDisconnected ConnectionState = "disconnected"
	StateExpired      ConnectionState = "expired"
)

// SessionState tracks the runtime state of a session
type SessionState struct {
	// Reconnection
	ReconnectToken   string
	TokenCreatedAt   time.Time
	ReconnectWindow  time.Duration

	// Connection tracking
	State            ConnectionState
	LastConnectTime  time.Time
	LastDisconnectTime time.Time
	ConnectionCount  int
	DisconnectCount  int

	// Terminal state
	TerminalBuffer   []byte
	LastCursorPos    CursorPosition
	TerminalSize     TerminalSize

	mu sync.RWMutex
}

// CursorPosition represents terminal cursor position
type CursorPosition struct {
	Row int
	Col int
}

// TerminalSize represents terminal dimensions
type TerminalSize struct {
	Cols uint16
	Rows uint16
}

// NewSessionState creates a new session state
func NewSessionState(reconnectWindow time.Duration) *SessionState {
	return &SessionState{
		ReconnectToken:  uuid.New().String(),
		TokenCreatedAt:  time.Now(),
		ReconnectWindow: reconnectWindow,
		State:           StateActive,
		LastConnectTime: time.Now(),
		ConnectionCount: 1,
		TerminalBuffer:  make([]byte, 0),
	}
}

// Connect marks a new connection
func (s *SessionState) Connect() {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.State = StateActive
	s.LastConnectTime = time.Now()
	s.ConnectionCount++
}

// Disconnect marks a disconnection
func (s *SessionState) Disconnect() {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.State = StateDisconnected
	s.LastDisconnectTime = time.Now()
	s.DisconnectCount++
}

// CanReconnect checks if reconnection is allowed
func (s *SessionState) CanReconnect(token string) bool {
	s.mu.RLock()
	defer s.mu.RUnlock()

	// Check token match
	if s.ReconnectToken != token {
		return false
	}

	// Check if expired
	if s.IsExpired() {
		return false
	}

	// Check if session is in reconnectable state
	return s.State == StateDisconnected
}

// IsExpired checks if the reconnection window has expired
func (s *SessionState) IsExpired() bool {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if s.State == StateExpired {
		return true
	}

	// Check if disconnected for too long
	if s.State == StateDisconnected {
		elapsed := time.Since(s.LastDisconnectTime)
		if elapsed > s.ReconnectWindow {
			return true
		}
	}

	return false
}

// MarkExpired marks the session as expired
func (s *SessionState) MarkExpired() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.State = StateExpired
}

// GetState returns current connection state
func (s *SessionState) GetState() ConnectionState {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.State
}

// GetReconnectToken returns the reconnection token
func (s *SessionState) GetReconnectToken() string {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.ReconnectToken
}

// UpdateTerminalBuffer updates the stored terminal buffer
func (s *SessionState) UpdateTerminalBuffer(data []byte) {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Keep last 100KB of output for reconnection
	maxBufferSize := 100 * 1024
	s.TerminalBuffer = append(s.TerminalBuffer, data...)

	// Trim if too large (keep most recent)
	if len(s.TerminalBuffer) > maxBufferSize {
		s.TerminalBuffer = s.TerminalBuffer[len(s.TerminalBuffer)-maxBufferSize:]
	}
}

// GetTerminalBuffer returns a copy of the terminal buffer
func (s *SessionState) GetTerminalBuffer() []byte {
	s.mu.RLock()
	defer s.mu.RUnlock()

	buffer := make([]byte, len(s.TerminalBuffer))
	copy(buffer, s.TerminalBuffer)
	return buffer
}

// UpdateTerminalSize updates the terminal dimensions
func (s *SessionState) UpdateTerminalSize(cols, rows uint16) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.TerminalSize = TerminalSize{Cols: cols, Rows: rows}
}

// GetTerminalSize returns the current terminal size
func (s *SessionState) GetTerminalSize() TerminalSize {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.TerminalSize
}

// GetConnectionStats returns connection statistics
func (s *SessionState) GetConnectionStats() map[string]interface{} {
	s.mu.RLock()
	defer s.mu.RUnlock()

	return map[string]interface{}{
		"state":              s.State,
		"connection_count":   s.ConnectionCount,
		"disconnect_count":   s.DisconnectCount,
		"last_connect_time":  s.LastConnectTime,
		"last_disconnect_time": s.LastDisconnectTime,
		"token_age_seconds":  time.Since(s.TokenCreatedAt).Seconds(),
	}
}
