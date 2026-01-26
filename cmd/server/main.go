package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/akonyukov/echobox/internal/anticheat"
	"github.com/akonyukov/echobox/internal/config"
	"github.com/akonyukov/echobox/internal/session"
	"github.com/akonyukov/echobox/internal/terminal"
	"github.com/akonyukov/echobox/internal/web"
)

func main() {
	// Set up logging
	log.SetFlags(log.LstdFlags | log.Lshortfile)
	log.Println("Starting SRE Interview Terminal...")

	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	log.Printf("Configuration loaded: Candidate=%s, Port=%d, Timeout=%v",
		cfg.CandidateName, cfg.Port, cfg.SessionTimeout)

	// Print MOTD
	log.Println(cfg.MOTD)

	// Create session manager
	sessionMgr, err := session.NewManager(cfg.OutputDir, cfg.CandidateName)
	if err != nil {
		log.Fatalf("Failed to create session manager: %v", err)
	}

	log.Printf("Session created: %s in %s", sessionMgr.GetSession().ID, sessionMgr.GetSessionDir())

	// Create recorder
	recorder, err := terminal.NewRecorder(sessionMgr.GetSessionDir(), cfg.FlushInterval)
	if err != nil {
		log.Fatalf("Failed to create recorder: %v", err)
	}
	defer recorder.Close()

	log.Println("Recorder initialized")

	// Create anti-cheat detector
	detector := anticheat.NewDetector(cfg.InputRateLimit)
	log.Printf("Anti-cheat detector initialized (rate limit: %d chars/sec)", cfg.InputRateLimit)

	// Create PTY
	pty, err := terminal.New(cfg.Shell)
	if err != nil {
		log.Fatalf("Failed to create PTY: %v", err)
	}
	defer pty.Close()

	log.Printf("PTY created, shell: %s", cfg.Shell)

	// Create WebSocket handler with recorder and detector
	wsHandler := web.NewWSHandler(pty, recorder, detector)

	// Create HTTP server
	server := web.New(cfg, wsHandler)

	// Set up graceful shutdown
	serverErrors := make(chan error, 1)
	shutdown := make(chan os.Signal, 1)
	signal.Notify(shutdown, syscall.SIGINT, syscall.SIGTERM)

	// Start server in goroutine
	go func() {
		log.Printf("Server listening on http://0.0.0.0:%d", cfg.Port)
		serverErrors <- server.Start()
	}()

	// Wait for shutdown signal, session finish, or server error
	select {
	case err := <-serverErrors:
		log.Fatalf("Server error: %v", err)

	case <-wsHandler.FinishSignal():
		log.Println("Session finished, initiating shutdown...")

		// Close PTY first to unblock any reads
		log.Println("Closing PTY...")
		if err := pty.Close(); err != nil {
			log.Printf("Error closing PTY: %v", err)
		}

		// Close recorder and flush logs
		log.Println("Closing recorder...")
		if err := recorder.Close(); err != nil {
			log.Printf("Error closing recorder: %v", err)
		}

		// Extract commands and complete session
		log.Println("Extracting commands...")
		if err := terminal.ExtractCommands(sessionMgr.GetSessionDir()); err != nil {
			log.Printf("Error extracting commands: %v", err)
		}

		// Generate anti-cheat analysis report
		log.Println("Generating anti-cheat analysis...")
		if report, err := anticheat.AnalyzeSession(sessionMgr.GetSessionDir()); err != nil {
			log.Printf("Error analyzing session: %v", err)
		} else {
			if err := anticheat.SaveReport(report, sessionMgr.GetSessionDir()); err != nil {
				log.Printf("Error saving analysis report: %v", err)
			} else {
				log.Printf("Analysis: %s (confidence: %.2f)", report.Verdict, report.Confidence)
			}
		}

		log.Println("Finalizing session metadata...")
		if err := sessionMgr.Complete(); err != nil {
			log.Printf("Error completing session: %v", err)
		}

		// Give time for WebSocket connections to close
		time.Sleep(500 * time.Millisecond)

		// Shutdown server
		log.Println("Shutting down HTTP server...")
		ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
		defer cancel()

		if err := server.Shutdown(ctx); err != nil {
			log.Printf("Error during server shutdown: %v", err)
		}

		log.Printf("Session complete: %s", sessionMgr.GetSessionDir())
		os.Exit(0)

	case sig := <-shutdown:
		log.Printf("Received signal %v, starting graceful shutdown...", sig)

		// Mark session as interrupted
		sessionMgr.SetMetadata("interrupted", true)
		sessionMgr.SetMetadata("signal", sig.String())

		// Close PTY first to unblock any reads
		log.Println("Closing PTY...")
		if err := pty.Close(); err != nil {
			log.Printf("Error closing PTY: %v", err)
		}

		// Close recorder and flush logs
		log.Println("Closing recorder...")
		if err := recorder.Close(); err != nil {
			log.Printf("Error closing recorder: %v", err)
		}

		// Extract commands
		log.Println("Extracting commands...")
		if err := terminal.ExtractCommands(sessionMgr.GetSessionDir()); err != nil {
			log.Printf("Error extracting commands: %v", err)
		}

		// Generate anti-cheat analysis report
		log.Println("Generating anti-cheat analysis...")
		if report, err := anticheat.AnalyzeSession(sessionMgr.GetSessionDir()); err != nil {
			log.Printf("Error analyzing session: %v", err)
		} else {
			if err := anticheat.SaveReport(report, sessionMgr.GetSessionDir()); err != nil {
				log.Printf("Error saving analysis report: %v", err)
			} else {
				log.Printf("Analysis: %s (confidence: %.2f)", report.Verdict, report.Confidence)
			}
		}

		// Complete session with error status
		log.Println("Finalizing session metadata...")
		sessionMgr.GetSession().Status = "interrupted"
		if err := sessionMgr.Complete(); err != nil {
			log.Printf("Error completing session: %v", err)
		}

		// Give time for WebSocket connections to close
		time.Sleep(500 * time.Millisecond)

		// Shutdown server with timeout
		log.Println("Shutting down HTTP server...")
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		if err := server.Shutdown(ctx); err != nil {
			log.Printf("Error during server shutdown: %v", err)
		}

		log.Printf("Graceful shutdown complete: %s", sessionMgr.GetSessionDir())

	}
}
