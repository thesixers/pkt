package lang

import (
	"fmt"
	"strings"
)

// Language defines the interface for programming language support
type Language interface {
	// Name returns the internal name (e.g., "javascript", "python", "go", "rust")
	Name() string

	// DisplayName returns the human-readable name (e.g., "JavaScript", "Python", "Go", "Rust")
	DisplayName() string

	// DetectProject checks if a directory contains a project of this language
	DetectProject(dir string) bool

	// GetManifestFile returns the path to the manifest file if it exists
	GetManifestFile(dir string) string

	// GetPackageManagers returns the list of supported package managers for this language
	GetPackageManagers() []string

	// DefaultPackageManager returns the default package manager for this language
	DefaultPackageManager() string

	// DetectPackageManager detects which package manager is being used in a project
	DetectPackageManager(dir string) string
}

// Manifest represents a project manifest file (package.json, pyproject.toml, etc.)
type Manifest interface {
	// Name returns the project name
	Name() string

	// Dependencies returns production dependencies (name -> version)
	Dependencies() map[string]string

	// DevDependencies returns development dependencies (name -> version)
	DevDependencies() map[string]string
}

// Registry holds all supported languages
var registry = map[string]Language{
	"javascript": &JavaScript{},
	"python":     &Python{},
	"go":         &Go{},
	"rust":       &Rust{},
}

// shortNames maps short codes to full language names
var shortNames = map[string]string{
	"js":   "javascript",
	"ts":   "javascript", // TypeScript is also JavaScript ecosystem
	"node": "javascript",
	"py":   "python",
	"rs":   "rust",
}

// NormalizeName converts a short code to the full language name
// e.g., "js" -> "javascript", "py" -> "python"
func NormalizeName(name string) string {
	name = strings.ToLower(strings.TrimSpace(name))
	if full, exists := shortNames[name]; exists {
		return full
	}
	return name
}

// Get retrieves a language by name (supports short codes)
func Get(name string) (Language, error) {
	normalized := NormalizeName(name)
	lang, exists := registry[normalized]
	if !exists {
		return nil, fmt.Errorf("unsupported language: %s (use: js, py, go, rs)", name)
	}
	return lang, nil
}

// List returns all supported language names
func List() []string {
	return []string{"javascript", "python", "go", "rust"}
}

// ListShort returns short codes for languages
func ListShort() []string {
	return []string{"js", "py", "go", "rs"}
}

// ListDisplay returns all supported languages with display names
func ListDisplay() map[string]string {
	return map[string]string{
		"javascript": "JavaScript/Node.js",
		"python":     "Python",
		"go":         "Go",
		"rust":       "Rust",
	}
}

// ListShortDisplay returns short codes with display names
func ListShortDisplay() map[string]string {
	return map[string]string{
		"js": "JavaScript/Node.js",
		"py": "Python",
		"go": "Go",
		"rs": "Rust",
	}
}

// Detect automatically detects the language of a project in a directory
func Detect(dir string) (Language, error) {
	// Check in order of specificity
	for _, name := range []string{"rust", "go", "python", "javascript"} {
		lang := registry[name]
		if lang.DetectProject(dir) {
			return lang, nil
		}
	}
	return nil, fmt.Errorf("could not detect project language in %s", dir)
}
