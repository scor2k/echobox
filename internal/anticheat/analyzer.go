package anticheat

import (
	"bufio"
	"encoding/json"
	"fmt"
	"math"
	"os"
	"strconv"
	"strings"
	"time"
)

// TypingStats represents typing statistics
type TypingStats struct {
	TotalKeystrokes   int     `json:"total_keystrokes"`
	SessionDuration   float64 `json:"session_duration_seconds"`
	AverageWPM        float64 `json:"average_wpm"`
	MedianWPM         float64 `json:"median_wpm"`
	MaxWPM            float64 `json:"max_wpm"`
	MinWPM            float64 `json:"min_wpm"`
	WPMStdDev         float64 `json:"wpm_std_dev"`
	TypingIntervals   []float64 `json:"-"` // Not exported to JSON
	AnomaliesDetected int     `json:"anomalies_detected"`
}

// AnalysisReport represents the complete analysis of a session
type AnalysisReport struct {
	SessionID         string                 `json:"session_id"`
	CandidateName     string                 `json:"candidate_name"`
	AnalysisTimestamp time.Time              `json:"analysis_timestamp"`
	TypingStats       TypingStats            `json:"typing_stats"`
	AntiCheatEvents   []Event                `json:"anticheat_events"`
	EventSummary      map[string]int         `json:"event_summary"`
	Verdict           string                 `json:"verdict"`
	Confidence        float64                `json:"confidence_score"`
	Flags             []string               `json:"flags"`
	Recommendations   []string               `json:"recommendations"`
}

// AnalyzeSession performs post-session typing pattern analysis
func AnalyzeSession(sessionDir string) (*AnalysisReport, error) {
	// Read metadata
	metadataPath := fmt.Sprintf("%s/metadata.json", sessionDir)
	metadataBytes, err := os.ReadFile(metadataPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read metadata: %w", err)
	}

	var metadata struct {
		ID            string  `json:"id"`
		CandidateName string  `json:"candidate_name"`
		Duration      float64 `json:"duration_seconds"`
	}
	if err := json.Unmarshal(metadataBytes, &metadata); err != nil {
		return nil, fmt.Errorf("failed to parse metadata: %w", err)
	}

	// Analyze keystrokes
	typingStats, err := analyzeKeystrokes(sessionDir, metadata.Duration)
	if err != nil {
		return nil, fmt.Errorf("failed to analyze keystrokes: %w", err)
	}

	// Analyze events
	events, err := loadEvents(sessionDir)
	if err != nil {
		return nil, fmt.Errorf("failed to load events: %w", err)
	}

	// Generate verdict
	verdict, confidence, flags := generateVerdict(typingStats, events)

	// Generate recommendations
	recommendations := generateRecommendations(events, typingStats)

	report := &AnalysisReport{
		SessionID:         metadata.ID,
		CandidateName:     metadata.CandidateName,
		AnalysisTimestamp: time.Now(),
		TypingStats:       *typingStats,
		AntiCheatEvents:   events,
		EventSummary:      summarizeEvents(events),
		Verdict:           verdict,
		Confidence:        confidence,
		Flags:             flags,
		Recommendations:   recommendations,
	}

	return report, nil
}

