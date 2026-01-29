package pm

import (
	"fmt"
	"os/exec"
)

// Bun implements the PackageManager interface for bun
type Bun struct{}

func (b *Bun) Name() string {
	return "bun"
}

func (b *Bun) IsAvailable() bool {
	return CheckAvailability("bun")
}

func (b *Bun) Add(workDir, pkg string, flags []string) error {
	args := []string{"add", pkg}
	args = append(args, flags...)

	cmd := exec.Command("bun", args...)
	cmd.Dir = workDir
	cmd.Stdout = nil
	cmd.Stderr = nil

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("bun add failed: %w", err)
	}

	return nil
}

func (b *Bun) Remove(workDir, pkg string) error {
	cmd := exec.Command("bun", "remove", pkg)
	cmd.Dir = workDir
	cmd.Stdout = nil
	cmd.Stderr = nil

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("bun remove failed: %w", err)
	}

	return nil
}

func (b *Bun) Init(workDir string) error {
	cmd := exec.Command("bun", "init")
	cmd.Dir = workDir
	cmd.Stdout = nil
	cmd.Stderr = nil

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("bun init failed: %w", err)
	}

	return nil
}
