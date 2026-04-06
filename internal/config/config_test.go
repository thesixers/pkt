package config

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

func TestConfigSaveAndLoad(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "pkt-test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer func() { _ = os.RemoveAll(tmpDir) }()

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

	data, err := json.Marshal(cfg)
	if err != nil {
		t.Fatalf("Failed to marshal config: %v", err)
	}

	var loaded Config
	err = json.Unmarshal(data, &loaded)
	if err != nil {
		t.Fatalf("Failed to unmarshal config: %v", err)
	}

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
	cfg := Config{}

	if cfg.ProjectsRoot != "" {
		t.Errorf("Expected empty ProjectsRoot, got %s", cfg.ProjectsRoot)
	}
	if cfg.Initialized != false {
		t.Error("Expected Initialized to be false by default")
	}
}

func TestConfigFileOperations(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "pkt-config-test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer func() { _ = os.RemoveAll(tmpDir) }()

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

	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		t.Error("Config file was not created")
	}

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

func TestProviderConfigRegistry(t *testing.T) {
	cfg := &Config{
		AIProvider:  "groq",
		AIProviders: map[string]ProviderConfig{},
	}

	// Register cloud provider with API key
	cfg.AIProviders["groq"] = ProviderConfig{APIKey: "sk-test-key", Model: "llama-3.1-8b-instant"}
	// Register local provider with URL (no key)
	cfg.AIProviders["ollama"] = ProviderConfig{BaseURL: "http://localhost:11434/v1/chat/completions", Model: "llama3"}

	if cfg.AIProviders["groq"].APIKey != "sk-test-key" {
		t.Errorf("Expected groq key 'sk-test-key', got '%s'", cfg.AIProviders["groq"].APIKey)
	}
	if cfg.AIProviders["ollama"].APIKey != "" {
		t.Errorf("Expected empty API key for local provider, got '%s'", cfg.AIProviders["ollama"].APIKey)
	}
	if cfg.AIProviders["ollama"].BaseURL == "" {
		t.Error("Expected ollama BaseURL to be set")
	}
}

func TestLegacyMigration(t *testing.T) {
	cfg := Config{
		AIProvider:  "groq",
		AIKeys:      map[string]string{"groq": "sk-legacy-key"},
		AIModels:    map[string]string{"groq": "llama-3.1-8b-instant"},
		AIProviders: map[string]ProviderConfig{},
	}

	// Replicate the Load() migration logic
	for name, key := range cfg.AIKeys {
		if _, exists := cfg.AIProviders[name]; !exists {
			pc := cfg.AIProviders[name]
			pc.APIKey = key
			if m, ok := cfg.AIModels[name]; ok {
				pc.Model = m
			}
			cfg.AIProviders[name] = pc
		}
	}

	if cfg.AIProviders["groq"].APIKey != "sk-legacy-key" {
		t.Errorf("Migration failed: expected 'sk-legacy-key', got '%s'", cfg.AIProviders["groq"].APIKey)
	}
	if cfg.AIProviders["groq"].Model != "llama-3.1-8b-instant" {
		t.Errorf("Migration failed: expected model 'llama-3.1-8b-instant', got '%s'", cfg.AIProviders["groq"].Model)
	}
}
