package pm

import "os/exec"

// GoMod implements PackageManager for Go modules
type GoMod struct{}

func (g *GoMod) Name() string {
	return "go"
}

func (g *GoMod) Language() string {
	return "go"
}

func (g *GoMod) Add(workDir string, packages []string, dev bool) error {
	// Go doesn't distinguish between dev and prod dependencies
	for _, pkg := range packages {
		if err := runCommand("go", []string{"get", pkg}, workDir); err != nil {
			return err
		}
	}
	return nil
}

func (g *GoMod) Remove(workDir string, packages []string) error {
	// In Go, we remove by editing go.mod and running go mod tidy
	// For now, just run go mod tidy after manual removal
	return runCommand("go", []string{"mod", "tidy"}, workDir)
}

func (g *GoMod) Install(workDir string) error {
	return runCommand("go", []string{"mod", "download"}, workDir)
}

func (g *GoMod) Init(workDir string) error {
	// Note: module name should be specified by the user
	// Using a placeholder that they can edit
	return runCommand("go", []string{"mod", "init", "example.com/project"}, workDir)
}

func (g *GoMod) Run(workDir string, script string, args []string) error {
	switch script {
	case "build":
		return runCommandInteractive("go", []string{"build", "./..."}, workDir)
	case "test":
		cmdArgs := []string{"test", "./..."}
		cmdArgs = append(cmdArgs, args...)
		return runCommandInteractive("go", cmdArgs, workDir)
	case "run":
		cmdArgs := []string{"run", "."}
		cmdArgs = append(cmdArgs, args...)
		return runCommandInteractive("go", cmdArgs, workDir)
	default:
		// Try to run as a go file or command
		cmdArgs := []string{"run", script}
		cmdArgs = append(cmdArgs, args...)
		return runCommandInteractive("go", cmdArgs, workDir)
	}
}

func (g *GoMod) Update(workDir string, packages []string) error {
	if len(packages) == 0 {
		// Update all dependencies
		return runCommand("go", []string{"get", "-u", "./..."}, workDir)
	}
	// Update specific packages
	for _, pkg := range packages {
		if err := runCommand("go", []string{"get", "-u", pkg}, workDir); err != nil {
			return err
		}
	}
	return nil
}

func (g *GoMod) IsAvailable() bool {
	_, err := exec.LookPath("go")
	return err == nil
}
