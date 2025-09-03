package termflow

import (
	"fmt"
	"sync"
	"time"
)

// SpinnerStyle defines different spinner animations
type SpinnerStyle int

const (
	SpinnerDot SpinnerStyle = iota
	SpinnerLine
	SpinnerCircle
)

// Spinner represents an animated spinner
type Spinner struct {
	frames   []string
	current  int
	style    SpinnerStyle
	running  bool
	stopCh   chan struct{}
	mu       sync.RWMutex
	interval time.Duration
}

// NewSpinner creates a new spinner with the specified style
func NewSpinner(style SpinnerStyle) *Spinner {
	s := &Spinner{
		style:    style,
		current:  0,
		running:  false,
		interval: 80 * time.Millisecond,
		stopCh:   make(chan struct{}),
	}

	s.setFrames()
	return s
}

// setFrames sets the animation frames based on the style
func (s *Spinner) setFrames() {
	switch s.style {
	case SpinnerDot:
		s.frames = []string{"⠋", "⠙", "⠹", "⠸", "⠼", "⠴", "⠦", "⠧", "⠇", "⠏"}
	case SpinnerLine:
		s.frames = []string{"|", "/", "-", "\\"}
	case SpinnerCircle:
		s.frames = []string{"◐", "◓", "◑", "◒"}
	default:
		s.frames = []string{"⠋", "⠙", "⠹", "⠸", "⠼", "⠴", "⠦", "⠧", "⠇", "⠏"}
	}
}

// Start begins the spinner animation
func (s *Spinner) Start() {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.running {
		return
	}

	s.running = true
	s.stopCh = make(chan struct{})

	go s.animate()
}

// Stop stops the spinner animation
func (s *Spinner) Stop() {
	s.mu.Lock()
	defer s.mu.Unlock()

	if !s.running {
		return
	}

	s.running = false
	close(s.stopCh)
}

// animate runs the spinner animation loop
func (s *Spinner) animate() {
	ticker := time.NewTicker(s.interval)
	defer ticker.Stop()

	for {
		select {
		case <-s.stopCh:
			return
		case <-ticker.C:
			s.mu.Lock()
			s.current = (s.current + 1) % len(s.frames)
			s.mu.Unlock()
		}
	}
}

// Frame returns the current spinner frame
func (s *Spinner) Frame() string {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if len(s.frames) == 0 {
		return ""
	}

	return s.frames[s.current]
}

// IsRunning returns whether the spinner is currently running
func (s *Spinner) IsRunning() bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.running
}

// ThinkingSpinner provides a thinking indicator with spinner
type ThinkingSpinner struct {
	spinner *Spinner
	client  *InteractiveClient
	message string
}

// NewThinkingSpinner creates a new thinking spinner
func NewThinkingSpinner(client *InteractiveClient, message string) *ThinkingSpinner {
	return &ThinkingSpinner{
		spinner: NewSpinner(SpinnerDot),
		client:  client,
		message: message,
	}
}

// Start begins the thinking animation
func (ts *ThinkingSpinner) Start() {
	ts.spinner.Start()

	// Clear any existing input line and show thinking message
	ts.client.Printf("\033[2K\r") // Clear entire line and move cursor to beginning
	ts.showThinkingInitial()

	// Update the display periodically
	go ts.updateDisplay()
}

// Stop stops the thinking animation and clears the display
func (ts *ThinkingSpinner) Stop() {
	ts.spinner.Stop()

	// Clear the thinking line
	ts.clearThinking()
}

// showThinking displays the thinking message with spinner
func (ts *ThinkingSpinner) showThinking() {
	frame := ts.spinner.Frame()
	if frame != "" {
		// Colored spinner frame (cyan like bubbletea)
		coloredFrame := fmt.Sprintf("\033[1;38;5;87m%s\033[0m", frame)
		// Italic thinking text (like bubbletea)
		coloredMessage := fmt.Sprintf("\033[3;38;5;117m %s\033[0m", ts.message)
		ts.client.Printf("\n%s%s", coloredFrame, coloredMessage)
	}
}

// clearThinking clears the current thinking line
func (ts *ThinkingSpinner) clearThinking() {
	// Move cursor up one line, clear it completely
	ts.client.Printf("\033[1A\033[2K\r")
}

// showThinkingInitial displays the initial thinking message
func (ts *ThinkingSpinner) showThinkingInitial() {
	frame := ts.spinner.Frame()
	if frame != "" {
		// Colored spinner frame (cyan like bubbletea)
		coloredFrame := fmt.Sprintf("\033[1;38;5;87m%s\033[0m", frame)
		// Italic thinking text (like bubbletea)
		coloredMessage := fmt.Sprintf("\033[3;38;5;117m %s\033[0m", ts.message)
		ts.client.Printf("\n%s%s", coloredFrame, coloredMessage)
	}
}

// updateDisplay updates the spinner display while it's running
func (ts *ThinkingSpinner) updateDisplay() {
	ticker := time.NewTicker(80 * time.Millisecond)
	defer ticker.Stop()

	for ts.spinner.IsRunning() {
		select {
		case <-ticker.C:
			if ts.spinner.IsRunning() {
				// Move cursor up one line, clear it, and redraw with new frame
				ts.client.Printf("\033[1A\033[2K\r")
				frame := ts.spinner.Frame()
				if frame != "" {
					// Colored spinner frame (cyan like bubbletea)
					coloredFrame := fmt.Sprintf("\033[1;38;5;87m%s\033[0m", frame)
					// Italic thinking text (like bubbletea)
					coloredMessage := fmt.Sprintf("\033[3;38;5;117m %s\033[0m", ts.message)
					ts.client.Printf("\n%s%s", coloredFrame, coloredMessage)
				}
			}
		}
	}
}
