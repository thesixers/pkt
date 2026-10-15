package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// ProviderConfig holds the settings for one AI provider.
// Cloud providers set APIKey; local providers set BaseURL instead.
type ProviderConfig struct {
	APIKeys      []string `json:"api_keys,omitempty"`
	Models       []string `json:"models,omitempty"`
	DefaultModel string   `json:"default_model,omitempty"`
	BaseURL      string   `json:"base_url,omitempty"`

	// Legacy fields are read from existing configuration and migrated on Load.
	APIKey string `json:"api_key,omitempty"`
	Model  string `json:"model,omitempty"`
}

// Config represents the pkt configuration
type Config struct {
	ProjectsRoot  string                    `json:"projects_root"`
	DefaultPMs    map[string]string         `json:"default_pms,omitempty"`
	EditorCommand string                    `json:"editor"`
	Initialized   bool                      `json:"initialized"`
	AIProvider    string                    `json:"ai_provider,omitempty"`
	AIProviders   map[string]ProviderConfig `json:"ai_providers,omitempty"`
	ProviderOrder []string                  `json:"provider_order,omitempty"`

	// Legacy fields — kept for migration only, do not use directly
	AIKey     string            `json:"ai_key,omitempty"`
	AIKeys    map[string]string `json:"ai_keys,omitempty"`
	AIModels  map[string]string `json:"ai_models,omitempty"`
	DefaultPM string            `json:"default_pm,omitempty"`
}

// configPath returns the path to the config file
func configPath() (string, error) {
	if custom := os.Getenv("PKT_CONFIG_PATH"); custom != "" {
		return custom, nil
	}
	if customDir := os.Getenv("PKT_CONFIG_DIR"); customDir != "" {
		return filepath.Join(customDir, "config.json"), nil
	}
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
	migrated := false

	// Migrate the former single package-manager default into per-language
	// defaults. The previous value was only meaningful for JavaScript.
	if cfg.DefaultPMs == nil {
		cfg.DefaultPMs = make(map[string]string)
	}
	if cfg.DefaultPM != "" && cfg.DefaultPMs["javascript"] == "" {
		cfg.DefaultPMs["javascript"] = cfg.DefaultPM
		migrated = true
	}
	for language, packageManager := range defaultPackageManagers {
		if cfg.DefaultPMs[language] == "" {
			cfg.DefaultPMs[language] = packageManager
			migrated = true
		}
	}
	if cfg.DefaultPM != "" {
		cfg.DefaultPM = ""
		migrated = true
	}

	// Migrate legacy flat maps and single-value provider settings into the
	// multi-key, multi-model registry.
	for name, key := range cfg.AIKeys {
		pc := cfg.AIProviders[name]
		if addUnique(&pc.APIKeys, key) {
			migrated = true
		}
		if m, ok := cfg.AIModels[name]; ok && addUnique(&pc.Models, m) {
			migrated = true
		}
		cfg.AIProviders[name] = pc
	}
	// Migrate the single legacy ai_key.
	if cfg.AIKey != "" && cfg.AIProvider != "" {
		pc := cfg.AIProviders[cfg.AIProvider]
		if addUnique(&pc.APIKeys, cfg.AIKey) {
			migrated = true
		}
		cfg.AIProviders[cfg.AIProvider] = pc
	}
	for name, pc := range cfg.AIProviders {
		if pc.APIKey != "" {
			addUnique(&pc.APIKeys, pc.APIKey)
			pc.APIKey = ""
			migrated = true
		}
		if pc.Model != "" {
			addUnique(&pc.Models, pc.Model)
			if pc.DefaultModel == "" {
				pc.DefaultModel = pc.Model
			}
			pc.Model = ""
			migrated = true
		}
		if pc.DefaultModel == "" && len(pc.Models) > 0 {
			pc.DefaultModel = pc.Models[0]
			migrated = true
		}
		cfg.AIProviders[name] = pc
		if !sliceContains(cfg.ProviderOrder, name) {
			cfg.ProviderOrder = append(cfg.ProviderOrder, name)
			migrated = true
		}
	}
	if cfg.AIProvider == "" && len(cfg.ProviderOrder) > 0 {
		cfg.AIProvider = cfg.ProviderOrder[0]
		migrated = true
	}
	if migrated {
		cfg.AIKeys = nil
		cfg.AIModels = nil
		cfg.AIKey = ""
		_ = Save(&cfg)
	}

	return &cfg, nil
}

var defaultPackageManagers = map[string]string{
	"javascript": "pnpm",
	"python":     "uv",
	"go":         "go",
	"rust":       "cargo",
}

// DefaultPMForLanguage returns the configured package manager for a language.
func (c *Config) DefaultPMForLanguage(language string) string {
	if c.DefaultPMs != nil && c.DefaultPMs[language] != "" {
		return c.DefaultPMs[language]
	}
	return defaultPackageManagers[language]
}

// SetDefaultPMForLanguage sets the package manager used for newly created
// projects of a language.
func (c *Config) SetDefaultPMForLanguage(language, packageManager string) {
	if c.DefaultPMs == nil {
		c.DefaultPMs = make(map[string]string)
	}
	c.DefaultPMs[language] = packageManager
}

func addUnique(items *[]string, value string) bool {
	if value == "" || sliceContains(*items, value) {
		return false
	}
	*items = append(*items, value)
	return true
}

func sliceContains(items []string, value string) bool {
	for _, item := range items {
		if item == value {
			return true
		}
	}
	return false
}

func removeSliceItem(items []string, value string) []string {
	var result []string
	for _, item := range items {
		if item != value {
			result = append(result, item)
		}
	}
	return result
}

