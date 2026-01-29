package pm

import (
	"fmt"
	"os/exec"
)

// PackageManager defines the interface for package manager operations
type PackageManager interface {
	Add(workDir, pkg string, flags []string) error
	Remove(workDir, pkg string) error
	Init(workDir string) error
	IsAvailable() bool
	Name() string
}

// Registry holds all available package managers
var registry = map[string]PackageManager{
	"pnpm": &PNPM{},
	"npm":  &NPM{},
	"bun":  &Bun{},
}

// GetPM retrieves a package manager by name
func GetPM(name string) (PackageManager, error) {
	pm, exists := registry[name]
	if !exists {
		return nil, fmt.Errorf("unknown package manager: %s", name)
	}
	return pm, nil
}

// CheckAvailability checks if a package manager is installed
func CheckAvailability(name string) bool {
	_, err := exec.LookPath(name)
	return err == nil
}

// ListAvailable returns a list of available package managers
func ListAvailable() []string {
	var available []string
	for name := range registry {
		if CheckAvailability(name) {
			available = append(available, name)
		}
	}
	return available
}

// InstallPnpm installs pnpm globally using npm
func InstallPnpm() error {
	// Check if npm is available
	if !CheckAvailability("npm") {
		return fmt.Errorf("npm is required to install pnpm")
	}

	// Run npm install -g pnpm
	cmd := exec.Command("npm", "install", "-g", "pnpm")
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to install pnpm: %w\nOutput: %s", err, string(output))
	}

	return nil
}
