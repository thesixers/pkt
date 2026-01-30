package lang

import (
	"os"
	"path/filepath"
)

// JavaScript implements the Language interface for JavaScript/Node.js projects
type JavaScript struct{}

func (j *JavaScript) Name() string {
	return "javascript"
}

func (j *JavaScript) DisplayName() string {
	return "JavaScript/Node.js"
}

func (j *JavaScript) DetectProject(dir string) bool {
	// Check for package.json
	pkgPath := filepath.Join(dir, "package.json")
	_, err := os.Stat(pkgPath)
	return err == nil
}

func (j *JavaScript) GetManifestFile(dir string) string {
	pkgPath := filepath.Join(dir, "package.json")
	if _, err := os.Stat(pkgPath); err == nil {
		return pkgPath
	}
	return ""
}

func (j *JavaScript) GetPackageManagers() []string {
	return []string{"pnpm", "npm", "bun"}
}

func (j *JavaScript) DefaultPackageManager() string {
	return "pnpm"
}

func (j *JavaScript) DetectPackageManager(dir string) string {
	// Check for lockfiles in order of preference
	lockfiles := map[string]string{
		"pnpm-lock.yaml":    "pnpm",
		"bun.lockb":         "bun",
		"package-lock.json": "npm",
	}

	for lockfile, pm := range lockfiles {
		if _, err := os.Stat(filepath.Join(dir, lockfile)); err == nil {
			return pm
		}
	}

	return j.DefaultPackageManager()
}
