package pm

import "os/exec"

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
	return runCommand("cargo", args, workDir)
}

func (c *Cargo) Remove(workDir string, packages []string) error {
	args := append([]string{"remove"}, packages...)
	return runCommand("cargo", args, workDir)
}

func (c *Cargo) Install(workDir string) error {
	return runCommand("cargo", []string{"build"}, workDir)
}

func (c *Cargo) Init(workDir string) error {
	return runCommand("cargo", []string{"init", "."}, workDir)
}

func (c *Cargo) IsAvailable() bool {
	_, err := exec.LookPath("cargo")
	return err == nil
}
