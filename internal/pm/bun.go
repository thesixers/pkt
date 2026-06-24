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
	return runCommandInteractive("bun", args, workDir)
}

func (b *Bun) Remove(workDir string, packages []string) error {
	args := append([]string{"remove"}, packages...)
	return runCommandInteractive("bun", args, workDir)
}

func (b *Bun) Install(workDir string) error {
	return runCommandInteractive("bun", []string{"install"}, workDir)
}

func (b *Bun) Init(workDir string) error {
	return runCommand("bun", []string{"init", "-y"}, workDir)
}

func (b *Bun) Run(workDir string, script string, args []string) error {
	cmdArgs := []string{"run", script}
	cmdArgs = append(cmdArgs, args...)
	return runCommandInteractive("bun", cmdArgs, workDir)
}

func (b *Bun) RunFile(workDir string, target string, interpFlags []string, args []string) error {
	cmdArgs := append([]string{}, interpFlags...)
	cmdArgs = append(cmdArgs, "run", target)
	cmdArgs = append(cmdArgs, args...)
	return runCommandInteractive("bun", cmdArgs, workDir)
}

func (b *Bun) Update(workDir string, packages []string) error {
	args := []string{"update"}
	args = append(args, packages...)
	return runCommandInteractive("bun", args, workDir)
}

func (b *Bun) IsAvailable() bool {
	_, err := exec.LookPath("bun")
	return err == nil
}
