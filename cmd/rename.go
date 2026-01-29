package cmd

import (
	"fmt"

	"github.com/genesix/pkt/internal/db"
	"github.com/spf13/cobra"
)

var renameCmd = &cobra.Command{
	Use:   "rename <old-name> <new-name>",
	Short: "Rename a tracked project",
	Long: `Rename a project in pkt's registry.
This only updates the project name in the database, not the folder on disk.`,
	Args: cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		oldName := args[0]
		newName := args[1]

		// Find the project by old name
		projects, err := db.GetProjectsByName(oldName)
		if err != nil {
			return fmt.Errorf("failed to query projects: %w", err)
		}

		if len(projects) == 0 {
			return fmt.Errorf("project '%s' not found", oldName)
		}

		// If multiple projects have the same name, use the first one
		// (most recent by created_at DESC)
		project := projects[0]

		// Check if new name already exists
		existingProjects, err := db.GetProjectsByName(newName)
		if err != nil {
			return fmt.Errorf("failed to check existing projects: %w", err)
		}

		if len(existingProjects) > 0 {
			return fmt.Errorf("a project named '%s' already exists", newName)
		}

		// Update the project name
		if err := db.RenameProject(project.ID, newName); err != nil {
			return fmt.Errorf("failed to rename project: %w", err)
		}

		fmt.Printf("✓ Renamed project '%s' to '%s'\n", oldName, newName)

		return nil
	},
}
