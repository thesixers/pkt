package pm

import (
	"fmt"
	"os/exec"
)

// Cargo implements PackageManager for Rust's Cargo
type Cargo struct{}

func (c *Cargo) Name() string {
	return "cargo"
}

func (c *Cargo) Language() string {
	return "rust"
}

func (c *Cargo) Add(workDir string, packages []string, dev bool) error {
	args := []string{"add"}
	if dev {
		args = append(args, "--dev")
	}
	args = append(args, packages...)
	return runCommandInteractive("cargo", args, workDir)
}

func (c *Cargo) Remove(workDir string, packages []string) error {
	args := append([]string{"remove"}, packages...)
	return runCommandInteractive("cargo", args, workDir)
}

func (c *Cargo) Install(workDir string) error {
	return runCommand("cargo", []string{"build"}, workDir)
}

func (c *Cargo) Init(workDir string) error {
	return runCommand("cargo", []string{"init", "."}, workDir)
}

func (c *Cargo) Run(workDir string, script string, args []string) error {
	switch script {
	case "build":
		return runCommandInteractive("cargo", []string{"build"}, workDir)
	case "test":
		cmdArgs := []string{"test"}
		cmdArgs = append(cmdArgs, args...)
		return runCommandInteractive("cargo", cmdArgs, workDir)
	case "run":
		cmdArgs := []string{"run"}
		cmdArgs = append(cmdArgs, args...)
		return runCommandInteractive("cargo", cmdArgs, workDir)
	default:
		// Default to cargo run with the script as binary name
		cmdArgs := []string{"run", "--bin", script}
		cmdArgs = append(cmdArgs, args...)
		return runCommandInteractive("cargo", cmdArgs, workDir)
	}
}

func (c *Cargo) RunFile(workDir string, target string, interpFlags []string, args []string) error {
	return fmt.Errorf("rust does not support running individual files outside a cargo project. Use 'pkt run run' or 'pkt run <binary-name>'")
}

func (c *Cargo) Update(workDir string, packages []string) error {
	args := []string{"update"}
	if len(packages) > 0 {
		for _, pkg := range packages {
			args = append(args, "-p", pkg)
		}
	}
	return runCommandInteractive("cargo", args, workDir)
}

func (c *Cargo) IsAvailable() bool {
	_, err := exec.LookPath("cargo")
	return err == nil
}
