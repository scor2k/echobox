package anticheat

import (
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/akonyukov/echobox/internal/security"
)

// Detector performs real-time anti-cheat detection
type Detector struct {
	rateLimiter   *security.RateLimiter
	burstDetector *security.BurstDetector
	logger        *Logger

	// Typing pattern tracking
	keystrokeCount   int
	sessionStartTime time.Time
	lastKeystroke    time.Time

	mu sync.Mutex
}

// NewDetector creates a new anti-cheat detector
func NewDetector(maxCharsPerSecond int) *Detector {
	return &Detector{
		rateLimiter:      security.NewRateLimiter(maxCharsPerSecond),
		burstDetector:    security.NewBurstDetector(30, 100*time.Millisecond), // 30 chars in 100ms
		logger:           NewLogger(),
		sessionStartTime: time.Now(),
		lastKeystroke:    time.Now(),
	}
}

// CheckInput checks input for anti-cheat violations
// Returns: allowed (bool), violations ([]Event)
func (d *Detector) CheckInput(data []byte) (bool, []Event) {
	d.mu.Lock()
	defer d.mu.Unlock()

	violations := make([]Event, 0)
	length := len(data)

	// Update keystroke tracking
	d.keystrokeCount += length
	now := time.Now()
	timeSinceLastKey := now.Sub(d.lastKeystroke)
	d.lastKeystroke = now

	// Check rate limit
	allowed, currentRate, rateViolation := d.rateLimiter.CheckInput(length)
	if rateViolation {
		event := d.logger.LogRapidInput(currentRate, length)
		violations = append(violations, *event)
		log.Printf("Anti-cheat: Rate limit exceeded - %d chars/sec (limit: %d)",
			currentRate, d.rateLimiter.MaxCharsPerSecond)
	}

	// Check for burst (paste detection)
	isBurst, burstSize := d.burstDetector.CheckBurst(length)
	if isBurst {
		event := d.logger.LogPasteAttempt("server_burst_detection", burstSize)
		violations = append(violations, *event)
		log.Printf("Anti-cheat: Paste detected - %d chars in burst", burstSize)
	}

	// Detect abnormally fast typing
	if length > 1 && timeSinceLastKey < 50*time.Millisecond {
		// Multiple characters in <50ms is suspicious
		event := d.logger.LogTypingAnomaly("fast_multi_char", map[string]interface{}{
			"chars":              length,
			"time_since_last_ms": timeSinceLastKey.Milliseconds(),
		})
		violations = append(violations, *event)
	}

	return allowed, violations
}

// RecordClientEvent records an event from the client (paste, focus, etc.)
func (d *Detector) RecordClientEvent(eventType string, data map[string]interface{}) *Event {
	d.mu.Lock()
	defer d.mu.Unlock()

	var severity Severity
	var description string

	switch eventType {
	case "paste_attempt":
		severity = SeverityCritical
		description = "Client-side paste attempt blocked"
	case "rapid_input":
		severity = SeverityWarning
		description = "Client detected rapid input"
	case "window_focus":
		severity = SeverityInfo
		if gained, ok := data["gained"].(bool); ok && !gained {
			description = "Window lost focus"
		} else {
			description = "Window gained focus"
		}
	case "tab_visibility":
		severity = SeverityInfo
		if hidden, ok := data["hidden"].(bool); ok && hidden {
			description = "Tab hidden"
		} else {
			description = "Tab visible"
		}
	default:
		severity = SeverityInfo
		description = fmt.Sprintf("Client event: %s", eventType)
	}

	event := d.logger.LogCustomEvent(severity, eventType, description, data)
	return event
}

// GetStatistics returns session statistics
func (d *Detector) GetStatistics() map[string]interface{} {
	d.mu.Lock()
	defer d.mu.Unlock()

	duration := time.Since(d.sessionStartTime).Seconds()
	wpm := 0.0
	if duration > 0 {
		// Rough WPM calculation: assume 5 chars per word
		wpm = (float64(d.keystrokeCount) / 5.0) / (duration / 60.0)
	}

	return map[string]interface{}{
		"total_keystrokes":  d.keystrokeCount,
		"session_duration":  duration,
		"average_wpm":       wpm,
		"current_rate":      d.rateLimiter.GetCurrentRate(),
		"event_summary":     d.logger.GetSummary(),
		"critical_events":   len(d.logger.GetEventsBySeverity(SeverityCritical)),
		"warning_events":    len(d.logger.GetEventsBySeverity(SeverityWarning)),
		"info_events":       len(d.logger.GetEventsBySeverity(SeverityInfo)),
	}
}

// GetLogger returns the event logger
func (d *Detector) GetLogger() *Logger {
	return d.logger
}
