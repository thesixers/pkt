package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/genesix/pkt/internal/db"
	"github.com/genesix/pkt/internal/utils"
	"github.com/jedib0t/go-pretty/v6/table"
	"github.com/spf13/cobra"
)

var outdatedCmd = &cobra.Command{
	Use:   "outdated [project]",
	Short: "Check for outdated dependencies",
	Long: `Check for outdated dependencies in a project.
If no project is specified, uses the current directory.

Supports: JavaScript (npm/pnpm), Python (pip/poetry), Go, Rust

Examples:
  pkt outdated           # Check current project
  pkt outdated my-app    # Check specific project`,
	Args:              cobra.MaximumNArgs(1),
	ValidArgsFunction: completeProjects,
	RunE: func(cmd *cobra.Command, args []string) error {
		var project *db.Project
		var err error
		var projectPath string

		// Determine which project to use
		if len(args) == 0 {
			// Use current directory
			cwd, err := os.Getwd()
			if err != nil {
				return fmt.Errorf("failed to get current directory: %w", err)
			}
			project, err = db.GetProjectByPath(cwd)
			if err != nil {
				return fmt.Errorf("not in a tracked project")
			}
			projectPath = cwd
		} else {
			// Resolve from argument
			project, err = utils.ResolveProject(args[0])
			if err != nil {
				return err
			}
			projectPath = project.Path
		}

		fmt.Printf("📦 Checking outdated dependencies for %s...\n\n", project.Name)

		// Run language-specific outdated check
		switch project.Language {
		case "javascript":
			return checkOutdatedJS(projectPath, project.PackageManager)
		case "python":
			return checkOutdatedPython(projectPath, project.PackageManager)
		case "go":
			return checkOutdatedGo(projectPath)
		case "rust":
			return checkOutdatedRust(projectPath)
		default:
			return fmt.Errorf("outdated check not supported for %s", project.Language)
		}
	},
}

// JavaScript outdated check
func checkOutdatedJS(workDir, pm string) error {
	var cmd *exec.Cmd
	switch pm {
	case "npm":
		cmd = exec.Command("npm", "outdated")
	case "pnpm":
		cmd = exec.Command("pnpm", "outdated")
	case "bun":
		// Bun doesn't have outdated, use npm
		cmd = exec.Command("npm", "outdated")
	default:
		cmd = exec.Command("npm", "outdated")
	}
	cmd.Dir = workDir
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	// npm outdated returns exit code 1 if there are outdated packages
	_ = cmd.Run()
	return nil
}

// Python outdated check
func checkOutdatedPython(workDir, pm string) error {
	var cmd *exec.Cmd
	switch pm {
	case "pip":
		// Use venv pip if available
		venvPip := workDir + "/venv/bin/pip"
		if _, err := os.Stat(venvPip); err == nil {
			cmd = exec.Command(venvPip, "list", "--outdated", "--format=json")
		} else {
			cmd = exec.Command("pip", "list", "--outdated", "--format=json")
		}
	case "poetry":
		cmd = exec.Command("poetry", "show", "--outdated")
		cmd.Dir = workDir
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		return cmd.Run()
	case "uv":
		// uv doesn't have a direct outdated command yet
		fmt.Println("Note: uv doesn't have a built-in outdated command.")
		fmt.Println("Consider using: pip list --outdated")
		return nil
	default:
		cmd = exec.Command("pip", "list", "--outdated", "--format=json")
	}

	cmd.Dir = workDir
	output, err := cmd.Output()
	if err != nil {
		return fmt.Errorf("failed to check outdated: %w", err)
	}

	// Parse JSON output
	var outdated []struct {
		Name           string `json:"name"`
		Version        string `json:"version"`
		LatestVersion  string `json:"latest_version"`
		LatestFiletype string `json:"latest_filetype"`
	}
	if err := json.Unmarshal(output, &outdated); err != nil {
		return fmt.Errorf("failed to parse outdated output: %w", err)
	}

	if len(outdated) == 0 {
		fmt.Println("All packages are up to date! ✓")
		return nil
	}

	// Display in table format
	t := table.NewWriter()
	t.SetOutputMirror(os.Stdout)
	t.SetStyle(table.StyleRounded)
	t.AppendHeader(table.Row{"NAME", "CURRENT", "LATEST", "TYPE"})

	for _, pkg := range outdated {
		t.AppendRow(table.Row{pkg.Name, pkg.Version, pkg.LatestVersion, "prod"})
	}

	t.Render()

	return nil
}

// Go outdated check
func checkOutdatedGo(workDir string) error {
	cmd := exec.Command("go", "list", "-u", "-m", "all")
	cmd.Dir = workDir
	output, err := cmd.Output()
	if err != nil {
		return fmt.Errorf("failed to check outdated: %w", err)
	}

	lines := strings.Split(string(output), "\n")
	hasOutdated := false

	t := table.NewWriter()
	t.SetOutputMirror(os.Stdout)
	t.SetStyle(table.StyleRounded)
	t.AppendHeader(table.Row{"NAME", "CURRENT", "LATEST"})

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		// Look for lines with [version] which indicate updates available
		if strings.Contains(line, "[") {
			parts := strings.Fields(line)
			if len(parts) >= 3 {
				name := parts[0]
				current := parts[1]
				latest := strings.Trim(parts[2], "[]")
				t.AppendRow(table.Row{name, current, latest})
				hasOutdated = true
			}
		}
	}

	if !hasOutdated {
		fmt.Println("All packages are up to date! ✓")
		return nil
	}

	t.Render()
	return nil
}

// Rust outdated check
func checkOutdatedRust(workDir string) error {
	// Check if cargo-outdated is installed
	if _, err := exec.LookPath("cargo-outdated"); err != nil {
		fmt.Println("Note: Install cargo-outdated for better output:")
		fmt.Println("  cargo install cargo-outdated")
		fmt.Println()
		fmt.Println("Using 'cargo update --dry-run' instead:")
		fmt.Println()

		cmd := exec.Command("cargo", "update", "--dry-run")
		cmd.Dir = workDir
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		return cmd.Run()
	}

	cmd := exec.Command("cargo", "outdated", "-R")
	cmd.Dir = workDir
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func init() {
	rootCmd.AddCommand(outdatedCmd)
}
