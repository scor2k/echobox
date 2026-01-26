package web

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"sync"

	"github.com/gorilla/websocket"

	"github.com/akonyukov/echobox/internal/terminal"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true // Allow all origins for now
	},
}

// Message represents a WebSocket message
type Message struct {
	Type string          `json:"type"`
	Data json.RawMessage `json:"data"`
}

// ResizeData represents terminal resize data
type ResizeData struct {
	Cols uint16 `json:"cols"`
	Rows uint16 `json:"rows"`
}

// WSHandler handles WebSocket connections
type WSHandler struct {
	pty          *terminal.PTY
	mu           sync.RWMutex
	shutdown     chan struct{}
	finishSignal chan struct{}
}

// NewWSHandler creates a new WebSocket handler
func NewWSHandler(pty *terminal.PTY) *WSHandler {
	return &WSHandler{
		pty:          pty,
		shutdown:     make(chan struct{}),
		finishSignal: make(chan struct{}, 1),
	}
}

// Shutdown signals all connections to close
func (h *WSHandler) Shutdown() {
	select {
	case <-h.shutdown:
		// Already closed
	default:
		close(h.shutdown)
	}
}

// FinishSignal returns a channel that signals when finish is requested
func (h *WSHandler) FinishSignal() <-chan struct{} {
	return h.finishSignal
}

// Handle handles WebSocket upgrade and communication
func (h *WSHandler) Handle(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("Failed to upgrade connection: %v", err)
		return
	}
	defer conn.Close()

	log.Printf("WebSocket connected: %s", r.RemoteAddr)

	// Write mutex to prevent concurrent writes to WebSocket
	var writeMu sync.Mutex

	// Create channels for coordination
	done := make(chan struct{})

	// PTY -> WebSocket (read from PTY, write to WebSocket)
	go func() {
		buf := make([]byte, 8192)
		for {
			n, err := h.pty.Read(buf)
			if err != nil {
				if err != io.EOF {
					log.Printf("PTY read error: %v", err)
				}
				select {
				case <-done:
					// Already closed
				default:
					close(done)
				}
				return
			}

			if n > 0 {
				writeMu.Lock()
				err := conn.WriteMessage(websocket.TextMessage, buf[:n])
				writeMu.Unlock()

				if err != nil {
					log.Printf("WebSocket write error: %v", err)
					select {
					case <-done:
						// Already closed
					default:
						close(done)
					}
					return
				}
			}
		}
	}()

	// WebSocket -> PTY (read from WebSocket, write to PTY)
	go func() {
		for {
			messageType, data, err := conn.ReadMessage()
			if err != nil {
				if !websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
					log.Printf("WebSocket closed normally: %s", r.RemoteAddr)
				} else {
					log.Printf("WebSocket read error: %v", err)
				}
				select {
				case <-done:
					// Already closed
				default:
					close(done)
				}
				return
			}

			switch messageType {
			case websocket.TextMessage:
				// Check if it's a JSON message (for resize, finish, etc.)
				var msg Message
				if err := json.Unmarshal(data, &msg); err == nil && msg.Type != "" {
					// Handle structured messages
					if err := h.handleMessage(&msg); err != nil {
						log.Printf("Error handling message: %v", err)
					}
				} else {
					// Plain text input - write to PTY
					if _, err := h.pty.Write(data); err != nil {
						log.Printf("PTY write error: %v", err)
						select {
						case <-done:
							// Already closed
						default:
							close(done)
						}
						return
					}
				}

			case websocket.BinaryMessage:
				// Write binary data directly to PTY
				if _, err := h.pty.Write(data); err != nil {
					log.Printf("PTY write error: %v", err)
					select {
					case <-done:
						// Already closed
					default:
						close(done)
					}
					return
				}
			}
		}
	}()

	// Wait for completion or shutdown
	select {
	case <-done:
		log.Printf("WebSocket disconnected: %s", r.RemoteAddr)
	case <-h.shutdown:
		log.Printf("Shutdown signal received, closing connection: %s", r.RemoteAddr)
	}
}

// handleMessage handles structured WebSocket messages
func (h *WSHandler) handleMessage(msg *Message) error {
	switch msg.Type {
	case "resize":
		var resize ResizeData
		if err := json.Unmarshal(msg.Data, &resize); err != nil {
			return fmt.Errorf("invalid resize data: %w", err)
		}
		return h.pty.Resize(resize.Cols, resize.Rows)

	case "finish":
		log.Println("Session finish requested by client")
		// Signal finish to main application
		select {
		case h.finishSignal <- struct{}{}:
		default:
			// Already signaled
		}
		return nil

	case "anticheat":
		// Log anti-cheat event (will be properly logged in Phase 2)
		log.Printf("Anti-cheat event: %s", string(msg.Data))
		return nil

	default:
		log.Printf("Unknown message type: %s", msg.Type)
	}

	return nil
}
