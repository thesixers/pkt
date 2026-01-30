package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"

	"github.com/genesix/pkt/internal/utils"
	"github.com/spf13/cobra"
)

var execCmd = &cobra.Command{
	Use:   "exec <project> <command> [args...]",
	Short: "Execute a command in a project's context",
	Long: `Execute a command in the context of a project.

This command will:
  - Change to the project directory
  - For Python projects: Activate the virtual environment
  - Run the specified command

Examples:
  pkt exec my-app "npm run build"
  pkt exec my-py-app "pytest -v"
  pkt exec my-go-app go build -o bin/app
  pkt exec my-rs-app cargo build --release`,
	Args: cobra.MinimumNArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		projectRef := args[0]
		command := args[1]
		cmdArgs := args[2:]

		// Resolve project
		project, err := utils.ResolveProject(projectRef)
		if err != nil {
			return err
		}

		fmt.Printf("📂 Executing in %s (%s)...\n\n", project.Name, project.Path)

		// Build command
		var execCmd *exec.Cmd

		// For Python with pip, use venv's environment
		if project.Language == "python" && project.PackageManager == "pip" {
			venvPython := filepath.Join(project.Path, ".venv", "bin", "python")
			if runtime.GOOS == "windows" {
				venvPython = filepath.Join(project.Path, ".venv", "Scripts", "python.exe")
			}

			// Check if venv exists
			if _, err := os.Stat(venvPython); err == nil {
				// Use bash -c to run the command with venv activated
				if runtime.GOOS == "windows" {
					// On Windows, use cmd /c with venv activation
					activateScript := filepath.Join(project.Path, ".venv", "Scripts", "activate.bat")
					fullCmd := fmt.Sprintf("%s && %s %s", activateScript, command, joinArgs(cmdArgs))
					execCmd = exec.Command("cmd", "/c", fullCmd)
				} else {
					// On Unix, source the activate script
					activateScript := filepath.Join(project.Path, ".venv", "bin", "activate")
					fullCmd := fmt.Sprintf("source %s && %s %s", activateScript, command, joinArgs(cmdArgs))
					execCmd = exec.Command("bash", "-c", fullCmd)
				}
			} else {
				// No venv, run directly
				execCmd = exec.Command("bash", "-c", command+" "+joinArgs(cmdArgs))
			}
		} else {
			// For other languages, run directly
			if len(cmdArgs) > 0 {
				execCmd = exec.Command(command, cmdArgs...)
			} else {
				// Use shell to handle complex commands
				execCmd = exec.Command("bash", "-c", command)
			}
		}

		execCmd.Dir = project.Path
		execCmd.Stdout = os.Stdout
		execCmd.Stderr = os.Stderr
		execCmd.Stdin = os.Stdin

		return execCmd.Run()
	},
}

func joinArgs(args []string) string {
	if len(args) == 0 {
		return ""
	}
	result := ""
	for i, arg := range args {
		if i > 0 {
			result += " "
		}
		// Quote args with spaces
		if contains(arg, " ") {
			result += "\"" + arg + "\""
		} else {
			result += arg
		}
	}
	return result
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > 0 && containsHelper(s, substr))
}

func containsHelper(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

func init() {
	rootCmd.AddCommand(execCmd)
}
