package lang

import (
	"os"
	"path/filepath"
)

// Go implements the Language interface for Go projects
type Go struct{}

func (g *Go) Name() string {
	return "go"
}

func (g *Go) DisplayName() string {
	return "Go"
}

func (g *Go) DetectProject(dir string) bool {
	// Check for go.mod
	goMod := filepath.Join(dir, "go.mod")
	_, err := os.Stat(goMod)
	return err == nil
}

func (g *Go) GetManifestFile(dir string) string {
	goMod := filepath.Join(dir, "go.mod")
	if _, err := os.Stat(goMod); err == nil {
		return goMod
	}
	return ""
}

func (g *Go) GetPackageManagers() []string {
	return []string{"go"}
}

func (g *Go) DefaultPackageManager() string {
	return "go"
}

func (g *Go) DetectPackageManager(dir string) string {
	// Go only has one package manager
	return "go"
}
