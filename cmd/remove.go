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

var removeCmd = &cobra.Command{
	Use:   "remove <package> [packages...]",
	Short: "Remove dependencies from the current project",
	Long: `Remove one or more dependencies from the current project using its package manager.
Must be run inside a tracked project folder.`,
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

		// Get package manager
		packageManager, err := pm.Get(project.Language, project.PackageManager)
		if err != nil {
			return err
		}

		// Remove dependencies
		packagesStr := strings.Join(packages, " ")
		fmt.Printf("Removing %s...\n", packagesStr)

		if err := packageManager.Remove(cwd, packages); err != nil {
			return fmt.Errorf("failed to remove dependencies: %w", err)
		}

		// Sync dependencies to database for all languages
		deps, err := utils.ParseDependencies(cwd, project.Language)
		if err != nil {
			fmt.Printf("⚠️  Warning: failed to parse dependencies: %v\n", err)
		} else {
			if err := db.SyncDependencies(project.ID, deps); err != nil {
				fmt.Printf("⚠️  Warning: failed to sync dependencies: %v\n", err)
			}
		}

		if len(packages) == 1 {
			fmt.Printf("✓ Removed %s\n", packages[0])
		} else {
			fmt.Printf("✓ Removed %d packages: %s\n", len(packages), packagesStr)
		}

		return nil
	},
}
