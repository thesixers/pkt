package cmd

import (
	"fmt"
	"os"

	"github.com/genesix/pkt/internal/db"
	"github.com/genesix/pkt/internal/pm"
	"github.com/genesix/pkt/internal/utils"
	"github.com/spf13/cobra"
)

var installCmd = &cobra.Command{
	Use:   "install",
	Short: "Install all dependencies for the current project",
	Long: `Install all dependencies for the current project.
This command must be run inside a tracked project folder.

Supports: JavaScript (npm/pnpm/bun), Python (uv/pip/poetry), Go, Rust

Example:
  cd my-project
  pkt install`,
	Args: cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		// Get current directory
		cwd, err := os.Getwd()
		if err != nil {
			return fmt.Errorf("failed to get current directory: %w", err)
		}

		// Get project from current directory
		project, err := db.GetProjectByPath(cwd)
		if err != nil {
			return fmt.Errorf("not in a tracked project folder\nRun 'pkt init .' to track this project first")
		}

		// Get package manager
		packageManager, err := pm.Get(project.Language, project.PackageManager)
		if err != nil {
			return err
		}

		fmt.Printf("📦 Installing dependencies using %s for %s project...\n", project.PackageManager, project.Language)

		// Run install command
		if err := packageManager.Install(cwd); err != nil {
			return fmt.Errorf("failed to install dependencies: %w", err)
		}

		// Sync dependencies to database (JavaScript only for now)
		if project.Language == "javascript" {
			deps, err := utils.ParsePackageJSON(cwd)
			if err != nil {
				fmt.Printf("⚠️  Warning: failed to parse dependencies: %v\n", err)
			} else if err := db.SyncDependencies(project.ID, deps); err != nil {
				fmt.Printf("⚠️  Warning: failed to sync dependencies: %v\n", err)
			} else {
				fmt.Printf("✓ Synced %d dependencies to database\n", len(deps))
			}
		}

		fmt.Println("✓ Dependencies installed")

		return nil
	},
}

func init() {
	rootCmd.AddCommand(installCmd)
}
