package history

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"
)

const (
	maxHistorySize = 10000 // Maximum number of entries to keep
	rigelDir       = ".rigel"
	historyFile    = "history"
)

// Entry represents a single history entry
type Entry struct {
	Command   string    `json:"command"`
	Timestamp time.Time `json:"timestamp"`
}

// Manager handles command history persistence
type Manager struct {
	filePath string
	entries  []Entry
	maxSize  int
}

// GetRigelDir returns the path to the Rigel configuration directory
func GetRigelDir() (string, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("failed to get home directory: %w", err)
	}
	return filepath.Join(homeDir, rigelDir), nil
}

// NewManager creates a new history manager
func NewManager() (*Manager, error) {
	rigelPath, err := GetRigelDir()
	if err != nil {
		return nil, err
	}

	// Create ~/.rigel directory if it doesn't exist
	if err := os.MkdirAll(rigelPath, 0755); err != nil {
		return nil, fmt.Errorf("failed to create .rigel directory: %w", err)
	}

	manager := &Manager{
		filePath: filepath.Join(rigelPath, historyFile),
		entries:  []Entry{},
		maxSize:  maxHistorySize,
	}

	// Migrate old history file if it exists
	if err := manager.migrateOldHistory(); err != nil {
		// Log the error but don't fail
		fmt.Fprintf(os.Stderr, "Warning: Failed to migrate old history: %v\n", err)
	}

	return manager, nil
}

// migrateOldHistory migrates history from old location (~/.rigel_history) to new location
func (m *Manager) migrateOldHistory() error {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return err
	}

	oldPath := filepath.Join(homeDir, ".rigel_history")

	// Check if old file exists
	if _, err := os.Stat(oldPath); os.IsNotExist(err) {
		return nil // No old file to migrate
	}

	// Check if new file already exists
	if _, err := os.Stat(m.filePath); err == nil {
		return nil // New file already exists, don't migrate
	}

	// Read old file
	oldFile, err := os.Open(oldPath)
	if err != nil {
		return fmt.Errorf("failed to open old history file: %w", err)
	}
	defer oldFile.Close()

	scanner := bufio.NewScanner(oldFile)
	for scanner.Scan() {
		var entry Entry
		if err := json.Unmarshal(scanner.Bytes(), &entry); err != nil {
			// Skip invalid entries
			continue
		}
		m.entries = append(m.entries, entry)
	}

	if err := scanner.Err(); err != nil {
		return fmt.Errorf("failed to read old history file: %w", err)
	}

	// Save to new location
	if err := m.Save(); err != nil {
		return fmt.Errorf("failed to save migrated history: %w", err)
	}

	// Optionally remove old file after successful migration
	// Commented out for safety - user can manually remove if desired
	// os.Remove(oldPath)

	return nil
}

// Load reads history from the file
func (m *Manager) Load() error {
	file, err := os.Open(m.filePath)
	if err != nil {
		if os.IsNotExist(err) {
			// File doesn't exist yet, that's ok
			return nil
		}
		return fmt.Errorf("failed to open history file: %w", err)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	m.entries = []Entry{}

	for scanner.Scan() {
		var entry Entry
		if err := json.Unmarshal(scanner.Bytes(), &entry); err != nil {
			// Skip invalid entries
			continue
		}
		m.entries = append(m.entries, entry)
	}

	if err := scanner.Err(); err != nil {
		return fmt.Errorf("failed to read history file: %w", err)
	}

	// Trim if over max size
	if len(m.entries) > m.maxSize {
		m.entries = m.entries[len(m.entries)-m.maxSize:]
	}

	return nil
}

// Save writes history to the file
func (m *Manager) Save() error {
	// Create directory if it doesn't exist
	dir := filepath.Dir(m.filePath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create directory: %w", err)
	}

	file, err := os.Create(m.filePath)
	if err != nil {
		return fmt.Errorf("failed to create history file: %w", err)
	}
	defer file.Close()

	writer := bufio.NewWriter(file)
	defer writer.Flush()

	for _, entry := range m.entries {
		data, err := json.Marshal(entry)
		if err != nil {
			continue // Skip entries that can't be marshaled
		}
		if _, err := writer.Write(data); err != nil {
			return fmt.Errorf("failed to write entry: %w", err)
		}
		if _, err := writer.WriteString("\n"); err != nil {
			return fmt.Errorf("failed to write newline: %w", err)
		}
	}

	return nil
}

// Add adds a new command to history
func (m *Manager) Add(command string) error {
	// Don't add empty commands
	if command == "" {
		return nil
	}

	// Don't add duplicate consecutive commands
	if len(m.entries) > 0 && m.entries[len(m.entries)-1].Command == command {
		return nil
	}

	entry := Entry{
		Command:   command,
		Timestamp: time.Now(),
	}

	m.entries = append(m.entries, entry)

	// Trim if over max size
	if len(m.entries) > m.maxSize {
		m.entries = m.entries[1:]
	}

	// Save immediately for persistence
	return m.Save()
}

// GetCommands returns all commands as a string slice
func (m *Manager) GetCommands() []string {
	commands := make([]string, len(m.entries))
	for i, entry := range m.entries {
		commands[i] = entry.Command
	}
	return commands
}

// GetRecentCommands returns the most recent n commands
func (m *Manager) GetRecentCommands(n int) []string {
	if n > len(m.entries) {
		n = len(m.entries)
	}

	commands := make([]string, n)
	start := len(m.entries) - n
	for i := 0; i < n; i++ {
		commands[i] = m.entries[start+i].Command
	}
	return commands
}

// Clear removes all history
func (m *Manager) Clear() error {
	m.entries = []Entry{}
	return m.Save()
}

// Size returns the number of entries in history
func (m *Manager) Size() int {
	return len(m.entries)
}
