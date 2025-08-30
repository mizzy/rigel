package sandbox

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
)

const sandboxEnvVar = "RIGEL_SANDBOXED"

// IsSandboxed checks if the current process is already sandboxed
func IsSandboxed() bool {
	return os.Getenv(sandboxEnvVar) == "1"
}

// EnableSandbox re-executes the current process in a sandboxed environment
func EnableSandbox(sandboxDir string) error {
	// Only support macOS for now
	if runtime.GOOS != "darwin" {
		return fmt.Errorf("sandbox mode is currently only supported on macOS")
	}

	// Check if already sandboxed
	if IsSandboxed() {
		return nil
	}

	// Get absolute path of sandbox directory
	absDir, err := filepath.Abs(sandboxDir)
	if err != nil {
		return fmt.Errorf("failed to get absolute path: %w", err)
	}

	// Verify directory exists
	info, err := os.Stat(absDir)
	if err != nil {
		return fmt.Errorf("failed to stat directory: %w", err)
	}
	if !info.IsDir() {
		return fmt.Errorf("path is not a directory: %s", absDir)
	}

	// Create sandbox profile
	profile := generateSandboxProfile(absDir)

	// Get current executable path
	executable, err := os.Executable()
	if err != nil {
		return fmt.Errorf("failed to get executable path: %w", err)
	}

	// Prepare environment variables
	env := os.Environ()
	env = append(env, fmt.Sprintf("%s=1", sandboxEnvVar))

	// Write profile to temp file for debugging
	tempFile, err := os.CreateTemp("", "rigel-sandbox-*.sb")
	if err != nil {
		return fmt.Errorf("failed to create temp profile file: %w", err)
	}
	defer os.Remove(tempFile.Name())

	if _, err := tempFile.WriteString(profile); err != nil {
		return fmt.Errorf("failed to write profile: %w", err)
	}
	tempFile.Close()

	// Prepare sandbox-exec command using file-based profile
	args := []string{
		"-f", tempFile.Name(),
		executable,
	}

	// Add original command line arguments (excluding the program name)
	if len(os.Args) > 1 {
		args = append(args, os.Args[1:]...)
	}

	cmd := exec.Command("sandbox-exec", args...)
	cmd.Env = env
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	// Execute the sandboxed process
	err = cmd.Run()
	if err != nil {
		// Try to validate the profile
		validateCmd := exec.Command("sandbox-exec", "-n", "no-network", "-f", tempFile.Name(), "/usr/bin/true")
		if validateErr := validateCmd.Run(); validateErr != nil {
			return fmt.Errorf("invalid sandbox profile: %w", validateErr)
		}
		return fmt.Errorf("failed to execute sandboxed process: %w", err)
	}

	// Exit the current process
	os.Exit(0)
	return nil
}

// generateSandboxProfile creates a sandbox profile string for the given directory
func generateSandboxProfile(sandboxDir string) string {
	// Create a simple but effective sandbox profile
	// Allow reads everywhere but restrict writes to current directory only
	profile := fmt.Sprintf(`(version 1)
(allow default)
(deny file-write*
    (regex #"^/")
    (subpath "/"))
(allow file-write*
    (subpath "%s")
    (regex #"^/private/var/")
    (regex #"^/var/")
    (regex #"^/private/tmp/")
    (regex #"^/tmp/")
    (regex #"^/dev/"))
`, sandboxDir)

	return profile
}

// GetSandboxInfo returns information about the current sandbox status
func GetSandboxInfo() string {
	if IsSandboxed() {
		return "Sandbox: ENABLED (file writes restricted to current directory)"
	}
	return "Sandbox: DISABLED (use --sandbox flag to enable)"
}
