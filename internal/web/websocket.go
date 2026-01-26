package web

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/gorilla/websocket"

	"github.com/akonyukov/echobox/internal/anticheat"
	"github.com/akonyukov/echobox/internal/session"
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
	recorder     *terminal.Recorder
	detector     *anticheat.Detector
	sessionState *session.SessionState
	mu           sync.RWMutex
	shutdown     chan struct{}
	finishSignal chan struct{}
}

// NewWSHandler creates a new WebSocket handler
func NewWSHandler(pty *terminal.PTY, recorder *terminal.Recorder, detector *anticheat.Detector, sessionState *session.SessionState) *WSHandler {
	return &WSHandler{
		pty:          pty,
		recorder:     recorder,
		detector:     detector,
		sessionState: sessionState,
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

	// Mark connection in session state
	if h.sessionState != nil {
		h.sessionState.Connect()
		log.Printf("Session state: %s (token: %s)", h.sessionState.GetState(), h.sessionState.GetReconnectToken())
	}

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
				// PTY closed - likely shell exited
				log.Printf("PTY closed (shell exited): %v", err)

				// Send session_ended message to client
				endMsg := map[string]interface{}{
					"type": "session_ended",
					"data": map[string]string{
						"reason": "shell_exited",
					},
				}
				msgBytes, _ := json.Marshal(endMsg)
				writeMu.Lock()
				conn.WriteMessage(websocket.TextMessage, msgBytes)
				writeMu.Unlock()

				// Give client time to receive message
				time.Sleep(500 * time.Millisecond)

				// Trigger session finish
				select {
				case h.finishSignal <- struct{}{}:
					log.Println("Triggered session finish after shell exit")
				default:
					// Already signaled
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
				// Record terminal output
				if h.recorder != nil {
					if err := h.recorder.RecordOutput(buf[:n]); err != nil {
						log.Printf("Failed to record output: %v", err)
					}
				}

				// Update terminal buffer for reconnection
				if h.sessionState != nil {
					h.sessionState.UpdateTerminalBuffer(buf[:n])
				}

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

			// Record WebSocket message
			if h.recorder != nil {
				msgTypeStr := "text"
				if messageType == websocket.BinaryMessage {
					msgTypeStr = "binary"
				}
				if err := h.recorder.RecordWebSocketMessage("client->server", msgTypeStr, data); err != nil {
					log.Printf("Failed to record WebSocket message: %v", err)
				}
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
					// Plain text input - check anti-cheat first
					if h.detector != nil {
						allowed, violations := h.detector.CheckInput(data)
						if !allowed {
							log.Printf("Anti-cheat: Input blocked")
							// Don't process this input
							continue
						}

						// Log violations
						for _, violation := range violations {
							if h.recorder != nil {
								eventJSON, _ := json.Marshal(violation)
								h.recorder.RecordEvent("anticheat_violation", string(eventJSON))
							}
						}
					}

					// Record keystroke input
					if h.recorder != nil {
						if err := h.recorder.RecordInput(data); err != nil {
							log.Printf("Failed to record input: %v", err)
						}
					}

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
				// Record binary input
				if h.recorder != nil {
					if err := h.recorder.RecordInput(data); err != nil {
						log.Printf("Failed to record input: %v", err)
					}
				}

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
		// Mark disconnection in session state
		if h.sessionState != nil {
			h.sessionState.Disconnect()
			log.Printf("Session state: %s (can reconnect for %v)",
				h.sessionState.GetState(),
				h.sessionState.ReconnectWindow)
		}
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

		// Update session state
		if h.sessionState != nil {
			h.sessionState.UpdateTerminalSize(resize.Cols, resize.Rows)
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
		// Parse client anti-cheat event
		var clientEvent map[string]interface{}
		if err := json.Unmarshal(msg.Data, &clientEvent); err != nil {
			log.Printf("Failed to parse anticheat event: %v", err)
			return err
		}

		// Process with detector
		if h.detector != nil {
			eventType := ""
			if et, ok := clientEvent["event"].(string); ok {
				eventType = et
			}
			event := h.detector.RecordClientEvent(eventType, clientEvent)

			// Record to file
			if h.recorder != nil {
				eventJSON, _ := json.Marshal(event)
				if err := h.recorder.RecordEvent("client_anticheat", string(eventJSON)); err != nil {
					log.Printf("Failed to record anticheat event: %v", err)
				}
			}
		}
		return nil

	default:
		log.Printf("Unknown message type: %s", msg.Type)
	}

	return nil
}
