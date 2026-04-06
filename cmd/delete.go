package cmd

import (
	"fmt"

	"github.com/AlecAivazis/survey/v2"
	"github.com/genesix/pkt/internal/db"
	"github.com/genesix/pkt/internal/utils"
	"github.com/spf13/cobra"
)

var deleteCmd = &cobra.Command{
	Use:   "delete <project | id>...",
	Short: "Delete one or more projects",
	Long: `Delete one or more project folders and remove them from the database.
This action cannot be undone!`,
	Args: cobra.MinimumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		var projectsToDel []*db.Project
		for _, input := range args {
			project, err := utils.ResolveProject(input)
			if err != nil {
				return fmt.Errorf("failed to resolve project '%s': %w", input, err)
			}
			projectsToDel = append(projectsToDel, project)
		}

		// Confirm deletion
		var confirm bool
		msg := fmt.Sprintf("Delete project '%s' and all its files?", projectsToDel[0].Name)
		if len(projectsToDel) > 1 {
			msg = fmt.Sprintf("Delete %d projects and all their files?", len(projectsToDel))
		}

		prompt := &survey.Confirm{
			Message: msg,
			Default: false,
		}
		if err := survey.AskOne(prompt, &confirm); err != nil {
			return fmt.Errorf("cancelled: %w", err)
		}

		if !confirm {
			fmt.Println("Deletion cancelled.")
			return nil
		}

		// Delete from filesystem and database
		for _, project := range projectsToDel {
			if err := utils.DeleteProjectDir(project.Path); err != nil {
				fmt.Printf("⚠️  Warning: failed to delete project directory %s: %v\n", project.Path, err)
			}
			if err := db.DeleteProject(project.ID); err != nil {
				fmt.Printf("⚠️  Warning: failed to delete from database %s: %v\n", project.ID, err)
			}
			fmt.Printf("✓ Deleted project: %s\n", project.Name)
		}

		return nil
	},
}
