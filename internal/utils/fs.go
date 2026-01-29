package utils

import (
	"fmt"
	"os"
	"path/filepath"
)

// CreateProjectDir creates a project directory and returns its absolute path
func CreateProjectDir(root, name string) (string, error) {
	// Expand home directory if needed
	if root[:2] == "~/" {
		home, err := os.UserHomeDir()
		if err != nil {
			return "", fmt.Errorf("failed to get home directory: %w", err)
		}
		root = filepath.Join(home, root[2:])
	}

	// Create absolute path
	projectPath := filepath.Join(root, name)

	// Check if directory already exists
	if _, err := os.Stat(projectPath); err == nil {
		return "", fmt.Errorf("directory already exists: %s", projectPath)
	}

	// Create directory
	if err := os.MkdirAll(projectPath, 0755); err != nil {
		return "", fmt.Errorf("failed to create directory: %w", err)
	}

	// Return absolute path
	absPath, err := filepath.Abs(projectPath)
	if err != nil {
		return "", fmt.Errorf("failed to get absolute path: %w", err)
	}

	return absPath, nil
}

// DeleteProjectDir removes a project directory
func DeleteProjectDir(path string) error {
	// Check if directory exists
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return fmt.Errorf("directory does not exist: %s", path)
	}

	// Remove directory and all contents
	if err := os.RemoveAll(path); err != nil {
		return fmt.Errorf("failed to delete directory: %w", err)
	}

	return nil
}

// IsInsideProject checks if the current directory is inside a project
func IsInsideProject(db interface{}, currentPath string) (bool, string, error) {
	// This is a placeholder - will be implemented with db access
	// For now, just return false
	return false, "", nil
}

// ExpandPath expands ~ in paths to home directory
func ExpandPath(path string) (string, error) {
	if len(path) == 0 {
		return path, nil
	}

	if path[:2] == "~/" {
		home, err := os.UserHomeDir()
		if err != nil {
			return "", fmt.Errorf("failed to get home directory: %w", err)
		}
		return filepath.Join(home, path[2:]), nil
	}

	return path, nil
}
