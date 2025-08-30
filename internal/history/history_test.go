package history

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestHistoryManager(t *testing.T) {
	// Create a temporary directory for testing
	tempDir, err := os.MkdirTemp("", "rigel_history_test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Create .rigel directory in temp dir
	rigelDir := filepath.Join(tempDir, ".rigel")
	if err := os.MkdirAll(rigelDir, 0755); err != nil {
		t.Fatalf("Failed to create .rigel dir: %v", err)
	}

	// Create a manager with a custom path
	manager := &Manager{
		filePath: filepath.Join(rigelDir, "history"),
		entries:  []Entry{},
		maxSize:  100,
	}

	t.Run("Add and Get Commands", func(t *testing.T) {
		// Add some commands
		err := manager.Add("echo hello")
		if err != nil {
			t.Errorf("Failed to add command: %v", err)
		}

		err = manager.Add("ls -la")
		if err != nil {
			t.Errorf("Failed to add command: %v", err)
		}

		err = manager.Add("cd /tmp")
		if err != nil {
			t.Errorf("Failed to add command: %v", err)
		}

		// Get all commands
		commands := manager.GetCommands()
		if len(commands) != 3 {
			t.Errorf("Expected 3 commands, got %d", len(commands))
		}

		if commands[0] != "echo hello" {
			t.Errorf("Expected first command to be 'echo hello', got '%s'", commands[0])
		}
	})

	t.Run("No Duplicate Consecutive Commands", func(t *testing.T) {
		initialSize := manager.Size()

		// Add the same command twice
		err := manager.Add("cd /tmp")
		if err != nil {
			t.Errorf("Failed to add command: %v", err)
		}

		// Size should not increase
		if manager.Size() != initialSize {
			t.Errorf("Expected size to remain %d, got %d", initialSize, manager.Size())
		}
	})

	t.Run("Get Recent Commands", func(t *testing.T) {
		recent := manager.GetRecentCommands(2)
		if len(recent) != 2 {
			t.Errorf("Expected 2 recent commands, got %d", len(recent))
		}

		if recent[1] != "cd /tmp" {
			t.Errorf("Expected last command to be 'cd /tmp', got '%s'", recent[1])
		}
	})

	t.Run("Save and Load", func(t *testing.T) {
		// Save current state
		err := manager.Save()
		if err != nil {
			t.Errorf("Failed to save history: %v", err)
		}

		// Create a new manager and load
		manager2 := &Manager{
			filePath: filepath.Join(rigelDir, "history"),
			entries:  []Entry{},
			maxSize:  100,
		}

		err = manager2.Load()
		if err != nil {
			t.Errorf("Failed to load history: %v", err)
		}

		// Check that the loaded history matches
		commands := manager2.GetCommands()
		if len(commands) != 3 {
			t.Errorf("Expected 3 commands after load, got %d", len(commands))
		}

		if commands[0] != "echo hello" {
			t.Errorf("Expected first command to be 'echo hello' after load, got '%s'", commands[0])
		}
	})

	t.Run("Clear History", func(t *testing.T) {
		err := manager.Clear()
		if err != nil {
			t.Errorf("Failed to clear history: %v", err)
		}

		if manager.Size() != 0 {
			t.Errorf("Expected size to be 0 after clear, got %d", manager.Size())
		}

		// Verify file is also cleared
		manager3 := &Manager{
			filePath: filepath.Join(rigelDir, "history"),
			entries:  []Entry{},
			maxSize:  100,
		}

		err = manager3.Load()
		if err != nil {
			t.Errorf("Failed to load history after clear: %v", err)
		}

		if manager3.Size() != 0 {
			t.Errorf("Expected loaded history to be empty after clear, got %d entries", manager3.Size())
		}
	})

	t.Run("Max Size Limit", func(t *testing.T) {
		// Create a manager with small max size
		smallManager := &Manager{
			filePath: filepath.Join(rigelDir, "history_small"),
			entries:  []Entry{},
			maxSize:  3,
		}

		// Add more than max size
		for i := 0; i < 5; i++ {
			err := smallManager.Add(string(rune('a' + i)))
			if err != nil {
				t.Errorf("Failed to add command: %v", err)
			}
		}

		// Should only have max size entries
		if smallManager.Size() != 3 {
			t.Errorf("Expected size to be 3, got %d", smallManager.Size())
		}

		// Should have the last 3 commands
		commands := smallManager.GetCommands()
		if commands[0] != "c" || commands[1] != "d" || commands[2] != "e" {
			t.Errorf("Expected last 3 commands (c,d,e), got %v", commands)
		}
	})

	t.Run("Empty Command Not Added", func(t *testing.T) {
		manager4 := &Manager{
			filePath: filepath.Join(rigelDir, "history_empty"),
			entries:  []Entry{},
			maxSize:  100,
		}

		err := manager4.Add("")
		if err != nil {
			t.Errorf("Failed to handle empty command: %v", err)
		}

		if manager4.Size() != 0 {
			t.Errorf("Expected size to remain 0 for empty command, got %d", manager4.Size())
		}
	})
}

func TestGetRigelDir(t *testing.T) {
	dir, err := GetRigelDir()
	if err != nil {
		t.Fatalf("Failed to get Rigel dir: %v", err)
	}

	// Should end with .rigel
	if !strings.HasSuffix(dir, ".rigel") {
		t.Errorf("Expected dir to end with '.rigel', got %s", dir)
	}

	// Should contain home directory
	homeDir, _ := os.UserHomeDir()
	if !strings.HasPrefix(dir, homeDir) {
		t.Errorf("Expected dir to start with home directory %s, got %s", homeDir, dir)
	}
}
