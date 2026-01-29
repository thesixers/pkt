package pm

import (
	"fmt"
	"os/exec"
)

// PNPM implements the PackageManager interface for pnpm
type PNPM struct{}

func (p *PNPM) Name() string {
	return "pnpm"
}

func (p *PNPM) IsAvailable() bool {
	return CheckAvailability("pnpm")
}

func (p *PNPM) Add(workDir, pkg string, flags []string) error {
	args := []string{"add", pkg}
	args = append(args, flags...)

	cmd := exec.Command("pnpm", args...)
	cmd.Dir = workDir
	cmd.Stdout = nil
	cmd.Stderr = nil

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("pnpm add failed: %w", err)
	}

	return nil
}

func (p *PNPM) AddMultiple(workDir string, packages []string, flags []string) error {
	args := []string{"add"}
	args = append(args, packages...)
	args = append(args, flags...)

	cmd := exec.Command("pnpm", args...)
	cmd.Dir = workDir
	cmd.Stdout = nil
	cmd.Stderr = nil

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("pnpm add failed: %w", err)
	}

	return nil
}

func (p *PNPM) Remove(workDir, pkg string) error {
	cmd := exec.Command("pnpm", "remove", pkg)
	cmd.Dir = workDir
	cmd.Stdout = nil
	cmd.Stderr = nil

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("pnpm remove failed: %w", err)
	}

	return nil
}

func (p *PNPM) Init(workDir string) error {
	cmd := exec.Command("pnpm", "init", "-y")
	cmd.Dir = workDir
	cmd.Stdout = nil
	cmd.Stderr = nil

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("pnpm init failed: %w", err)
	}

	return nil
}
