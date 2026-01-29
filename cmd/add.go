package cmd

import (
	"fmt"
	"os"

	"github.com/genesix/pkt/internal/db"
	"github.com/genesix/pkt/internal/pm"
	"github.com/genesix/pkt/internal/utils"
	"github.com/spf13/cobra"
)

var (
	devFlag bool
)

var addCmd = &cobra.Command{
	Use:   "add <package>",
	Short: "Add a dependency to the current project",
	Long: `Add a dependency to the current project using its package manager.
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

		// Check if package.json exists, create if needed
		pkgJSONPath := cwd + "/package.json"
		if _, err := os.Stat(pkgJSONPath); os.IsNotExist(err) {
			fmt.Println("Creating package.json...")
			if err := utils.CreatePackageJSON(cwd, project.Name); err != nil {
				return fmt.Errorf("failed to create package.json: %w", err)
			}
		}

		// Build flags
		var flags []string
		if devFlag {
			flags = append(flags, "-D")
		}

		// Add dependency
		fmt.Printf("Adding %s...\n", packageName)
		if err := packageManager.Add(cwd, packageName, flags); err != nil {
			return fmt.Errorf("failed to add dependency: %w", err)
		}

		// Sync dependencies to database
		deps, err := utils.ParsePackageJSON(cwd)
		if err != nil {
			return fmt.Errorf("failed to parse package.json: %w", err)
		}

		if err := db.SyncDependencies(project.ID, deps); err != nil {
			return fmt.Errorf("failed to sync dependencies: %w", err)
		}

		fmt.Printf("✓ Added %s\n", packageName)

		return nil
	},
}

func init() {
	addCmd.Flags().BoolVarP(&devFlag, "dev", "D", false, "Add as dev dependency")
}
