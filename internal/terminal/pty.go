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
	file    *os.File
	cmd     *exec.Cmd
	mu      sync.Mutex
	closed  bool
	readers []io.Reader
	writers []io.Writer
}

// New creates a new PTY and spawns the specified shell
func New(shell string) (*PTY, error) {
	// Create command
	cmd := exec.Command(shell)

	// Set up environment
	cmd.Env = append(os.Environ(),
		"TERM=xterm-256color",
		"COLORTERM=truecolor",
	)

	// Start the command with a PTY
	ptmx, err := pty.Start(cmd)
	if err != nil {
		return nil, fmt.Errorf("failed to start PTY: %w", err)
	}

	p := &PTY{
		file:    ptmx,
		cmd:     cmd,
		readers: make([]io.Reader, 0),
		writers: make([]io.Writer, 0),
	}

	return p, nil
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
