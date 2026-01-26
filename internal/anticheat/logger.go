package anticheat

import (
	"encoding/json"
	"fmt"
	"time"
)

// Severity levels for anti-cheat events
type Severity string

const (
	SeverityInfo     Severity = "info"
	SeverityWarning  Severity = "warning"
	SeverityCritical Severity = "critical"
)

// Event represents an anti-cheat event
type Event struct {
	Timestamp   time.Time              `json:"timestamp"`
	Severity    Severity               `json:"severity"`
	Type        string                 `json:"type"`
	Description string                 `json:"description"`
	Data        map[string]interface{} `json:"data,omitempty"`
}

// Logger logs anti-cheat events
type Logger struct {
	events []Event
}

// NewLogger creates a new anti-cheat logger
func NewLogger() *Logger {
	return &Logger{
		events: make([]Event, 0),
	}
}

// LogPasteAttempt logs a paste attempt
func (l *Logger) LogPasteAttempt(source string, length int) *Event {
	event := &Event{
		Timestamp:   time.Now(),
		Severity:    SeverityCritical,
		Type:        "paste_attempt",
		Description: fmt.Sprintf("Paste attempt detected from %s", source),
		Data: map[string]interface{}{
			"source": source,
			"length": length,
		},
	}
	l.events = append(l.events, *event)
	return event
}

// LogRapidInput logs rapid input detection
func (l *Logger) LogRapidInput(charsPerSecond int, burstSize int) *Event {
	event := &Event{
		Timestamp:   time.Now(),
		Severity:    SeverityWarning,
		Type:        "rapid_input",
		Description: fmt.Sprintf("Rapid input detected: %d chars/sec", charsPerSecond),
		Data: map[string]interface{}{
			"chars_per_second": charsPerSecond,
			"burst_size":       burstSize,
		},
	}
	l.events = append(l.events, *event)
	return event
}

// LogTypingAnomaly logs unusual typing pattern
func (l *Logger) LogTypingAnomaly(anomalyType string, details map[string]interface{}) *Event {
	event := &Event{
		Timestamp:   time.Now(),
		Severity:    SeverityWarning,
		Type:        "typing_anomaly",
		Description: fmt.Sprintf("Typing anomaly detected: %s", anomalyType),
		Data:        details,
	}
	l.events = append(l.events, *event)
	return event
}

// LogFocusLoss logs window/tab focus loss
func (l *Logger) LogFocusLoss(duration time.Duration) *Event {
	event := &Event{
		Timestamp:   time.Now(),
		Severity:    SeverityInfo,
		Type:        "focus_loss",
		Description: "Window/tab lost focus",
		Data: map[string]interface{}{
			"duration_ms": duration.Milliseconds(),
		},
	}
	l.events = append(l.events, *event)
	return event
}

// LogCustomEvent logs a custom anti-cheat event
func (l *Logger) LogCustomEvent(severity Severity, eventType string, description string, data map[string]interface{}) *Event {
	event := &Event{
		Timestamp:   time.Now(),
		Severity:    severity,
		Type:        eventType,
		Description: description,
		Data:        data,
	}
	l.events = append(l.events, *event)
	return event
}

// GetEvents returns all logged events
func (l *Logger) GetEvents() []Event {
	return l.events
}

// GetEventsBySeverity returns events filtered by severity
func (l *Logger) GetEventsBySeverity(severity Severity) []Event {
	filtered := make([]Event, 0)
	for _, event := range l.events {
		if event.Severity == severity {
			filtered = append(filtered, event)
		}
	}
	return filtered
}

// GetSummary returns a summary of events by type and severity
func (l *Logger) GetSummary() map[string]int {
	summary := make(map[string]int)

	for _, event := range l.events {
		key := fmt.Sprintf("%s_%s", event.Severity, event.Type)
		summary[key]++
		summary["total"]++
		summary[string(event.Severity)]++
	}

	return summary
}

// ToJSON converts events to JSON
func (l *Logger) ToJSON() ([]byte, error) {
	return json.MarshalIndent(l.events, "", "  ")
}

// FormatEvent formats an event for logging
func (e *Event) FormatEvent() string {
	data, _ := json.Marshal(e.Data)
	return fmt.Sprintf("[%s] [%s] %s: %s | %s",
		e.Timestamp.Format("15:04:05.000"),
		e.Severity,
		e.Type,
		e.Description,
		string(data))
}
