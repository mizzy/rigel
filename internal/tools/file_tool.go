package tools

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

type FileTool struct {
	BaseTool
}

func NewFileTool() *FileTool {
	return &FileTool{
		BaseTool: BaseTool{
			name:        "file_operations",
			description: "Perform file operations like read, write, and list files",
		},
	}
}

func (f *FileTool) Execute(ctx context.Context, input string) (string, error) {
	parts := strings.Fields(input)
	if len(parts) == 0 {
		return "", fmt.Errorf("no operation specified")
	}

	operation := parts[0]
	args := parts[1:]

	switch operation {
	case "read":
		if len(args) == 0 {
			return "", fmt.Errorf("no file path specified")
		}
		return f.readFile(args[0])
	case "write":
		if len(args) < 2 {
			return "", fmt.Errorf("usage: write <path> <content>")
		}
		content := strings.Join(args[1:], " ")
		return f.writeFile(args[0], content)
	case "list":
		path := "."
		if len(args) > 0 {
			path = args[0]
		}
		return f.listFiles(path)
	case "exists":
		if len(args) == 0 {
			return "", fmt.Errorf("no file path specified")
		}
		return f.checkExists(args[0])
	case "delete":
		if len(args) == 0 {
			return "", fmt.Errorf("no file path specified")
		}
		return f.deleteFile(args[0])
	default:
		return "", fmt.Errorf("unknown operation: %s", operation)
	}
}

func (f *FileTool) readFile(path string) (string, error) {
	absPath, err := filepath.Abs(path)
	if err != nil {
		return "", err
	}

	content, err := os.ReadFile(absPath)
	if err != nil {
		return "", fmt.Errorf("failed to read file: %w", err)
	}

	return string(content), nil
}

func (f *FileTool) writeFile(path, content string) (string, error) {
	absPath, err := filepath.Abs(path)
	if err != nil {
		return "", err
	}

	dir := filepath.Dir(absPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return "", fmt.Errorf("failed to create directory: %w", err)
	}

	if err := os.WriteFile(absPath, []byte(content), 0644); err != nil {
		return "", fmt.Errorf("failed to write file: %w", err)
	}

	return fmt.Sprintf("File written successfully: %s", absPath), nil
}

func (f *FileTool) listFiles(path string) (string, error) {
	absPath, err := filepath.Abs(path)
	if err != nil {
		return "", err
	}

	entries, err := os.ReadDir(absPath)
	if err != nil {
		return "", fmt.Errorf("failed to list files: %w", err)
	}

	var result []string
	for _, entry := range entries {
		info, err := entry.Info()
		if err != nil {
			continue
		}

		fileType := "file"
		if entry.IsDir() {
			fileType = "dir"
		}

		result = append(result, fmt.Sprintf("[%s] %s (size: %d bytes)",
			fileType, entry.Name(), info.Size()))
	}

	return strings.Join(result, "\n"), nil
}

func (f *FileTool) checkExists(path string) (string, error) {
	absPath, err := filepath.Abs(path)
	if err != nil {
		return "", err
	}

	info, err := os.Stat(absPath)
	if os.IsNotExist(err) {
		return "File does not exist", nil
	}
	if err != nil {
		return "", fmt.Errorf("failed to check file: %w", err)
	}

	fileType := "file"
	if info.IsDir() {
		fileType = "directory"
	}

	return fmt.Sprintf("Exists: %s (type: %s, size: %d bytes)",
		absPath, fileType, info.Size()), nil
}

func (f *FileTool) deleteFile(path string) (string, error) {
	absPath, err := filepath.Abs(path)
	if err != nil {
		return "", err
	}

	if err := os.Remove(absPath); err != nil {
		return "", fmt.Errorf("failed to delete file: %w", err)
	}

	return fmt.Sprintf("File deleted successfully: %s", absPath), nil
}
