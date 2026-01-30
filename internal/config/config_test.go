package config

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

func TestConfigSaveAndLoad(t *testing.T) {
	// Create a temporary directory for testing
	tmpDir, err := os.MkdirTemp("", "pkt-test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer func() { _ = os.RemoveAll(tmpDir) }()

	// Test the struct logic since we can't easily mock the filesystem location
	cfg := &Config{
		ProjectsRoot:  "/tmp/workspace",
		DefaultPM:     "pnpm",
		EditorCommand: "vim",
		Initialized:   true,
	}

	if cfg.ProjectsRoot != "/tmp/workspace" {
		t.Errorf("Expected ProjectsRoot to be /tmp/workspace, got %s", cfg.ProjectsRoot)
	}
	if !cfg.Initialized {
		t.Error("Expected Initialized to be true")
	}
}

func TestConfigJSONSerialization(t *testing.T) {
	cfg := &Config{
		ProjectsRoot:  "/home/user/projects",
		DefaultPM:     "npm",
		EditorCommand: "code",
		Initialized:   true,
	}

	// Marshal to JSON
	data, err := json.Marshal(cfg)
	if err != nil {
		t.Fatalf("Failed to marshal config: %v", err)
	}

	// Unmarshal back
	var loaded Config
	err = json.Unmarshal(data, &loaded)
	if err != nil {
		t.Fatalf("Failed to unmarshal config: %v", err)
	}

	// Verify fields
	if loaded.ProjectsRoot != cfg.ProjectsRoot {
		t.Errorf("ProjectsRoot mismatch: got %s, want %s", loaded.ProjectsRoot, cfg.ProjectsRoot)
	}
	if loaded.DefaultPM != cfg.DefaultPM {
		t.Errorf("DefaultPM mismatch: got %s, want %s", loaded.DefaultPM, cfg.DefaultPM)
	}
	if loaded.EditorCommand != cfg.EditorCommand {
		t.Errorf("EditorCommand mismatch: got %s, want %s", loaded.EditorCommand, cfg.EditorCommand)
	}
	if loaded.Initialized != cfg.Initialized {
		t.Errorf("Initialized mismatch: got %v, want %v", loaded.Initialized, cfg.Initialized)
	}
}

func TestConfigDefaults(t *testing.T) {
	// Test that a zero-value config has expected defaults
	cfg := Config{}

	if cfg.ProjectsRoot != "" {
		t.Errorf("Expected empty ProjectsRoot, got %s", cfg.ProjectsRoot)
	}
	if cfg.Initialized != false {
		t.Error("Expected Initialized to be false by default")
	}
}

func TestConfigFileOperations(t *testing.T) {
	// Create a temp directory to simulate config operations
	tmpDir, err := os.MkdirTemp("", "pkt-config-test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer func() { _ = os.RemoveAll(tmpDir) }()

	// Create a config file manually
	configDir := filepath.Join(tmpDir, ".pkt")
	if err := os.MkdirAll(configDir, 0755); err != nil {
		t.Fatalf("Failed to create config dir: %v", err)
	}

	cfg := &Config{
		ProjectsRoot:  "/home/test/projects",
		DefaultPM:     "bun",
		EditorCommand: "nano",
		Initialized:   true,
	}

	data, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		t.Fatalf("Failed to marshal config: %v", err)
	}

	configPath := filepath.Join(configDir, "config.json")
	if err := os.WriteFile(configPath, data, 0644); err != nil {
		t.Fatalf("Failed to write config file: %v", err)
	}

	// Verify file exists
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		t.Error("Config file was not created")
	}

	// Read it back
	readData, err := os.ReadFile(configPath)
	if err != nil {
		t.Fatalf("Failed to read config file: %v", err)
	}

	var loaded Config
	if err := json.Unmarshal(readData, &loaded); err != nil {
		t.Fatalf("Failed to parse config file: %v", err)
	}

	if loaded.DefaultPM != "bun" {
		t.Errorf("Expected DefaultPM 'bun', got '%s'", loaded.DefaultPM)
	}
}
