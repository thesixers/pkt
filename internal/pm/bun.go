package pm

import "os/exec"

// Bun implements PackageManager for bun
type Bun struct{}

func (b *Bun) Name() string {
	return "bun"
}

func (b *Bun) Language() string {
	return "javascript"
}

func (b *Bun) Add(workDir string, packages []string, dev bool) error {
	args := []string{"add"}
	if dev {
		args = append(args, "-d")
	}
	args = append(args, packages...)
	return runCommand("bun", args, workDir)
}

func (b *Bun) Remove(workDir string, packages []string) error {
	args := append([]string{"remove"}, packages...)
	return runCommand("bun", args, workDir)
}

func (b *Bun) Install(workDir string) error {
	return runCommand("bun", []string{"install"}, workDir)
}

func (b *Bun) Init(workDir string) error {
	return runCommand("bun", []string{"init", "-y"}, workDir)
}

func (b *Bun) IsAvailable() bool {
	_, err := exec.LookPath("bun")
	return err == nil
}
