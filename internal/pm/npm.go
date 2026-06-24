package pm

import (
	"os/exec"
	"strings"
)

// NPM implements PackageManager for npm
type NPM struct{}

func (n *NPM) Name() string {
	return "npm"
}

func (n *NPM) Language() string {
	return "javascript"
}

func (n *NPM) Add(workDir string, packages []string, dev bool) error {
	args := []string{"install"}
	if dev {
		args = append(args, "--save-dev")
	}
	args = append(args, packages...)
	return runCommandInteractive("npm", args, workDir)
}

func (n *NPM) Remove(workDir string, packages []string) error {
	args := append([]string{"uninstall"}, packages...)
	return runCommandInteractive("npm", args, workDir)
}

func (n *NPM) Install(workDir string) error {
	return runCommandInteractive("npm", []string{"install"}, workDir)
}

func (n *NPM) Init(workDir string) error {
	return runCommand("npm", []string{"init", "-y"}, workDir)
}

func (n *NPM) Run(workDir string, script string, args []string) error {
	cmdArgs := []string{"run", script}
	if len(args) > 0 {
		cmdArgs = append(cmdArgs, "--")
		cmdArgs = append(cmdArgs, args...)
	}
	return runCommandInteractive("npm", cmdArgs, workDir)
}

func (n *NPM) RunFile(workDir string, target string, interpFlags []string, args []string) error {
	var cmd string
	var cmdArgs []string

	if strings.HasSuffix(target, ".ts") || strings.HasSuffix(target, ".tsx") {
		cmd = "npx"
		cmdArgs = append([]string{"tsx"}, interpFlags...)
		cmdArgs = append(cmdArgs, target)
	} else {
		cmd = "node"
		cmdArgs = append(cmdArgs, interpFlags...)
		cmdArgs = append(cmdArgs, target)
	}

	cmdArgs = append(cmdArgs, args...)
	return runCommandInteractive(cmd, cmdArgs, workDir)
}

func (n *NPM) Update(workDir string, packages []string) error {
	args := []string{"update"}
	args = append(args, packages...)
	return runCommandInteractive("npm", args, workDir)
}

func (n *NPM) IsAvailable() bool {
	_, err := exec.LookPath("npm")
	return err == nil
}
