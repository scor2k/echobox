package web

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/akonyukov/echobox/internal/config"
)

// Server represents the HTTP server
type Server struct {
	config     *config.Config
	httpServer *http.Server
	wsHandler  *WSHandler
}

// New creates a new HTTP server
func New(cfg *config.Config, wsHandler *WSHandler) *Server {
	s := &Server{
		config:    cfg,
		wsHandler: wsHandler,
	}

	mux := http.NewServeMux()

	// WebSocket endpoint
	mux.HandleFunc("/ws", wsHandler.Handle)

	// Health check endpoint
	mux.HandleFunc("/health", s.handleHealth)

	// Serve static files
	mux.HandleFunc("/", s.handleStatic)

	s.httpServer = &http.Server{
		Addr:         fmt.Sprintf(":%d", cfg.Port),
		Handler:      s.addMiddleware(mux),
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	return s
}

// Start starts the HTTP server
func (s *Server) Start() error {
	log.Printf("Starting server on port %d...", s.config.Port)
	if err := s.httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		return fmt.Errorf("server error: %w", err)
	}
	return nil
}

// Shutdown gracefully shuts down the server
func (s *Server) Shutdown(ctx context.Context) error {
	log.Println("Shutting down server...")

	// Signal WebSocket handler to close all connections
	s.wsHandler.Shutdown()

	// Shutdown HTTP server (waits for connections to close)
	err := s.httpServer.Shutdown(ctx)

	if err == context.DeadlineExceeded {
		log.Println("Shutdown timeout exceeded, forcing close...")
	} else if err != nil {
		log.Printf("Shutdown error: %v", err)
	} else {
		log.Println("HTTP server shutdown complete")
	}

	return err
}

// addMiddleware adds common middleware to the handler
func (s *Server) addMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Security headers
		w.Header().Set("X-Frame-Options", "DENY")
		w.Header().Set("X-Content-Type-Options", "nosniff")
		w.Header().Set("X-XSS-Protection", "1; mode=block")
		w.Header().Set("Content-Security-Policy", "default-src 'self' 'unsafe-inline' 'unsafe-eval'")

		// Logging
		log.Printf("%s %s %s", r.Method, r.URL.Path, r.RemoteAddr)

		next.ServeHTTP(w, r)
	})
}

// handleHealth handles health check requests
func (s *Server) handleHealth(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, `{"status":"ok","candidate":"%s"}`, s.config.CandidateName)
}

// handleStatic serves static files
func (s *Server) handleStatic(w http.ResponseWriter, r *http.Request) {
	// Get the web directory path (relative to binary location)
	webDir := "./web"

	// Map URL path to file
	path := r.URL.Path
	if path == "/" {
		path = "/index.html"
	}

	filePath := filepath.Join(webDir, path)

	// Check if file exists
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		http.NotFound(w, r)
		return
	}

	// Serve the file
	http.ServeFile(w, r, filePath)
}
