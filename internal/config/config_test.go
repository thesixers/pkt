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
		DefaultPMs:    map[string]string{"javascript": "pnpm", "python": "uv", "go": "go", "rust": "cargo"},
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
		DefaultPMs:    map[string]string{"javascript": "npm", "python": "uv", "go": "go", "rust": "cargo"},
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
	if loaded.DefaultPMForLanguage("javascript") != cfg.DefaultPMForLanguage("javascript") {
		t.Errorf("JavaScript default mismatch: got %s, want %s", loaded.DefaultPMForLanguage("javascript"), cfg.DefaultPMForLanguage("javascript"))
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
		DefaultPMs:    map[string]string{"javascript": "bun", "python": "uv", "go": "go", "rust": "cargo"},
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

	if loaded.DefaultPMForLanguage("javascript") != "bun" {
		t.Errorf("Expected JavaScript default 'bun', got '%s'", loaded.DefaultPMForLanguage("javascript"))
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

	tmpDir, err := os.MkdirTemp("", "pkt-test-migrate")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer func() { _ = os.RemoveAll(tmpDir) }()

	cfgPath := filepath.Join(tmpDir, "config.json")
	data, _ := json.Marshal(cfg)
	_ = os.WriteFile(cfgPath, data, 0644)

	t.Setenv("PKT_CONFIG_PATH", cfgPath)

	loaded, err := Load()
	if err != nil {
		t.Fatalf("Load failed: %v", err)
	}

	groq := loaded.AIProviders["groq"]
	if len(groq.APIKeys) != 1 || groq.APIKeys[0] != "sk-legacy-key" {
		t.Errorf("Migration failed: expected ['sk-legacy-key'], got %v", groq.APIKeys)
	}
	if groq.DefaultModel != "llama-3.1-8b-instant" {
		t.Errorf("Migration failed: expected default model 'llama-3.1-8b-instant', got '%s'", groq.DefaultModel)
	}
	if len(groq.Models) != 1 || groq.Models[0] != "llama-3.1-8b-instant" {
		t.Errorf("Migration failed: expected models ['llama-3.1-8b-instant'], got %v", groq.Models)
	}
}

func TestLegacyDefaultPMMigration(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.json")
	t.Setenv("PKT_CONFIG_PATH", configPath)

	legacy := []byte(`{"projects_root":"/tmp/projects","default_pm":"bun","initialized":true}`)
	if err := os.WriteFile(configPath, legacy, 0644); err != nil {
		t.Fatal(err)
	}

	cfg, err := Load()
	if err != nil {
		t.Fatal(err)
	}
	if cfg.DefaultPMForLanguage("javascript") != "bun" {
		t.Fatalf("expected legacy default_pm to migrate to JavaScript, got %q", cfg.DefaultPMForLanguage("javascript"))
	}
	if cfg.DefaultPMForLanguage("python") != "uv" || cfg.DefaultPMForLanguage("go") != "go" || cfg.DefaultPMForLanguage("rust") != "cargo" {
		t.Fatalf("expected defaults for every language, got %#v", cfg.DefaultPMs)
	}

	persisted, err := os.ReadFile(configPath)
	if err != nil {
		t.Fatal(err)
	}
	if string(persisted) == string(legacy) || string(persisted) == "" {
		t.Fatal("expected migrated configuration to be persisted")
	}
}

func TestMaskKey(t *testing.T) {
	if got := MaskKey("short"); got != "••••••••" {
		t.Errorf("Expected '••••••••', got '%s'", got)
	}
	if got := MaskKey("12345678"); got != "••••••••" {
		t.Errorf("Expected '••••••••', got '%s'", got)
	}
	if got := MaskKey("gsk-1234567890abcdef"); got != "gsk-••••cdef" {
		t.Errorf("Expected 'gsk-••••cdef', got '%s'", got)
	}
}

func TestAIProviderKeyModelOperations(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "pkt-test-ai-ops")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer func() { _ = os.RemoveAll(tmpDir) }()

	cfgPath := filepath.Join(tmpDir, "config.json")
	t.Setenv("PKT_CONFIG_PATH", cfgPath)

	cfg := &Config{
		AIProviders: make(map[string]ProviderConfig),
	}
	if err := Save(cfg); err != nil {
		t.Fatalf("Failed to save initial config: %v", err)
	}

	// 1. Adding key or model before provider is configured should fail with exact error message
	err = cfg.AddKey("groq", "gsk-key-1")
	if err == nil {
		t.Fatalf("Expected error adding key to non-existent provider")
	}
	expectedErr := "provider \"groq\" is not configured. Add it first with:\npkt change ai add provider groq"
	if err.Error() != expectedErr {
		t.Errorf("Expected error '%s', got '%s'", expectedErr, err.Error())
	}

	// 2. Add provider
	isNew, err := cfg.AddProvider("groq", "")
	if err != nil || !isNew {
		t.Fatalf("AddProvider failed: %v", err)
	}
	if cfg.AIProvider != "groq" {
		t.Errorf("First provider added should be default: got '%s'", cfg.AIProvider)
	}

	// 3. Add keys
	if err := cfg.AddKey("groq", "gsk-key-1"); err != nil {
		t.Fatalf("AddKey failed: %v", err)
	}
	if err := cfg.AddKey("groq", "gsk-key-2"); err != nil {
		t.Fatalf("AddKey 2 failed: %v", err)
	}
	if len(cfg.AIProviders["groq"].APIKeys) != 2 {
		t.Fatalf("Expected 2 keys, got %d", len(cfg.AIProviders["groq"].APIKeys))
	}

	// 4. Add models
	if err := cfg.AddModel("groq", "llama-3.3-70b-versatile"); err != nil {
		t.Fatalf("AddModel failed: %v", err)
	}
	if cfg.AIProviders["groq"].DefaultModel != "llama-3.3-70b-versatile" {
		t.Errorf("First model added should be default model: got '%s'", cfg.AIProviders["groq"].DefaultModel)
	}
	if err := cfg.AddModel("groq", "llama-3.1-8b-instant"); err != nil {
		t.Fatalf("AddModel 2 failed: %v", err)
	}
	// Default model should remain the first model
	if cfg.AIProviders["groq"].DefaultModel != "llama-3.3-70b-versatile" {
		t.Errorf("Default model changed unexpectedly: got '%s'", cfg.AIProviders["groq"].DefaultModel)
	}

	// 5. Set default model
	if err := cfg.SetDefaultModel("groq", "llama-3.1-8b-instant"); err != nil {
		t.Fatalf("SetDefaultModel failed: %v", err)
	}
	if cfg.AIProviders["groq"].DefaultModel != "llama-3.1-8b-instant" {
		t.Errorf("Expected default model 'llama-3.1-8b-instant', got '%s'", cfg.AIProviders["groq"].DefaultModel)
	}

	// 6. Remove key
	if err := cfg.RemoveKey("groq", "gsk-key-1"); err != nil {
		t.Fatalf("RemoveKey failed: %v", err)
	}
	if len(cfg.AIProviders["groq"].APIKeys) != 1 || cfg.AIProviders["groq"].APIKeys[0] != "gsk-key-2" {
		t.Errorf("Expected remaining key 'gsk-key-2', got %v", cfg.AIProviders["groq"].APIKeys)
	}

	// 7. Remove model
	if err := cfg.RemoveModel("groq", "llama-3.1-8b-instant"); err != nil {
		t.Fatalf("RemoveModel failed: %v", err)
	}
	// Default model should fall back to first remaining model
	if cfg.AIProviders["groq"].DefaultModel != "llama-3.3-70b-versatile" {
		t.Errorf("Expected default model to fall back, got '%s'", cfg.AIProviders["groq"].DefaultModel)
	}

	// 8. Remove provider
	if err := cfg.RemoveProvider("groq"); err != nil {
		t.Fatalf("RemoveProvider failed: %v", err)
	}
	if cfg.AIProvider != "" {
		t.Errorf("Expected empty active AIProvider, got '%s'", cfg.AIProvider)
	}
}
