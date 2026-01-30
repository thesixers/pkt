package cmd

import (
	"fmt"
	"os"
	"strings"

	"github.com/genesix/pkt/internal/db"
	"github.com/genesix/pkt/internal/pm"
	"github.com/genesix/pkt/internal/utils"
	"github.com/spf13/cobra"
)

var (
	devFlag bool
)

var addCmd = &cobra.Command{
	Use:   "add <package> [packages...]",
	Short: "Add dependencies to the current project",
	Long: `Add one or more dependencies to the current project using its package manager.
Must be run inside a tracked project folder.

Supports: JavaScript (npm/pnpm/bun), Python (uv/pip/poetry), Go, Rust

Examples:
  pkt add axios                    # JavaScript
  pkt add requests flask           # Python
  pkt add -D typescript eslint     # Dev dependencies`,
	Args: cobra.MinimumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		packages := args

		// Get current directory
		cwd, err := os.Getwd()
		if err != nil {
			return fmt.Errorf("failed to get current directory: %w", err)
		}

		// Get project from current directory
		project, err := db.GetProjectByPath(cwd)
		if err != nil {
			return fmt.Errorf("not in a tracked project. Run this command inside a project folder")
		}

		// Get package manager for this language
		packageManager, err := pm.Get(project.Language, project.PackageManager)
		if err != nil {
			return err
		}

		// Add dependencies
		packagesStr := strings.Join(packages, " ")
		fmt.Printf("Adding %s...\n", packagesStr)

		if err := packageManager.Add(cwd, packages, devFlag); err != nil {
			return fmt.Errorf("failed to add dependencies: %w", err)
		}

		// Sync dependencies to database (JavaScript only for now)
		if project.Language == "javascript" {
			deps, err := utils.ParsePackageJSON(cwd)
			if err != nil {
				fmt.Printf("⚠️  Warning: failed to parse dependencies: %v\n", err)
			} else if err := db.SyncDependencies(project.ID, deps); err != nil {
				fmt.Printf("⚠️  Warning: failed to sync dependencies: %v\n", err)
			}
		}

		if len(packages) == 1 {
			fmt.Printf("✓ Added %s\n", packages[0])
		} else {
			fmt.Printf("✓ Added %d packages: %s\n", len(packages), packagesStr)
		}

		return nil
	},
}

func init() {
	addCmd.Flags().BoolVarP(&devFlag, "dev", "D", false, "Add as dev dependency")
}
