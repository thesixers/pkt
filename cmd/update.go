package cmd

import (
	"fmt"
	"os"

	"github.com/genesix/pkt/internal/db"
	"github.com/genesix/pkt/internal/pm"
	"github.com/genesix/pkt/internal/utils"
	"github.com/spf13/cobra"
)

var updateCmd = &cobra.Command{
	Use:   "update [packages...]",
	Short: "Update dependencies in the current project",
	Long: `Update one or more dependencies to their latest versions.
If no packages are specified, all dependencies are updated.
Must be run inside a tracked project folder.

Examples:
  pkt update              # Update all dependencies
  pkt update lodash       # Update specific package
  pkt update axios react  # Update multiple packages`,
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
			return fmt.Errorf("not in a tracked project. Run 'pkt init .' first")
		}

		// Get package manager
		packageManager, err := pm.Get(project.Language, project.PackageManager)
		if err != nil {
			return err
		}

		// Update dependencies
		if len(packages) == 0 {
			fmt.Println("📦 Updating all dependencies...")
		} else {
			fmt.Printf("📦 Updating %v...\n", packages)
		}

		if err := packageManager.Update(cwd, packages); err != nil {
			return fmt.Errorf("failed to update dependencies: %w", err)
		}

		// Sync dependencies to database
		deps, err := utils.ParseDependencies(cwd, project.Language)
		if err != nil {
			fmt.Printf("⚠️  Warning: failed to parse dependencies: %v\n", err)
		} else if len(deps) > 0 {
			if err := db.SyncDependencies(project.ID, deps); err != nil {
				fmt.Printf("⚠️  Warning: failed to sync dependencies: %v\n", err)
			}
		}

		fmt.Println("✓ Dependencies updated")

		return nil
	},
}

func init() {
	rootCmd.AddCommand(updateCmd)
}
