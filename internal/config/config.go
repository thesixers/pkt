package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

// ProviderConfig holds the settings for one AI provider.
// Cloud providers set APIKey; local providers set BaseURL instead.
type ProviderConfig struct {
	APIKey  string `json:"api_key,omitempty"`  // empty for local providers
	BaseURL string `json:"base_url,omitempty"` // override or localhost URL
	Model   string `json:"model,omitempty"`    // pinned model name
}

// Config represents the pkt configuration
type Config struct {
	ProjectsRoot  string                    `json:"projects_root"`
	DefaultPM     string                    `json:"default_pm"`
	EditorCommand string                    `json:"editor"`
	Initialized   bool                      `json:"initialized"`
	AIProvider    string                    `json:"ai_provider,omitempty"`
	AIProviders   map[string]ProviderConfig `json:"ai_providers,omitempty"`

	// Legacy fields — kept for migration only, do not use directly
	AIKey    string            `json:"ai_key,omitempty"`
	AIKeys   map[string]string `json:"ai_keys,omitempty"`
	AIModels map[string]string `json:"ai_models,omitempty"`
}

// configPath returns the path to the config file
func configPath() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("failed to get home directory: %w", err)
	}
	return filepath.Join(home, ".pkt", "config.json"), nil
}

// Exists checks if the config file exists
func Exists() (bool, error) {
	path, err := configPath()
	if err != nil {
		return false, err
	}

	_, err = os.Stat(path)
	if os.IsNotExist(err) {
		return false, nil
	}
	if err != nil {
		return false, err
	}

	return true, nil
}

// Load reads the configuration from ~/.pkt/config.json
func Load() (*Config, error) {
	path, err := configPath()
	if err != nil {
		return nil, err
	}

	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, fmt.Errorf("config file not found, run 'pkt start' first")
		}
		return nil, fmt.Errorf("failed to read config: %w", err)
	}

	var cfg Config
	if err := json.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("failed to parse config: %w", err)
	}

	if cfg.AIProviders == nil {
		cfg.AIProviders = make(map[string]ProviderConfig)
	}

	// Migrate legacy flat maps into the new registry
	migrated := false
	for name, key := range cfg.AIKeys {
		if _, exists := cfg.AIProviders[name]; !exists {
			pc := cfg.AIProviders[name]
			pc.APIKey = key
			if m, ok := cfg.AIModels[name]; ok {
				pc.Model = m
			}
			cfg.AIProviders[name] = pc
			migrated = true
		}
	}
	// Migrate the single legacy ai_key
	if cfg.AIKey != "" && cfg.AIProvider != "" {
		if p, exists := cfg.AIProviders[cfg.AIProvider]; !exists || p.APIKey == "" {
			pc := cfg.AIProviders[cfg.AIProvider]
			pc.APIKey = cfg.AIKey
			cfg.AIProviders[cfg.AIProvider] = pc
			migrated = true
		}
	}
	if migrated {
		cfg.AIKeys = nil
		cfg.AIModels = nil
		cfg.AIKey = ""
		_ = Save(&cfg)
	}

	return &cfg, nil
}

// Save writes the configuration to ~/.pkt/config.json
func Save(cfg *Config) error {
	path, err := configPath()
	if err != nil {
		return err
	}

	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}

	data, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	if err := os.WriteFile(path, data, 0644); err != nil {
		return fmt.Errorf("failed to write config: %w", err)
	}

	return nil
}
