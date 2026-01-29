package cmd

import (
	"fmt"

	"github.com/AlecAivazis/survey/v2"
	"github.com/genesix/pkt/internal/db"
	"github.com/genesix/pkt/internal/utils"
	"github.com/spf13/cobra"
)

var deleteCmd = &cobra.Command{
	Use:   "delete <project | id>",
	Short: "Delete a project",
	Long: `Delete a project folder and remove it from the database.
This action cannot be undone!`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		input := args[0]

		// Resolve project
		project, err := utils.ResolveProject(input)
		if err != nil {
			return err
		}

		// Confirm deletion
		var confirm bool
		prompt := &survey.Confirm{
			Message: fmt.Sprintf("Delete project '%s' and all its files?", project.Name),
			Default: false,
		}
		if err := survey.AskOne(prompt, &confirm); err != nil {
			return fmt.Errorf("cancelled: %w", err)
		}

		if !confirm {
			fmt.Println("Deletion cancelled.")
			return nil
		}

		// Delete from filesystem
		if err := utils.DeleteProjectDir(project.Path); err != nil {
			return fmt.Errorf("failed to delete project directory: %w", err)
		}

		// Delete from database
		if err := db.DeleteProject(project.ID); err != nil {
			return fmt.Errorf("failed to delete from database: %w", err)
		}

		fmt.Printf("✓ Deleted project: %s\n", project.Name)

		return nil
	},
}