// analyzeKeystrokes analyzes keystroke patterns from keystrokes.log
func analyzeKeystrokes(sessionDir string, sessionDuration float64) (*TypingStats, error) {
	keystrokesPath := fmt.Sprintf("%s/keystrokes.log", sessionDir)

	file, err := os.Open(keystrokesPath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	stats := &TypingStats{
		SessionDuration: sessionDuration,
		TypingIntervals: make([]float64, 0),
	}

	scanner := bufio.NewScanner(file)
	lastTimestamp := int64(0)

	for scanner.Scan() {
		line := scanner.Text()
		parts := strings.SplitN(line, " ", 2)
		if len(parts) < 2 {
			continue
		}

		timestamp, err := strconv.ParseInt(parts[0], 10, 64)
		if err != nil {
			continue
		}

		stats.TotalKeystrokes++

		if lastTimestamp > 0 {
			interval := float64(timestamp-lastTimestamp) / 1000.0 // Convert to seconds
			if interval > 0 && interval < 10 { // Ignore long pauses
				stats.TypingIntervals = append(stats.TypingIntervals, interval)
			}
		}

		lastTimestamp = timestamp
	}

	if err := scanner.Err(); err != nil {
		return nil, err
	}

	// Calculate WPM statistics
	calculateWPMStats(stats)

	return stats, nil
}

// calculateWPMStats calculates WPM metrics
func calculateWPMStats(stats *TypingStats) {
	if stats.SessionDuration == 0 || stats.TotalKeystrokes == 0 {
		return
	}

	// Average WPM (assuming 5 chars per word)
	stats.AverageWPM = (float64(stats.TotalKeystrokes) / 5.0) / (stats.SessionDuration / 60.0)

	// Calculate WPM in sliding windows for more detailed analysis
	if len(stats.TypingIntervals) < 10 {
		stats.MedianWPM = stats.AverageWPM
		stats.MaxWPM = stats.AverageWPM
		stats.MinWPM = stats.AverageWPM
		return
	}

	// Simple min/max detection from intervals
	minInterval := stats.TypingIntervals[0]
	maxInterval := stats.TypingIntervals[0]

	for _, interval := range stats.TypingIntervals {
		if interval < minInterval {
			minInterval = interval
		}
		if interval > maxInterval {
			maxInterval = interval
		}
	}

	// WPM is inversely proportional to interval
	// Fast typing = small intervals = high WPM
	stats.MaxWPM = (12.0 / minInterval) // 12 = 60s/min / 5 chars/word
	stats.MinWPM = (12.0 / maxInterval)
	stats.MedianWPM = stats.AverageWPM

	// Calculate standard deviation of intervals
	mean := 0.0
	for _, interval := range stats.TypingIntervals {
		mean += interval
	}
	mean /= float64(len(stats.TypingIntervals))

	variance := 0.0
	for _, interval := range stats.TypingIntervals {
		diff := interval - mean
		variance += diff * diff
	}
	variance /= float64(len(stats.TypingIntervals))
	stats.WPMStdDev = math.Sqrt(variance) * 12.0 // Convert to WPM scale
}

// loadEvents loads anti-cheat events from events.log
func loadEvents(sessionDir string) ([]Event, error) {
	eventsPath := fmt.Sprintf("%s/events.log", sessionDir)

	file, err := os.Open(eventsPath)
	if err != nil {
		if os.IsNotExist(err) {
			return []Event{}, nil
		}
		return nil, err
	}
	defer file.Close()

	events := make([]Event, 0)
	scanner := bufio.NewScanner(file)

	for scanner.Scan() {
		line := scanner.Text()
		parts := strings.SplitN(line, " ", 3)
		if len(parts) < 3 {
			continue
		}

		// Parse timestamp (milliseconds since session start)
		timestampMS, err := strconv.ParseInt(parts[0], 10, 64)
		if err != nil {
			continue
		}

		eventType := parts[1]
		dataStr := parts[2]

		// Parse JSON data
		var data map[string]interface{}
		if err := json.Unmarshal([]byte(dataStr), &data); err != nil {
			// Not JSON, treat as plain text
			data = map[string]interface{}{"raw": dataStr}
		}

		// Determine severity
		severity := SeverityInfo
		if eventType == "anticheat" {
			if eventStr, ok := data["event"].(string); ok {
				if strings.Contains(eventStr, "paste") {
					severity = SeverityCritical
				} else if strings.Contains(eventStr, "rapid") {
					severity = SeverityWarning
				}
			}
		}

		event := Event{
			Timestamp:   time.UnixMilli(timestampMS),
			Severity:    severity,
			Type:        eventType,
			Description: fmt.Sprintf("Event: %s", eventType),
			Data:        data,
		}

		events = append(events, event)
	}

	if err := scanner.Err(); err != nil {
		return nil, err
	}

	return events, nil
}

// summarizeEvents creates a summary of events by type
func summarizeEvents(events []Event) map[string]int {
	summary := make(map[string]int)

	for _, event := range events {
		summary[event.Type]++
		summary[string(event.Severity)]++
		summary["total"]++
	}

	return summary
}

// generateVerdict determines if session shows signs of cheating
func generateVerdict(stats *TypingStats, events []Event) (string, float64, []string) {
	flags := make([]string, 0)
	suspicionScore := 0.0

	// Check for critical events (paste attempts)
	criticalCount := 0
	warningCount := 0
	for _, event := range events {
		if event.Severity == SeverityCritical {
			criticalCount++
			suspicionScore += 30.0
		} else if event.Severity == SeverityWarning {
			warningCount++
			suspicionScore += 10.0
		}
	}

	if criticalCount > 0 {
		flags = append(flags, fmt.Sprintf("%d paste attempt(s) detected", criticalCount))
	}

	// Check typing patterns
	if stats.AverageWPM > 120 {
		flags = append(flags, fmt.Sprintf("Unusually high WPM: %.1f", stats.AverageWPM))
		suspicionScore += 20.0
	}

	if stats.WPMStdDev > 50 {
		flags = append(flags, fmt.Sprintf("High WPM variance: %.1f", stats.WPMStdDev))
		suspicionScore += 15.0
	}

	// Generate verdict
	verdict := "CLEAN"
	confidence := 0.95

	if suspicionScore >= 50 {
		verdict = "SUSPICIOUS"
		confidence = math.Min(suspicionScore/100.0, 0.95)
	} else if suspicionScore >= 30 {
		verdict = "REVIEW_RECOMMENDED"
		confidence = 0.70
	} else if suspicionScore >= 10 {
		verdict = "MINOR_CONCERNS"
		confidence = 0.85
	}

	if len(flags) == 0 {
		flags = append(flags, "No anomalies detected")
	}

	return verdict, confidence, flags
}

// generateRecommendations provides recommendations based on analysis
func generateRecommendations(events []Event, stats *TypingStats) []string {
	recommendations := make([]string, 0)

	criticalCount := 0
	for _, event := range events {
		if event.Severity == SeverityCritical {
			criticalCount++
		}
	}

	if criticalCount > 0 {
		recommendations = append(recommendations,
			"Review session replay carefully - paste attempts detected")
		recommendations = append(recommendations,
			"Cross-check solutions with other candidates for similarity")
	}

	if stats.AverageWPM > 120 {
		recommendations = append(recommendations,
			"Very high typing speed - verify coding patterns and approach authenticity")
	}

	if stats.TotalKeystrokes < 100 {
		recommendations = append(recommendations,
			"Very few keystrokes - candidate may not have engaged fully with tasks")
	}

	if len(recommendations) == 0 {
		recommendations = append(recommendations, "No specific concerns identified")
	}

	return recommendations
}

// SaveReport saves the analysis report to a JSON file with restricted permissions
func SaveReport(report *AnalysisReport, sessionDir string) error {
	reportPath := fmt.Sprintf("%s/analysis.json", sessionDir)

	data, err := json.MarshalIndent(report, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal report: %w", err)
	}

	// Create with read-only permissions (root only access)
	// Mode 0400 = owner (root) read-only, no access for others
	if err := os.WriteFile(reportPath, data, 0400); err != nil {
		return fmt.Errorf("failed to write report: %w", err)
	}

	return nil
}
