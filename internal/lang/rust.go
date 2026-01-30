package lang

import (
	"os"
	"path/filepath"
)

// Rust implements the Language interface for Rust projects
type Rust struct{}

func (r *Rust) Name() string {
	return "rust"
}

func (r *Rust) DisplayName() string {
	return "Rust"
}

func (r *Rust) DetectProject(dir string) bool {
	// Check for Cargo.toml
	cargoToml := filepath.Join(dir, "Cargo.toml")
	_, err := os.Stat(cargoToml)
	return err == nil
}

func (r *Rust) GetManifestFile(dir string) string {
	cargoToml := filepath.Join(dir, "Cargo.toml")
	if _, err := os.Stat(cargoToml); err == nil {
		return cargoToml
	}
	return ""
}

func (r *Rust) GetPackageManagers() []string {
	return []string{"cargo"}
}

func (r *Rust) DefaultPackageManager() string {
	return "cargo"
}

func (r *Rust) DetectPackageManager(dir string) string {
	// Rust only has cargo as its package manager
	return "cargo"
}
