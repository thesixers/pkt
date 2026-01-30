package pm

import (
	"fmt"
	"os"
	"os/exec"
)

// PackageManager defines the interface for package manager operations
type PackageManager interface {
	// Name returns the package manager name
	Name() string

	// Language returns the language this PM supports
	Language() string

	// Add adds packages to a project
	Add(workDir string, packages []string, dev bool) error

	// Remove removes packages from a project
	Remove(workDir string, packages []string) error

	// Install installs all dependencies
	Install(workDir string) error

	// Init initializes a new project
	Init(workDir string) error

	// Run runs a script or command
	Run(workDir string, script string, args []string) error

	// Update updates packages (empty slice = update all)
	Update(workDir string, packages []string) error

	// IsAvailable checks if this package manager is installed
	IsAvailable() bool
}

// OutdatedDep represents an outdated dependency
type OutdatedDep struct {
	Name    string
	Current string
	Latest  string
	DepType string
}

// Registry holds all available package managers by language
var registry = map[string]map[string]PackageManager{
	"javascript": {
		"pnpm": &PNPM{},
		"npm":  &NPM{},
		"bun":  &Bun{},
	},
	"python": {
		"uv":     &UV{},
		"pip":    &Pip{},
		"poetry": &Poetry{},
	},
	"go": {
		"go": &GoMod{},
	},
	"rust": {
		"cargo": &Cargo{},
	},
}

// Get retrieves a package manager by language and name
func Get(language, name string) (PackageManager, error) {
	langPMs, exists := registry[language]
	if !exists {
		return nil, fmt.Errorf("unsupported language: %s", language)
	}
	pm, exists := langPMs[name]
	if !exists {
		return nil, fmt.Errorf("unknown package manager '%s' for language '%s'", name, language)
	}
	return pm, nil
}

// GetPM retrieves a package manager by name (searches all languages)
// Deprecated: Use Get(language, name) instead
func GetPM(name string) (PackageManager, error) {
	for _, langPMs := range registry {
		if pm, exists := langPMs[name]; exists {
			return pm, nil
		}
	}
	return nil, fmt.Errorf("unknown package manager: %s", name)
}

// ListForLanguage returns available package managers for a language
func ListForLanguage(language string) []string {
	langPMs, exists := registry[language]
	if !exists {
		return nil
	}
	var available []string
	for name, pm := range langPMs {
		if pm.IsAvailable() {
			available = append(available, name)
		}
	}
	return available
}

// ListAvailable returns a list of all available package managers
func ListAvailable() []string {
	var available []string
	for _, langPMs := range registry {
		for name, pm := range langPMs {
			if pm.IsAvailable() {
				available = append(available, name)
			}
		}
	}
	return available
}

// CheckAvailability checks if a package manager is installed
func CheckAvailability(name string) bool {
	_, err := exec.LookPath(name)
	return err == nil
}

// InstallPnpm installs pnpm globally using npm
func InstallPnpm() error {
	if !CheckAvailability("npm") {
		return fmt.Errorf("npm is required to install pnpm")
	}

	cmd := exec.Command("npm", "install", "-g", "pnpm")
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to install pnpm: %w\nOutput: %s", err, string(output))
	}

	return nil
}

// runCommand is a helper to run commands
func runCommand(name string, args []string, workDir string) error {
	cmd := exec.Command(name, args...)
	cmd.Dir = workDir
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("command failed: %w\nOutput: %s", err, string(output))
	}
	return nil
}

// runCommandInteractive runs a command with stdout/stderr connected to terminal
func runCommandInteractive(name string, args []string, workDir string) error {
	cmd := exec.Command(name, args...)
	cmd.Dir = workDir
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin
	return cmd.Run()
}
