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
	Long: `Install all dependencies listed in package.json for the current project.
This command must be run inside a tracked project folder.

It reads dependencies from package.json and installs them using the project's
configured package manager, then syncs the database with the installed versions.

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

		// Check for package.json
		pkgJSONPath := cwd + "/package.json"
		if _, err := os.Stat(pkgJSONPath); os.IsNotExist(err) {
			return fmt.Errorf("no package.json found in current directory")
		}

		// Parse package.json to get dependencies
		deps, err := utils.ParsePackageJSON(cwd)
		if err != nil {
			return fmt.Errorf("failed to parse package.json: %w", err)
		}

		if len(deps) == 0 {
			fmt.Println("No dependencies to install.")
			return nil
		}

		// Separate prod and dev dependencies
		var prodDeps []string
		var devDeps []string
		for name, dep := range deps {
			if dep.DepType == "dev" {
				devDeps = append(devDeps, name)
			} else {
				prodDeps = append(prodDeps, name)
			}
		}

		// Get package manager
		packageManager, err := pm.GetPM(project.PackageManager)
		if err != nil {
			return err
		}

		fmt.Printf("📦 Installing dependencies using %s...\n", project.PackageManager)

		// Install prod dependencies
		if len(prodDeps) > 0 {
			fmt.Printf("  Installing %d production dependencies...\n", len(prodDeps))
			if err := packageManager.AddMultiple(cwd, prodDeps, nil); err != nil {
				return fmt.Errorf("failed to install production dependencies: %w", err)
			}
		}

		// Install dev dependencies
		if len(devDeps) > 0 {
			fmt.Printf("  Installing %d dev dependencies...\n", len(devDeps))
			if err := packageManager.AddMultiple(cwd, devDeps, []string{"-D"}); err != nil {
				return fmt.Errorf("failed to install dev dependencies: %w", err)
			}
		}

		// Re-parse package.json to get actual installed versions
		installedDeps, err := utils.ParsePackageJSON(cwd)
		if err != nil {
			fmt.Printf("⚠️  Warning: failed to re-parse dependencies: %v\n", err)
		} else {
			// Sync dependencies to database
			if err := db.SyncDependencies(project.ID, installedDeps); err != nil {
				fmt.Printf("⚠️  Warning: failed to sync dependencies to database: %v\n", err)
			}
		}

		fmt.Printf("✓ Installed %d dependencies (%d prod, %d dev)\n",
			len(prodDeps)+len(devDeps), len(prodDeps), len(devDeps))

		return nil
	},
}

func init() {
	rootCmd.AddCommand(installCmd)
}
