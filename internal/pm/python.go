package pm

import (
	"os"
	"os/exec"
	"path/filepath"
)

// UV implements PackageManager for uv (fast Python package manager)
type UV struct{}

func (u *UV) Name() string {
	return "uv"
}

func (u *UV) Language() string {
	return "python"
}

func (u *UV) Add(workDir string, packages []string, dev bool) error {
	args := []string{"add"}
	if dev {
		args = append(args, "--dev")
	}
	args = append(args, packages...)
	return runCommand("uv", args, workDir)
}

func (u *UV) Remove(workDir string, packages []string) error {
	args := append([]string{"remove"}, packages...)
	return runCommand("uv", args, workDir)
}

func (u *UV) Install(workDir string) error {
	return runCommand("uv", []string{"sync"}, workDir)
}

func (u *UV) Init(workDir string) error {
	return runCommand("uv", []string{"init"}, workDir)
}

func (u *UV) IsAvailable() bool {
	_, err := exec.LookPath("uv")
	return err == nil
}

// Pip implements PackageManager for pip
type Pip struct{}

func (p *Pip) Name() string {
	return "pip"
}

func (p *Pip) Language() string {
	return "python"
}

func (p *Pip) Add(workDir string, packages []string, dev bool) error {
	args := append([]string{"install"}, packages...)
	if err := runCommand("pip", args, workDir); err != nil {
		return err
	}
	// Update requirements.txt
	return p.freezeRequirements(workDir, dev)
}

func (p *Pip) freezeRequirements(workDir string, dev bool) error {
	filename := "requirements.txt"
	if dev {
		filename = "requirements-dev.txt"
	}
	reqPath := filepath.Join(workDir, filename)

	cmd := exec.Command("pip", "freeze")
	cmd.Dir = workDir
	output, err := cmd.Output()
	if err != nil {
		return err
	}

	return os.WriteFile(reqPath, output, 0644)
}

func (p *Pip) Remove(workDir string, packages []string) error {
	args := append([]string{"uninstall", "-y"}, packages...)
	return runCommand("pip", args, workDir)
}

func (p *Pip) Install(workDir string) error {
	reqPath := filepath.Join(workDir, "requirements.txt")
	if _, err := os.Stat(reqPath); err == nil {
		return runCommand("pip", []string{"install", "-r", "requirements.txt"}, workDir)
	}
	return nil
}

func (p *Pip) Init(workDir string) error {
	// Create empty requirements.txt
	reqPath := filepath.Join(workDir, "requirements.txt")
	return os.WriteFile(reqPath, []byte("# Python dependencies\n"), 0644)
}

func (p *Pip) IsAvailable() bool {
	_, err := exec.LookPath("pip")
	return err == nil
}

// Poetry implements PackageManager for poetry
type Poetry struct{}

func (p *Poetry) Name() string {
	return "poetry"
}

func (p *Poetry) Language() string {
	return "python"
}

func (p *Poetry) Add(workDir string, packages []string, dev bool) error {
	args := []string{"add"}
	if dev {
		args = append(args, "--group", "dev")
	}
	args = append(args, packages...)
	return runCommand("poetry", args, workDir)
}

func (p *Poetry) Remove(workDir string, packages []string) error {
	args := append([]string{"remove"}, packages...)
	return runCommand("poetry", args, workDir)
}

func (p *Poetry) Install(workDir string) error {
	return runCommand("poetry", []string{"install"}, workDir)
}

func (p *Poetry) Init(workDir string) error {
	return runCommand("poetry", []string{"init", "-n"}, workDir)
}

func (p *Poetry) IsAvailable() bool {
	_, err := exec.LookPath("poetry")
	return err == nil
}
