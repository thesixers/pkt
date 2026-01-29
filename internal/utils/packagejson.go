package utils

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/genesix/pkt/internal/db"
)

// PackageJSON represents the structure of package.json
type PackageJSON struct {
	Name            string            `json:"name"`
	Version         string            `json:"version"`
	Dependencies    map[string]string `json:"dependencies,omitempty"`
	DevDependencies map[string]string `json:"devDependencies,omitempty"`
	Scripts         map[string]string `json:"scripts,omitempty"`
}

// ParsePackageJSON reads and parses package.json from a project path
func ParsePackageJSON(projectPath string) (map[string]*db.Dependency, error) {
	pkgJSONPath := filepath.Join(projectPath, "package.json")

	// Read file
	data, err := os.ReadFile(pkgJSONPath)
	if err != nil {
		if os.IsNotExist(err) {
			return make(map[string]*db.Dependency), nil
		}
		return nil, fmt.Errorf("failed to read package.json: %w", err)
	}

	// Parse JSON
	var pkg PackageJSON
	if err := json.Unmarshal(data, &pkg); err != nil {
		return nil, fmt.Errorf("failed to parse package.json: %w", err)
	}

	// Convert to dependency map
	deps := make(map[string]*db.Dependency)

	for name, version := range pkg.Dependencies {
		deps[name] = &db.Dependency{
			Name:    name,
			Version: version,
			DepType: "prod",
		}
	}

	for name, version := range pkg.DevDependencies {
		deps[name] = &db.Dependency{
			Name:    name,
			Version: version,
			DepType: "dev",
		}
	}

	return deps, nil
}

// CreatePackageJSON creates a minimal package.json file
func CreatePackageJSON(projectPath, projectName string) error {
	pkgJSONPath := filepath.Join(projectPath, "package.json")

	// Check if file already exists
	if _, err := os.Stat(pkgJSONPath); err == nil {
		return nil // Already exists, no need to create
	}

	// Create minimal package.json
	pkg := PackageJSON{
		Name:    projectName,
		Version: "1.0.0",
	}

	// Marshal to JSON
	data, err := json.MarshalIndent(pkg, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal package.json: %w", err)
	}

	// Write to file
	if err := os.WriteFile(pkgJSONPath, data, 0644); err != nil {
		return fmt.Errorf("failed to write package.json: %w", err)
	}

	return nil
}

// RewriteScripts updates package.json scripts to use the correct package manager
func RewriteScripts(projectPath, pm string) error {
	pkgJSONPath := filepath.Join(projectPath, "package.json")

	// Read existing package.json
	data, err := os.ReadFile(pkgJSONPath)
	if err != nil {
		return fmt.Errorf("failed to read package.json: %w", err)
	}

	// Parse JSON
	var pkg PackageJSON
	if err := json.Unmarshal(data, &pkg); err != nil {
		return fmt.Errorf("failed to parse package.json: %w", err)
	}

	// If no scripts, nothing to rewrite
	if len(pkg.Scripts) == 0 {
		return nil
	}

	// Map of package managers to replace
	pms := []string{"npm", "pnpm", "bun", "yarn"}

	// Rewrite scripts
	for key, script := range pkg.Scripts {
		for _, oldPM := range pms {
			if oldPM != pm {
				script = strings.ReplaceAll(script, oldPM+" ", pm+" ")
				script = strings.ReplaceAll(script, oldPM+"\t", pm+"\t")
			}
		}
		pkg.Scripts[key] = script
	}

	// Marshal back to JSON
	data, err = json.MarshalIndent(pkg, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal package.json: %w", err)
	}

	// Write back to file
	if err := os.WriteFile(pkgJSONPath, data, 0644); err != nil {
		return fmt.Errorf("failed to write package.json: %w", err)
	}

	return nil
}
