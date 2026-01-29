package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/genesix/pkt/internal/db"
	"github.com/spf13/cobra"
)

var renameCmd = &cobra.Command{
	Use:   "rename <old-name> <new-name>",
	Short: "Rename a tracked project",
	Long: `Rename a project in pkt's registry and on disk.
This updates both the project name in the database and renames the folder.`,
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

		// Check if new name already exists in database
		existingProjects, err := db.GetProjectsByName(newName)
		if err != nil {
			return fmt.Errorf("failed to check existing projects: %w", err)
		}

		if len(existingProjects) > 0 {
			return fmt.Errorf("a project named '%s' already exists", newName)
		}

		// Calculate new path (replace old folder name with new name)
		oldPath := project.Path
		parentDir := filepath.Dir(oldPath)
		newPath := filepath.Join(parentDir, newName)

		// Check if new path already exists on disk
		if _, err := os.Stat(newPath); err == nil {
			return fmt.Errorf("folder '%s' already exists on disk", newPath)
		}

		// Rename the folder on disk
		if err := os.Rename(oldPath, newPath); err != nil {
			return fmt.Errorf("failed to rename folder: %w", err)
		}

		// Update the project name and path in database
		if err := db.RenameProjectWithPath(project.ID, newName, newPath); err != nil {
			// Try to revert the folder rename if database update fails
			_ = os.Rename(newPath, oldPath)
			return fmt.Errorf("failed to update database: %w", err)
		}

		fmt.Printf("✓ Renamed project '%s' to '%s'\n", oldName, newName)
		fmt.Printf("  Folder: %s → %s\n", oldPath, newPath)

		return nil
	},
}
