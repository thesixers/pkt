package cmd

import (
	"fmt"

	"github.com/genesix/pkt/internal/db"
	"github.com/genesix/pkt/internal/pm"
	"github.com/genesix/pkt/internal/utils"
	"github.com/spf13/cobra"
)

var pmCmd = &cobra.Command{
	Use:   "pm",
	Short: "Package manager commands",
	Long:  `Manage package manager settings for projects.`,
}

var pmSetCmd = &cobra.Command{
	Use:   "set <pm> <project | id | .>",
	Short: "Set the package manager for a project",
	Long: `Change the package manager for a project.
Available package managers: pnpm, npm, bun`,
	Args: cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		newPM := args[0]
		projectInput := args[1]

		// Check if PM is valid and available
		if !pm.CheckAvailability(newPM) {
			return fmt.Errorf("package manager '%s' is not available on this system", newPM)
		}

		// Validate PM is supported
		if _, err := pm.GetPM(newPM); err != nil {
			return fmt.Errorf("unsupported package manager: %s (use pnpm, npm, or bun)", newPM)
		}

		// Resolve project
		project, err := utils.ResolveProject(projectInput)
		if err != nil {
			return err
		}

		// Check if already using this PM
		if project.PackageManager == newPM {
			fmt.Printf("Project %s is already using %s\n", project.Name, newPM)
			return nil
		}

		// Update database
		if err := db.UpdateProjectPM(project.ID, newPM); err != nil {
			return fmt.Errorf("failed to update package manager: %w", err)
		}

		// Rewrite package.json scripts
		if err := utils.RewriteScripts(project.Path, newPM); err != nil {
			// Log warning but don't fail
			fmt.Printf("Warning: failed to rewrite scripts: %v\n", err)
		}

		// Re-sync dependencies
		deps, err := utils.ParsePackageJSON(project.Path)
		if err != nil {
			// Log warning but don't fail
			fmt.Printf("Warning: failed to sync dependencies: %v\n", err)
		} else {
			if err := db.SyncDependencies(project.ID, deps); err != nil {
				fmt.Printf("Warning: failed to sync dependencies: %v\n", err)
			}
		}

		fmt.Printf("✓ Updated package manager for %s: %s → %s\n",
			project.Name,
			project.PackageManager,
			newPM,
		)

		return nil
	},
}

func init() {
	pmCmd.AddCommand(pmSetCmd)
}
