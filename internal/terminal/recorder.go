package terminal

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"sync"
	"time"
)

// Recorder records terminal session to multiple log files
type Recorder struct {
	sessionDir string
	startTime  time.Time

	// File handles
	keystrokesFile *os.File
	websocketFile  *os.File
	eventsFile     *os.File

	// Buffered writers
	keystrokesWriter *bufio.Writer
	websocketWriter  *bufio.Writer
	eventsWriter     *bufio.Writer

	// Flush control
	flushTicker *time.Ticker
	flushDone   chan struct{}

	mu     sync.Mutex
	closed bool
}

// NewRecorder creates a new session recorder
func NewRecorder(sessionDir string, flushInterval time.Duration) (*Recorder, error) {
	r := &Recorder{
		sessionDir: sessionDir,
		startTime:  time.Now(),
		flushDone:  make(chan struct{}),
	}

	// Open all log files with write-only permissions during recording
	// Files will be made read-only on Close() to prevent tampering
	var err error

	r.keystrokesFile, err = os.OpenFile(
		fmt.Sprintf("%s/keystrokes.log", sessionDir),
		os.O_CREATE|os.O_WRONLY|os.O_TRUNC,
		0600, // Owner read/write during recording
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create keystrokes.log: %w", err)
	}

	r.websocketFile, err = os.OpenFile(
		fmt.Sprintf("%s/websocket.log", sessionDir),
		os.O_CREATE|os.O_WRONLY|os.O_TRUNC,
		0600,
	)
	if err != nil {
		r.keystrokesFile.Close()
		return nil, fmt.Errorf("failed to create websocket.log: %w", err)
	}

	r.eventsFile, err = os.OpenFile(
		fmt.Sprintf("%s/events.log", sessionDir),
		os.O_CREATE|os.O_WRONLY|os.O_TRUNC,
		0600,
	)
	if err != nil {
		r.keystrokesFile.Close()
		r.websocketFile.Close()
		return nil, fmt.Errorf("failed to create events.log: %w", err)
	}

	// Create buffered writers
	r.keystrokesWriter = bufio.NewWriter(r.keystrokesFile)
	r.websocketWriter = bufio.NewWriter(r.websocketFile)
	r.eventsWriter = bufio.NewWriter(r.eventsFile)

	// Start periodic flush
	r.flushTicker = time.NewTicker(flushInterval)
	go r.flushLoop()

	log.Printf("Recorder: Started session recording in %s", sessionDir)
	return r, nil
}

// RecordInput records keystroke input with timestamp
func (r *Recorder) RecordInput(data []byte) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if r.closed {
		return fmt.Errorf("recorder is closed")
	}

	timestamp := time.Since(r.startTime).Milliseconds()

	// Format: timestamp_ms data_as_hex_or_printable
	_, err := fmt.Fprintf(r.keystrokesWriter, "%d %q\n", timestamp, string(data))
	return err
}

// RecordWebSocketMessage records WebSocket message
func (r *Recorder) RecordWebSocketMessage(direction string, messageType string, data []byte) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if r.closed {
		return fmt.Errorf("recorder is closed")
	}

	timestamp := time.Since(r.startTime).Milliseconds()

	// Format: timestamp direction type length [data_sample]
	dataSample := string(data)
	if len(dataSample) > 100 {
		dataSample = dataSample[:100] + "..."
	}

	_, err := fmt.Fprintf(r.websocketWriter, "%d %s %s %d %q\n",
		timestamp, direction, messageType, len(data), dataSample)
	return err
}

// RecordEvent records anti-cheat or session events
func (r *Recorder) RecordEvent(eventType string, data string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if r.closed {
		return fmt.Errorf("recorder is closed")
	}

	timestamp := time.Since(r.startTime).Milliseconds()

	// Format: timestamp event_type data
	_, err := fmt.Fprintf(r.eventsWriter, "%d %s %s\n", timestamp, eventType, data)
	return err
}

// flushLoop periodically flushes all buffers
func (r *Recorder) flushLoop() {
	for {
		select {
		case <-r.flushTicker.C:
			r.Flush()
		case <-r.flushDone:
			return
		}
	}
}

