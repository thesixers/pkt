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

func (g *GoMod) IsAvailable() bool {
	_, err := exec.LookPath("go")
	return err == nil
}
