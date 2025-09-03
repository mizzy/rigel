package termflow

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// HistoryManager manages persistent command history
type HistoryManager struct {
	filepath   string
	maxEntries int
	entries    []string
}

// NewHistoryManager creates a new history manager
func NewHistoryManager(filename string) *HistoryManager {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		homeDir = "."
	}

	filepath := filepath.Join(homeDir, filename)

	return &HistoryManager{
		filepath:   filepath,
		maxEntries: 1000,
		entries:    []string{},
	}
}

// Load loads history from file
func (hm *HistoryManager) Load() error {
	file, err := os.Open(hm.filepath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil // File doesn't exist yet, that's OK
		}
		return fmt.Errorf("failed to open history file: %w", err)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line != "" {
			hm.entries = append(hm.entries, line)
		}
	}

	return scanner.Err()
}

// Save saves history to file
func (hm *HistoryManager) Save() error {
	// Create directory if it doesn't exist
	dir := filepath.Dir(hm.filepath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create history directory: %w", err)
	}

	file, err := os.Create(hm.filepath)
	if err != nil {
		return fmt.Errorf("failed to create history file: %w", err)
	}
	defer file.Close()

	for _, entry := range hm.entries {
		if _, err := fmt.Fprintln(file, entry); err != nil {
			return fmt.Errorf("failed to write history entry: %w", err)
		}
	}

	return nil
}

// Add adds an entry to the history
func (hm *HistoryManager) Add(entry string) {
	entry = strings.TrimSpace(entry)
	if entry == "" {
		return
	}

	// Remove duplicate if it exists at the end
	if len(hm.entries) > 0 && hm.entries[len(hm.entries)-1] == entry {
		return
	}

	hm.entries = append(hm.entries, entry)

	// Limit history size
	if len(hm.entries) > hm.maxEntries {
		hm.entries = hm.entries[1:]
	}
}

// GetAll returns all history entries
func (hm *HistoryManager) GetAll() []string {
	return append([]string{}, hm.entries...) // Return a copy
}

// Clear clears all history
func (hm *HistoryManager) Clear() error {
	hm.entries = []string{}
	return hm.Save()
}

// GetLatest returns the latest N entries
func (hm *HistoryManager) GetLatest(n int) []string {
	if n <= 0 || len(hm.entries) == 0 {
		return []string{}
	}

	start := len(hm.entries) - n
	if start < 0 {
		start = 0
	}

	return append([]string{}, hm.entries[start:]...)
}

// Search searches for entries containing the given text
func (hm *HistoryManager) Search(query string) []string {
	if query == "" {
		return []string{}
	}

	var matches []string
	query = strings.ToLower(query)

	for _, entry := range hm.entries {
		if strings.Contains(strings.ToLower(entry), query) {
			matches = append(matches, entry)
		}
	}

	return matches
}