// MaskKey returns a masked representation of an API key for display.
func MaskKey(key string) string {
	if len(key) <= 8 {
		return "••••••••"
	}
	return key[:4] + "••••" + key[len(key)-4:]
}

// AddProvider registers a provider or updates its base URL.
// The first provider added becomes the active default provider.
func (c *Config) AddProvider(name, baseURL string) (bool, error) {
	name = strings.TrimSpace(name)
	if name == "" {
		return false, fmt.Errorf("provider name cannot be empty")
	}
	if c.AIProviders == nil {
		c.AIProviders = make(map[string]ProviderConfig)
	}

	pc, exists := c.AIProviders[name]
	if !exists {
		pc = ProviderConfig{}
	}
	if baseURL != "" {
		pc.BaseURL = baseURL
	}
	c.AIProviders[name] = pc

	if !sliceContains(c.ProviderOrder, name) {
		c.ProviderOrder = append(c.ProviderOrder, name)
	}
	if c.AIProvider == "" {
		c.AIProvider = name
	}
	return !exists, Save(c)
}

// RemoveProvider deletes an AI provider and updates default provider if needed.
func (c *Config) RemoveProvider(name string) error {
	name = strings.TrimSpace(name)
	if _, exists := c.AIProviders[name]; !exists {
		return fmt.Errorf("provider '%s' is not configured", name)
	}

	delete(c.AIProviders, name)
	c.ProviderOrder = removeSliceItem(c.ProviderOrder, name)

	if c.AIProvider == name {
		if len(c.ProviderOrder) > 0 {
			c.AIProvider = c.ProviderOrder[0]
		} else {
			c.AIProvider = ""
		}
	}
	return Save(c)
}

// AddKey adds an API key to an existing provider.
func (c *Config) AddKey(provider, key string) error {
	provider = strings.TrimSpace(provider)
	key = strings.TrimSpace(key)
	if key == "" {
		return fmt.Errorf("API key cannot be empty")
	}

	pc, exists := c.AIProviders[provider]
	if !exists {
		return fmt.Errorf("provider \"%s\" is not configured. Add it first with:\npkt change ai add provider %s", provider, provider)
	}

	if addUnique(&pc.APIKeys, key) {
		c.AIProviders[provider] = pc
		return Save(c)
	}
	return nil
}

// RemoveKey removes an API key from an existing provider.
func (c *Config) RemoveKey(provider, key string) error {
	provider = strings.TrimSpace(provider)
	key = strings.TrimSpace(key)

	pc, exists := c.AIProviders[provider]
	if !exists {
		return fmt.Errorf("provider \"%s\" is not configured. Add it first with:\npkt change ai add provider %s", provider, provider)
	}

	origLen := len(pc.APIKeys)
	pc.APIKeys = removeSliceItem(pc.APIKeys, key)
	if len(pc.APIKeys) == origLen {
		return fmt.Errorf("key not found for provider '%s'", provider)
	}

	c.AIProviders[provider] = pc
	return Save(c)
}

// AddModel adds a model to an existing provider.
// The first model added becomes the default model.
func (c *Config) AddModel(provider, model string) error {
	provider = strings.TrimSpace(provider)
	model = strings.TrimSpace(model)
	if model == "" {
		return fmt.Errorf("model name cannot be empty")
	}

	pc, exists := c.AIProviders[provider]
	if !exists {
		return fmt.Errorf("provider \"%s\" is not configured. Add it first with:\npkt change ai add provider %s", provider, provider)
	}

	addUnique(&pc.Models, model)
	if pc.DefaultModel == "" {
		pc.DefaultModel = model
	}
	c.AIProviders[provider] = pc
	return Save(c)
}

// RemoveModel removes a model from an existing provider.
func (c *Config) RemoveModel(provider, model string) error {
	provider = strings.TrimSpace(provider)
	model = strings.TrimSpace(model)

	pc, exists := c.AIProviders[provider]
	if !exists {
		return fmt.Errorf("provider \"%s\" is not configured. Add it first with:\npkt change ai add provider %s", provider, provider)
	}

	origLen := len(pc.Models)
	pc.Models = removeSliceItem(pc.Models, model)
	if len(pc.Models) == origLen {
		return fmt.Errorf("model '%s' not found for provider '%s'", model, provider)
	}

	if pc.DefaultModel == model {
		if len(pc.Models) > 0 {
			pc.DefaultModel = pc.Models[0]
		} else {
			pc.DefaultModel = ""
		}
	}
	c.AIProviders[provider] = pc
	return Save(c)
}

// SetDefaultModel sets the default model for an existing provider.
func (c *Config) SetDefaultModel(provider, model string) error {
	provider = strings.TrimSpace(provider)
	model = strings.TrimSpace(model)
	if model == "" {
		return fmt.Errorf("model name cannot be empty")
	}

	pc, exists := c.AIProviders[provider]
	if !exists {
		return fmt.Errorf("provider \"%s\" is not configured. Add it first with:\npkt change ai add provider %s", provider, provider)
	}

	addUnique(&pc.Models, model)
	pc.DefaultModel = model
	c.AIProviders[provider] = pc
	return Save(c)
}

// SetActiveProvider sets the active AI provider.
func (c *Config) SetActiveProvider(provider string) error {
	provider = strings.TrimSpace(provider)
	if _, exists := c.AIProviders[provider]; !exists {
		return fmt.Errorf("provider \"%s\" is not configured. Add it first with:\npkt change ai add provider %s", provider, provider)
	}
	c.AIProvider = provider
	return Save(c)
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
