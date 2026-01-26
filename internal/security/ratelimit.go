package security

import (
	"sync"
	"time"
)

// InputEvent represents a single input event
type InputEvent struct {
	Timestamp time.Time
	Length    int
}

// RateLimiter tracks input rate to detect paste attempts
type RateLimiter struct {
	MaxCharsPerSecond int
	windowSize        time.Duration
	events            []InputEvent
	mu                sync.Mutex
}

// NewRateLimiter creates a new rate limiter
func NewRateLimiter(maxCharsPerSecond int) *RateLimiter {
	return &RateLimiter{
		MaxCharsPerSecond: maxCharsPerSecond,
		windowSize:        time.Second,
		events:            make([]InputEvent, 0),
	}
}

// CheckInput checks if input is within acceptable rate
// Returns: allowed (bool), currentRate (int), violation (bool)
func (r *RateLimiter) CheckInput(length int) (bool, int, bool) {
	r.mu.Lock()
	defer r.mu.Unlock()

	now := time.Now()
	event := InputEvent{
		Timestamp: now,
		Length:    length,
	}

	// Add current event
	r.events = append(r.events, event)

	// Remove events outside the window
	cutoff := now.Add(-r.windowSize)
	validEvents := make([]InputEvent, 0)
	for _, e := range r.events {
		if e.Timestamp.After(cutoff) {
			validEvents = append(validEvents, e)
		}
	}
	r.events = validEvents

	// Calculate total characters in current window
	totalChars := 0
	for _, e := range r.events {
		totalChars += e.Length
	}

	// Check if rate is exceeded
	currentRate := totalChars
	violation := currentRate > r.MaxCharsPerSecond

	// Allow input but report violation
	return true, currentRate, violation
}

// GetCurrentRate returns the current input rate (chars/second)
func (r *RateLimiter) GetCurrentRate() int {
	r.mu.Lock()
	defer r.mu.Unlock()

	now := time.Now()
	cutoff := now.Add(-r.windowSize)

	totalChars := 0
	for _, e := range r.events {
		if e.Timestamp.After(cutoff) {
			totalChars += e.Length
		}
	}

	return totalChars
}

// Reset clears all tracked events
func (r *RateLimiter) Reset() {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.events = make([]InputEvent, 0)
}

// BurstDetector detects rapid bursts of input (potential paste)
type BurstDetector struct {
	maxCharsInBurst int
	burstWindow     time.Duration
	lastInput       time.Time
	burstChars      int
	mu              sync.Mutex
}

// NewBurstDetector creates a new burst detector
func NewBurstDetector(maxCharsInBurst int, burstWindow time.Duration) *BurstDetector {
	return &BurstDetector{
		maxCharsInBurst: maxCharsInBurst,
		burstWindow:     burstWindow,
		lastInput:       time.Time{},
		burstChars:      0,
	}
}

// CheckBurst checks for rapid input burst
// Returns: isBurst (bool), burstSize (int)
func (b *BurstDetector) CheckBurst(length int) (bool, int) {
	b.mu.Lock()
	defer b.mu.Unlock()

	now := time.Now()

	// If this is a new burst (after window expired)
	if b.lastInput.IsZero() || now.Sub(b.lastInput) > b.burstWindow {
		b.burstChars = length
		b.lastInput = now
		return false, length
	}

	// Add to current burst
	b.burstChars += length
	b.lastInput = now

	// Check if burst threshold exceeded
	isBurst := b.burstChars > b.maxCharsInBurst

	return isBurst, b.burstChars
}

// Reset clears burst tracking
func (b *BurstDetector) Reset() {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.burstChars = 0
	b.lastInput = time.Time{}
}
