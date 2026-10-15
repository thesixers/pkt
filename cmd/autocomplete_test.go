package cmd

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestDetectShell(t *testing.T) {
	origShell := os.Getenv("SHELL")
	defer os.Setenv("SHELL", origShell)

	tests := []struct {
		envVal   string
		expected string
	}{
		{"/bin/bash", "bash"},
		{"/usr/bin/bash", "bash"},
		{"/bin/zsh", "zsh"},
		{"/usr/local/bin/zsh", "zsh"},
		{"/usr/bin/fish", "fish"},
	}

	for _, tt := range tests {
		os.Setenv("SHELL", tt.envVal)
		detected := DetectShell()
		if detected != tt.expected {
			t.Errorf("DetectShell() for SHELL=%q: got %q, want %q", tt.envVal, detected, tt.expected)
		}
	}
}

func TestInstallBashCompletion(t *testing.T) {
	tempHome := t.TempDir()

	res, err := InstallAutocompleteTo(rootCmd, "bash", tempHome)
	if err != nil {
		t.Fatalf("InstallAutocompleteTo(bash) error: %v", err)
	}

	if res.Shell != "bash" {
		t.Errorf("expected shell bash, got %s", res.Shell)
	}

	// Verify bash-completion script was written
	compPath := filepath.Join(tempHome, ".local", "share", "bash-completion", "completions", "pkt")
	if _, err := os.Stat(compPath); err != nil {
		t.Errorf("expected completion file at %s, got error: %v", compPath, err)
	}

	// Verify fallback script was written
	pktPath := filepath.Join(tempHome, ".pkt", "completions", "pkt.bash")
	if _, err := os.Stat(pktPath); err != nil {
		t.Errorf("expected fallback completion file at %s, got error: %v", pktPath, err)
	}

	// Verify .bashrc was created/updated
	rcPath := filepath.Join(tempHome, ".bashrc")
	content, err := os.ReadFile(rcPath)
	if err != nil {
		t.Fatalf("failed to read .bashrc: %v", err)
	}
	if !strings.Contains(string(content), "pkt shell completion") {
		t.Errorf("expected .bashrc to contain pkt completion snippet, got:\n%s", string(content))
	}

	// Test idempotency: running again should not duplicate snippet
	res2, err := InstallAutocompleteTo(rootCmd, "bash", tempHome)
	if err != nil {
		t.Fatalf("second InstallAutocompleteTo error: %v", err)
	}
	if res2.UpdatedRc {
		t.Errorf("expected second install not to modify rc file again")
	}

	content2, _ := os.ReadFile(rcPath)
	count := strings.Count(string(content2), "# pkt shell completion")
	if count != 1 {
		t.Errorf("expected exactly 1 completion snippet in .bashrc, found %d", count)
	}
}

func TestInstallZshCompletion(t *testing.T) {
	tempHome := t.TempDir()

	res, err := InstallAutocompleteTo(rootCmd, "zsh", tempHome)
	if err != nil {
		t.Fatalf("InstallAutocompleteTo(zsh) error: %v", err)
	}

	if res.Shell != "zsh" {
		t.Errorf("expected shell zsh, got %s", res.Shell)
	}

	// Verify script in ~/.pkt/completions/pkt.zsh
	compPath := filepath.Join(tempHome, ".pkt", "completions", "pkt.zsh")
	if _, err := os.Stat(compPath); err != nil {
		t.Errorf("expected zsh completion at %s, got: %v", compPath, err)
	}

	// Verify .zshrc
	rcPath := filepath.Join(tempHome, ".zshrc")
	content, err := os.ReadFile(rcPath)
	if err != nil {
		t.Fatalf("failed to read .zshrc: %v", err)
	}
	if !strings.Contains(string(content), "pkt shell completion") {
		t.Errorf("expected .zshrc to contain snippet, got:\n%s", string(content))
	}
}

func TestInstallFishCompletion(t *testing.T) {
	tempHome := t.TempDir()

	res, err := InstallAutocompleteTo(rootCmd, "fish", tempHome)
	if err != nil {
		t.Fatalf("InstallAutocompleteTo(fish) error: %v", err)
	}

	if res.Shell != "fish" {
		t.Errorf("expected shell fish, got %s", res.Shell)
	}

	// Verify script in ~/.config/fish/completions/pkt.fish
	compPath := filepath.Join(tempHome, ".config", "fish", "completions", "pkt.fish")
	if _, err := os.Stat(compPath); err != nil {
		t.Errorf("expected fish completion at %s, got: %v", compPath, err)
	}
}

func TestInstallPowerShellCompletion(t *testing.T) {
	tempHome := t.TempDir()

	res, err := InstallAutocompleteTo(rootCmd, "powershell", tempHome)
	if err != nil {
		t.Fatalf("InstallAutocompleteTo(powershell) error: %v", err)
	}

	if res.Shell != "powershell" {
		t.Errorf("expected shell powershell, got %s", res.Shell)
	}

	// Verify script in ~/.pkt/completions/pkt.ps1
	compPath := filepath.Join(tempHome, ".pkt", "completions", "pkt.ps1")
	if _, err := os.Stat(compPath); err != nil {
		t.Errorf("expected powershell completion at %s, got: %v", compPath, err)
	}
}

func TestInstallUnsupportedShell(t *testing.T) {
	tempHome := t.TempDir()
	_, err := InstallAutocompleteTo(rootCmd, "invalid_shell_xyz", tempHome)
	if err == nil {
		t.Errorf("expected error for unsupported shell, got nil")
	}
}
