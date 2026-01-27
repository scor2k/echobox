package terminal

import (
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"sync"
	"syscall"
	"unsafe"

	"github.com/creack/pty"
)

// PTY represents a pseudo-terminal
type PTY struct {
	file        *os.File
	cmd         *exec.Cmd
	mu          sync.Mutex
	closed      bool
	readers     []io.Reader
	writers     []io.Writer
	candidateHome string
}

// New creates a new PTY and spawns the specified shell as the given UID
// The shell runs as shellUID (random, isolated), while logs stay owned by root
func New(shell string, shellUID uint32) (*PTY, error) {
	// Create home directory for this UID if it doesn't exist
	homeDir := fmt.Sprintf("/home/candidate-%d", shellUID)
	if err := os.MkdirAll(homeDir+"/solutions", 0755); err != nil {
		log.Printf("Warning: Could not create home directory: %v", err)
		homeDir = "/tmp" // Fallback
	} else {
		// Set ownership to the shell UID
		os.Chown(homeDir, int(shellUID), int(shellUID))
		os.Chown(homeDir+"/solutions", int(shellUID), int(shellUID))

		// Setup candidate environment (copy tasks, create directories)
		setupScript := "/tasks/setup-candidate-env.sh"
		if _, err := os.Stat(setupScript); err == nil {
			setupCmd := exec.Command("/bin/bash", setupScript, homeDir)
			if err := setupCmd.Run(); err != nil {
				log.Printf("Warning: Could not setup candidate environment: %v", err)
			} else {
				log.Printf("PTY: Candidate environment setup complete")
			}
		}
	}

	// Create command
	cmd := exec.Command(shell)

	// Set up environment for the shell user
	cmd.Env = []string{
		"TERM=xterm-256color",
		"COLORTERM=truecolor",
		fmt.Sprintf("HOME=%s", homeDir),
		fmt.Sprintf("USER=candidate-%d", shellUID),
		"PATH=/usr/local/bin:/usr/bin:/bin:/usr/sbin:/sbin",
	}

	// Set working directory
	cmd.Dir = homeDir

	// Run shell as random UID for isolation (only if running as root)
	// This prevents candidates from tampering with each other's sessions
	if os.Getuid() == 0 {
		cmd.SysProcAttr = &syscall.SysProcAttr{
			Credential: &syscall.Credential{
				Uid: shellUID,
				Gid: shellUID,
			},
		}
		log.Printf("PTY: Starting shell as UID %d (home: %s)", shellUID, homeDir)
	} else {
		log.Printf("PTY: Starting shell as current user (not root, cannot setuid)")
	}

	// Start the command with a PTY
	ptmx, err := pty.Start(cmd)
	if err != nil {
		return nil, fmt.Errorf("failed to start PTY: %w", err)
	}

	p := &PTY{
		file:          ptmx,
		cmd:           cmd,
		readers:       make([]io.Reader, 0),
		writers:       make([]io.Writer, 0),
		candidateHome: homeDir,
	}

	return p, nil
}

// GetCandidateHome returns the candidate's home directory path
func (p *PTY) GetCandidateHome() string {
	return p.candidateHome
}

// Read reads data from the PTY
func (p *PTY) Read(buf []byte) (int, error) {
	// Check closed state without holding lock during I/O
	p.mu.Lock()
	if p.closed {
		p.mu.Unlock()
		return 0, io.EOF
	}
	file := p.file
	p.mu.Unlock()

	// Do blocking I/O without holding lock
	n, err := file.Read(buf)

	// Broadcast to all registered readers
	if n > 0 {
		p.mu.Lock()
		readers := p.readers
		p.mu.Unlock()

		for _, r := range readers {
			if w, ok := r.(io.Writer); ok {
				w.Write(buf[:n])
			}
		}
	}

	return n, err
}

// Write writes data to the PTY
func (p *PTY) Write(data []byte) (int, error) {
	// Check closed state without holding lock during I/O
	p.mu.Lock()
	if p.closed {
		p.mu.Unlock()
		return 0, io.EOF
	}
	file := p.file
	p.mu.Unlock()

	// Do blocking I/O without holding lock
	n, err := file.Write(data)

	// Broadcast to all registered writers
	if n > 0 {
		p.mu.Lock()
		writers := p.writers
		p.mu.Unlock()

		for _, w := range writers {
			w.Write(data[:n])
		}
	}

	return n, err
}

// AddReader adds a reader that will receive copies of data read from the PTY
func (p *PTY) AddReader(r io.Reader) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.readers = append(p.readers, r)
}

// AddWriter adds a writer that will receive copies of data written to the PTY
func (p *PTY) AddWriter(w io.Writer) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.writers = append(p.writers, w)
}

// Resize resizes the PTY to the specified dimensions
func (p *PTY) Resize(cols, rows uint16) error {
	p.mu.Lock()
	defer p.mu.Unlock()

	if p.closed {
		return fmt.Errorf("PTY is closed")
	}

	size := &struct {
		rows uint16
		cols uint16
		x    uint16
		y    uint16
	}{
		rows: rows,
		cols: cols,
	}

	_, _, errno := syscall.Syscall(
		syscall.SYS_IOCTL,
		p.file.Fd(),
		syscall.TIOCSWINSZ,
		uintptr(unsafe.Pointer(size)),
	)

	if errno != 0 {
		return fmt.Errorf("failed to resize PTY: %v", errno)
	}

	return nil
}

// Close closes the PTY and terminates the shell process
func (p *PTY) Close() error {
	p.mu.Lock()
	if p.closed {
		p.mu.Unlock()
		return nil
	}
	p.closed = true
	p.mu.Unlock()

	log.Println("PTY: Closing file descriptor")

	// Close the PTY file first
	if err := p.file.Close(); err != nil {
		log.Printf("PTY: Error closing file: %v", err)
	}

	// Kill the process immediately (don't wait)
	if p.cmd.Process != nil {
		log.Printf("PTY: Terminating shell process (PID: %d)", p.cmd.Process.Pid)
		if err := p.cmd.Process.Kill(); err != nil {
			log.Printf("PTY: Error killing process: %v", err)
		}

		// Wait in background (non-blocking)
		go func() {
			if err := p.cmd.Wait(); err != nil {
				log.Printf("PTY: Process exited: %v", err)
			}
		}()
	}

	log.Println("PTY: Close complete")
	return nil
}

// File returns the underlying file descriptor
func (p *PTY) File() *os.File {
	return p.file
}

// Wait waits for the command to exit
func (p *PTY) Wait() error {
	if p.cmd == nil {
		return nil
	}
	return p.cmd.Wait()
}