// Flush flushes all buffered writers
func (r *Recorder) Flush() error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if r.closed {
		return nil
	}

	var errs []error

	if err := r.keystrokesWriter.Flush(); err != nil {
		errs = append(errs, fmt.Errorf("keystrokes: %w", err))
	}
	if err := r.websocketWriter.Flush(); err != nil {
		errs = append(errs, fmt.Errorf("websocket: %w", err))
	}
	if err := r.eventsWriter.Flush(); err != nil {
		errs = append(errs, fmt.Errorf("events: %w", err))
	}

	// Sync to disk
	r.keystrokesFile.Sync()
	r.websocketFile.Sync()
	r.eventsFile.Sync()

	if len(errs) > 0 {
		return fmt.Errorf("flush errors: %v", errs)
	}

	return nil
}

// Close closes the recorder and all files
func (r *Recorder) Close() error {
	r.mu.Lock()
	if r.closed {
		r.mu.Unlock()
		return nil
	}
	r.closed = true
	r.mu.Unlock()

	log.Println("Recorder: Closing and flushing all logs...")

	// Stop flush loop
	r.flushTicker.Stop()
	close(r.flushDone)

	// Final flush
	r.Flush()

	// Close all files
	r.keystrokesFile.Close()
	r.websocketFile.Close()
	r.eventsFile.Close()

	// Make log files read-only to prevent tampering
	// After session ends, files become 0400 (owner read-only)
	logFiles := []string{
		fmt.Sprintf("%s/keystrokes.log", r.sessionDir),
		fmt.Sprintf("%s/websocket.log", r.sessionDir),
		fmt.Sprintf("%s/events.log", r.sessionDir),
	}

	for _, logFile := range logFiles {
		if err := os.Chmod(logFile, 0400); err != nil {
			log.Printf("Warning: Could not make %s read-only: %v", logFile, err)
		}
	}

	log.Println("Recorder: All logs closed and protected (read-only)")
	return nil
}

// ExtractCommands extracts shell commands from keystrokes.log
func ExtractCommands(sessionDir string) error {
	keystrokesLogPath := fmt.Sprintf("%s/keystrokes.log", sessionDir)
	commandsLogPath := fmt.Sprintf("%s/commands.log", sessionDir)

	keystrokesFile, err := os.Open(keystrokesLogPath)
	if err != nil {
		return fmt.Errorf("failed to read keystrokes.log: %w", err)
	}
	defer keystrokesFile.Close()

	commandsFile, err := os.OpenFile(commandsLogPath, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0600)
	if err != nil {
		return fmt.Errorf("failed to create commands.log: %w", err)
	}
	defer commandsFile.Close()

	commandsFile.WriteString("# Commands extracted from keystrokes.log\n\n")

	// Parse keystrokes and extract commands
	// Format: timestamp_ms "keystroke"
	var currentCommand []byte
	scanner := bufio.NewScanner(keystrokesFile)

	for scanner.Scan() {
		line := scanner.Text()

		// Parse the quoted keystroke data
		// Format: 1234 "keystroke data"
		var timestamp int64
		var keystroke string
		_, err := fmt.Sscanf(line, "%d %q", &timestamp, &keystroke)
		if err != nil {
			continue // Skip malformed lines
		}

		for _, ch := range []byte(keystroke) {
			switch ch {
			case '\r', '\n': // Enter pressed - command complete
				if len(currentCommand) > 0 {
					fmt.Fprintf(commandsFile, "%d %s\n", timestamp, string(currentCommand))
					currentCommand = nil
				}
			case 127, '\b': // Backspace (DEL or BS)
				if len(currentCommand) > 0 {
					currentCommand = currentCommand[:len(currentCommand)-1]
				}
			case 3: // Ctrl+C
				currentCommand = nil // Cancel current command
			case 21: // Ctrl+U
				currentCommand = nil // Clear line
			default:
				if ch >= 32 && ch < 127 { // Printable ASCII
					currentCommand = append(currentCommand, ch)
				}
			}
		}
	}

	// Handle any remaining command (session ended without Enter)
	if len(currentCommand) > 0 {
		fmt.Fprintf(commandsFile, "# Incomplete: %s\n", string(currentCommand))
	}

	// Make file read-only immediately after writing
	commandsFile.Close()
	if err := os.Chmod(commandsLogPath, 0400); err != nil {
		log.Printf("Warning: Could not protect commands.log: %v", err)
	}

	return nil
}
