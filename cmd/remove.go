package cmd

import (
	"fmt"
	"os"

	"github.com/genesix/pkt/internal/db"
	"github.com/genesix/pkt/internal/pm"
	"github.com/genesix/pkt/internal/utils"
	"github.com/spf13/cobra"
)

var removeCmd = &cobra.Command{
	Use:   "remove <package>",
	Short: "Remove a dependency from the current project",
	Long: `Remove a dependency from the current project using its package manager.
Must be run inside a tracked project folder.`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		packageName := args[0]

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
		packageManager, err := pm.GetPM(project.PackageManager)
		if err != nil {
			return err
		}

		// Remove dependency
		fmt.Printf("Removing %s...\n", packageName)
		if err := packageManager.Remove(cwd, packageName); err != nil {
			return fmt.Errorf("failed to remove dependency: %w", err)
		}

		// Sync dependencies to database
		deps, err := utils.ParsePackageJSON(cwd)
		if err != nil {
			return fmt.Errorf("failed to parse package.json: %w", err)
		}

		if err := db.SyncDependencies(project.ID, deps); err != nil {
			return fmt.Errorf("failed to sync dependencies: %w", err)
		}

		fmt.Printf("✓ Removed %s\n", packageName)

		return nil
	},
}
