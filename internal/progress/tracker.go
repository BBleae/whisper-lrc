package progress

import (
	"fmt"
	"strings"
	"sync"
	"time"
)

// Tracker displays progress for batch processing
type Tracker struct {
	total     int
	current   int
	fileName  string
	status    string
	errors    []string
	completed []string
	mu        sync.Mutex
	done      chan struct{}
	started   bool
}

// NewTracker creates a new progress tracker
func NewTracker(total int) *Tracker {
	return &Tracker{
		total:     total,
		errors:    make([]string, 0),
		completed: make([]string, 0),
		done:      make(chan struct{}),
	}
}

// Start begins the progress display
func (t *Tracker) Start() {
	t.started = true
	go t.render()
}

// Stop ends the progress display
func (t *Tracker) Stop() {
	if t.started {
		close(t.done)
		t.started = false
		// Clear the progress line
		fmt.Print("\r" + strings.Repeat(" ", 80) + "\r")
	}
}

// SetCurrent updates the current file being processed
func (t *Tracker) SetCurrent(index int, fileName string) {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.current = index
	t.fileName = fileName
	t.status = "Processing..."
}

// SetStatus updates the status message
func (t *Tracker) SetStatus(status string) {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.status = status
}

// Complete marks a file as completed
func (t *Tracker) Complete(input, output string) {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.completed = append(t.completed, fmt.Sprintf("%s -> %s", input, output))
	t.printCompleted(input, output)
}

// Error records an error
func (t *Tracker) Error(input string, err error) {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.errors = append(t.errors, fmt.Sprintf("%s: %v", input, err))
	t.printError(input, err)
}

func (t *Tracker) render() {
	ticker := time.NewTicker(100 * time.Millisecond)
	defer ticker.Stop()

	spinChars := []string{"⠋", "⠙", "⠹", "⠸", "⠼", "⠴", "⠦", "⠧", "⠇", "⠏"}
	spinIdx := 0

	for {
		select {
		case <-t.done:
			return
		case <-ticker.C:
			t.mu.Lock()
			if t.fileName != "" {
				// Truncate filename if too long
				name := t.fileName
				if len(name) > 30 {
					name = name[:27] + "..."
				}

				progress := fmt.Sprintf("\r%s [%d/%d] %s: %s",
					spinChars[spinIdx],
					t.current,
					t.total,
					name,
					t.status,
				)
				// Pad with spaces and truncate to terminal width
				if len(progress) < 80 {
					progress += strings.Repeat(" ", 80-len(progress))
				} else {
					progress = progress[:80]
				}
				fmt.Print(progress)
			}
			t.mu.Unlock()
			spinIdx = (spinIdx + 1) % len(spinChars)
		}
	}
}

func (t *Tracker) printCompleted(input, output string) {
	// Clear progress line and print completion
	fmt.Printf("\r%s✓ %s -> %s\n", strings.Repeat(" ", 80)+"\r", truncate(input, 30), truncate(output, 30))
}

func (t *Tracker) printError(input string, err error) {
	// Clear progress line and print error
	fmt.Printf("\r%s✗ %s: %v\n", strings.Repeat(" ", 80)+"\r", truncate(input, 30), err)
}

func truncate(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen-3] + "..."
}
