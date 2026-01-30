package pm

import "os/exec"

// PNPM implements PackageManager for pnpm
type PNPM struct{}

func (p *PNPM) Name() string {
	return "pnpm"
}

func (p *PNPM) Language() string {
	return "javascript"
}

func (p *PNPM) Add(workDir string, packages []string, dev bool) error {
	args := []string{"add"}
	if dev {
		args = append(args, "-D")
	}
	args = append(args, packages...)
	return runCommand("pnpm", args, workDir)
}

func (p *PNPM) Remove(workDir string, packages []string) error {
	args := append([]string{"remove"}, packages...)
	return runCommand("pnpm", args, workDir)
}

func (p *PNPM) Install(workDir string) error {
	return runCommand("pnpm", []string{"install"}, workDir)
}

func (p *PNPM) Init(workDir string) error {
	return runCommand("pnpm", []string{"init"}, workDir)
}

func (p *PNPM) Run(workDir string, script string, args []string) error {
	cmdArgs := []string{"run", script}
	cmdArgs = append(cmdArgs, args...)
	return runCommandInteractive("pnpm", cmdArgs, workDir)
}

func (p *PNPM) Update(workDir string, packages []string) error {
	args := []string{"update"}
	args = append(args, packages...)
	return runCommand("pnpm", args, workDir)
}

func (p *PNPM) IsAvailable() bool {
	_, err := exec.LookPath("pnpm")
	return err == nil
}
