package output

import (
	"fmt"
	"strings"

	"github.com/BBleae/whisper-lrc/internal/whisper"
)

// Formatter interface for output format implementations
type Formatter interface {
	Format(result *whisper.TranscriptionResult) string
}

// LRCFormatter formats transcription as LRC lyrics
type LRCFormatter struct{}

// NewLRCFormatter creates a new LRC formatter
func NewLRCFormatter() *LRCFormatter {
	return &LRCFormatter{}
}

// Format converts transcription result to LRC format
func (f *LRCFormatter) Format(result *whisper.TranscriptionResult) string {
	var sb strings.Builder

	// Add metadata header
	sb.WriteString("[re:whisper-lrc]\n")
	if result.Language != "" {
		sb.WriteString(fmt.Sprintf("[la:%s]\n", result.Language))
	}
	sb.WriteString("\n")

	// Add lyrics with timestamps
	for _, seg := range result.Segments {
		timestamp := formatLRCTimestamp(seg.Start)
		text := strings.TrimSpace(seg.Text)
		sb.WriteString(fmt.Sprintf("[%s]%s\n", timestamp, text))
	}

	return sb.String()
}

// formatLRCTimestamp converts seconds to LRC timestamp format [mm:ss.xx]
func formatLRCTimestamp(seconds float64) string {
	totalMs := int(seconds * 1000)
	mins := totalMs / 60000
	secs := (totalMs % 60000) / 1000
	centisecs := (totalMs % 1000) / 10

	return fmt.Sprintf("%02d:%02d.%02d", mins, secs, centisecs)
}

// SRTFormatter formats transcription as SRT subtitles
type SRTFormatter struct{}

// NewSRTFormatter creates a new SRT formatter
func NewSRTFormatter() *SRTFormatter {
	return &SRTFormatter{}
}

// Format converts transcription result to SRT format
func (f *SRTFormatter) Format(result *whisper.TranscriptionResult) string {
	var sb strings.Builder

	for i, seg := range result.Segments {
		// Sequence number (1-based)
		sb.WriteString(fmt.Sprintf("%d\n", i+1))

		// Timestamps: 00:00:00,000 --> 00:00:00,000
		startTS := formatSRTTimestamp(seg.Start)
		endTS := formatSRTTimestamp(seg.End)
		sb.WriteString(fmt.Sprintf("%s --> %s\n", startTS, endTS))

		// Text
		text := strings.TrimSpace(seg.Text)
		sb.WriteString(text + "\n")

		// Blank line separator
		sb.WriteString("\n")
	}

	return sb.String()
}

// formatSRTTimestamp converts seconds to SRT timestamp format 00:00:00,000
func formatSRTTimestamp(seconds float64) string {
	totalMs := int(seconds * 1000)
	hours := totalMs / 3600000
	mins := (totalMs % 3600000) / 60000
	secs := (totalMs % 60000) / 1000
	ms := totalMs % 1000

	return fmt.Sprintf("%02d:%02d:%02d,%03d", hours, mins, secs, ms)
}
