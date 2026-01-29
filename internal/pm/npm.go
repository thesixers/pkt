package pm

import (
	"fmt"
	"os/exec"
)

// NPM implements the PackageManager interface for npm
type NPM struct{}

func (n *NPM) Name() string {
	return "npm"
}

func (n *NPM) IsAvailable() bool {
	return CheckAvailability("npm")
}

func (n *NPM) Add(workDir, pkg string, flags []string) error {
	args := []string{"install", pkg}
	args = append(args, flags...)

	cmd := exec.Command("npm", args...)
	cmd.Dir = workDir
	cmd.Stdout = nil
	cmd.Stderr = nil

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("npm install failed: %w", err)
	}

	return nil
}

func (n *NPM) AddMultiple(workDir string, packages []string, flags []string) error {
	args := []string{"install"}
	args = append(args, packages...)
	args = append(args, flags...)

	cmd := exec.Command("npm", args...)
	cmd.Dir = workDir
	cmd.Stdout = nil
	cmd.Stderr = nil

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("npm install failed: %w", err)
	}

	return nil
}

func (n *NPM) Remove(workDir, pkg string) error {
	cmd := exec.Command("npm", "uninstall", pkg)
	cmd.Dir = workDir
	cmd.Stdout = nil
	cmd.Stderr = nil

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("npm uninstall failed: %w", err)
	}

	return nil
}

func (n *NPM) Init(workDir string) error {
	cmd := exec.Command("npm", "init", "-y")
	cmd.Dir = workDir
	cmd.Stdout = nil
	cmd.Stderr = nil

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("npm init failed: %w", err)
	}

	return nil
}
