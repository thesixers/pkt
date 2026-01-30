package lang

import (
	"os"
	"path/filepath"
)

// Python implements the Language interface for Python projects
type Python struct{}

func (p *Python) Name() string {
	return "python"
}

func (p *Python) DisplayName() string {
	return "Python"
}

func (p *Python) DetectProject(dir string) bool {
	// Check for Python project files
	files := []string{
		"pyproject.toml",
		"requirements.txt",
		"setup.py",
		"Pipfile",
	}

	for _, file := range files {
		if _, err := os.Stat(filepath.Join(dir, file)); err == nil {
			return true
		}
	}
	return false
}

func (p *Python) GetManifestFile(dir string) string {
	// Prefer pyproject.toml, then requirements.txt
	files := []string{"pyproject.toml", "requirements.txt", "setup.py"}
	for _, file := range files {
		path := filepath.Join(dir, file)
		if _, err := os.Stat(path); err == nil {
			return path
		}
	}
	return ""
}

func (p *Python) GetPackageManagers() []string {
	return []string{"uv", "pip", "poetry"}
}

func (p *Python) DefaultPackageManager() string {
	return "uv"
}

func (p *Python) DetectPackageManager(dir string) string {
	// Check for lockfiles/config files in order of preference
	indicators := map[string]string{
		"uv.lock":      "uv",
		"poetry.lock":  "poetry",
		"Pipfile.lock": "pip",
	}

	for indicator, pm := range indicators {
		if _, err := os.Stat(filepath.Join(dir, indicator)); err == nil {
			return pm
		}
	}

	// Check pyproject.toml for tool-specific config
	pyproject := filepath.Join(dir, "pyproject.toml")
	if _, err := os.Stat(pyproject); err == nil {
		// If pyproject.toml exists, default to uv (modern)
		return "uv"
	}

	// If only requirements.txt exists, use pip
	if _, err := os.Stat(filepath.Join(dir, "requirements.txt")); err == nil {
		return "pip"
	}

	return p.DefaultPackageManager()
}
