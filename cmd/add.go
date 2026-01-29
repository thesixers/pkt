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

Examples:
  pkt add axios
  pkt add axios nodemon express
  pkt add -D typescript eslint`,
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

		// Add all dependencies at once
		packagesStr := strings.Join(packages, " ")
		fmt.Printf("Adding %s...\n", packagesStr)

		if err := packageManager.AddMultiple(cwd, packages, flags); err != nil {
			return fmt.Errorf("failed to add dependencies: %w", err)
		}

		// Sync dependencies to database
		deps, err := utils.ParsePackageJSON(cwd)
		if err != nil {
			return fmt.Errorf("failed to parse package.json: %w", err)
		}

		if err := db.SyncDependencies(project.ID, deps); err != nil {
			return fmt.Errorf("failed to sync dependencies: %w", err)
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
