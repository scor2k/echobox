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
	terminalFile   *os.File
	timingFile     *os.File
	websocketFile  *os.File
	eventsFile     *os.File

	// Buffered writers
	keystrokesWriter *bufio.Writer
	terminalWriter   *bufio.Writer
	timingWriter     *bufio.Writer
	websocketWriter  *bufio.Writer
	eventsWriter     *bufio.Writer

	// Timing tracking
	lastOutputTime time.Time

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

	r.terminalFile, err = os.OpenFile(
		fmt.Sprintf("%s/terminal.log", sessionDir),
		os.O_CREATE|os.O_WRONLY|os.O_TRUNC,
		0600,
	)
	if err != nil {
		r.keystrokesFile.Close()
		return nil, fmt.Errorf("failed to create terminal.log: %w", err)
	}

	r.timingFile, err = os.OpenFile(
		fmt.Sprintf("%s/timing.log", sessionDir),
		os.O_CREATE|os.O_WRONLY|os.O_TRUNC,
		0600,
	)
	if err != nil {
		r.keystrokesFile.Close()
		r.terminalFile.Close()
		return nil, fmt.Errorf("failed to create timing.log: %w", err)
	}

	r.websocketFile, err = os.OpenFile(
		fmt.Sprintf("%s/websocket.log", sessionDir),
		os.O_CREATE|os.O_WRONLY|os.O_TRUNC,
		0600,
	)
	if err != nil {
		r.keystrokesFile.Close()
		r.terminalFile.Close()
		r.timingFile.Close()
		return nil, fmt.Errorf("failed to create websocket.log: %w", err)
	}

	r.eventsFile, err = os.OpenFile(
		fmt.Sprintf("%s/events.log", sessionDir),
		os.O_CREATE|os.O_WRONLY|os.O_TRUNC,
		0600,
	)
	if err != nil {
		r.keystrokesFile.Close()
		r.terminalFile.Close()
		r.timingFile.Close()
		r.websocketFile.Close()
		return nil, fmt.Errorf("failed to create events.log: %w", err)
	}

	// Create buffered writers
	r.keystrokesWriter = bufio.NewWriter(r.keystrokesFile)
	r.terminalWriter = bufio.NewWriter(r.terminalFile)
	r.timingWriter = bufio.NewWriter(r.timingFile)
	r.websocketWriter = bufio.NewWriter(r.websocketFile)
	r.eventsWriter = bufio.NewWriter(r.eventsFile)

	r.lastOutputTime = r.startTime

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

// RecordOutput records terminal output in script/scriptreplay format
func (r *Recorder) RecordOutput(data []byte) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if r.closed {
		return fmt.Errorf("recorder is closed")
	}

	now := time.Now()
	elapsed := now.Sub(r.lastOutputTime).Seconds()
	r.lastOutputTime = now

	// Write timing info (for scriptreplay)
	// Format: seconds data_length
	_, err := fmt.Fprintf(r.timingWriter, "%.6f %d\n", elapsed, len(data))
	if err != nil {
		return err
	}

	// Write terminal output
	_, err = r.terminalWriter.Write(data)
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
	if err := r.terminalWriter.Flush(); err != nil {
		errs = append(errs, fmt.Errorf("terminal: %w", err))
	}
	if err := r.timingWriter.Flush(); err != nil {
		errs = append(errs, fmt.Errorf("timing: %w", err))
	}
	if err := r.websocketWriter.Flush(); err != nil {
		errs = append(errs, fmt.Errorf("websocket: %w", err))
	}
	if err := r.eventsWriter.Flush(); err != nil {
		errs = append(errs, fmt.Errorf("events: %w", err))
	}

	// Sync to disk
	r.keystrokesFile.Sync()
	r.terminalFile.Sync()
	r.timingFile.Sync()
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
	r.terminalFile.Close()
	r.timingFile.Close()
	r.websocketFile.Close()
	r.eventsFile.Close()

	// Make log files read-only to prevent tampering
	// After session ends, files become 0400 (owner read-only)
	logFiles := []string{
		fmt.Sprintf("%s/keystrokes.log", r.sessionDir),
		fmt.Sprintf("%s/terminal.log", r.sessionDir),
		fmt.Sprintf("%s/timing.log", r.sessionDir),
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

// ExtractCommands extracts shell commands from terminal log (basic implementation)
func ExtractCommands(sessionDir string) error {
	// This is a simple implementation - can be enhanced later
	terminalLogPath := fmt.Sprintf("%s/terminal.log", sessionDir)
	commandsLogPath := fmt.Sprintf("%s/commands.log", sessionDir)

	terminalData, err := os.ReadFile(terminalLogPath)
	if err != nil {
		return fmt.Errorf("failed to read terminal.log: %w", err)
	}

	commandsFile, err := os.Create(commandsLogPath)
	if err != nil {
		return fmt.Errorf("failed to create commands.log: %w", err)
	}
	defer commandsFile.Close()

	// Simple extraction: look for common shell prompts and extract what follows
	// This is a placeholder - real implementation would parse ANSI codes properly
	_, err = commandsFile.Write([]byte(fmt.Sprintf("# Commands extracted from session\n# Terminal log size: %d bytes\n\n", len(terminalData))))
	if err != nil {
		return err
	}

	// TODO: Implement proper command extraction with ANSI parsing
	commandsFile.WriteString("# Command extraction not yet implemented\n")
	commandsFile.WriteString("# Use 'scriptreplay' to view the full session\n")

	return nil
}
