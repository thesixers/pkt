package cmd

import (
	"fmt"

	"github.com/genesix/pkt/internal/config"
	"github.com/genesix/pkt/internal/db"
	"github.com/genesix/pkt/internal/utils"
	"github.com/spf13/cobra"
)

var createCmd = &cobra.Command{
	Use:   "create <project-name>",
	Short: "Create a new project",
	Long: `Create a new project folder inside the projects root and track it.
The project will be assigned a unique ID and tracked in the database.`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		projectName := args[0]

		// Load config
		cfg, err := config.Load()
		if err != nil {
			return err
		}

		// Create project directory
		projectPath, err := utils.CreateProjectDir(cfg.ProjectsRoot, projectName)
		if err != nil {
			return fmt.Errorf("failed to create project directory: %w", err)
		}

		// Generate project ID
		projectID := utils.GenerateID()

		// Insert into database
		project, err := db.CreateProject(projectID, projectName, projectPath, cfg.DefaultPM)
		if err != nil {
			// Clean up directory if database insert fails
			_ = utils.DeleteProjectDir(projectPath)
			return fmt.Errorf("failed to create project in database: %w", err)
		}

		fmt.Printf("✓ Created project: %s\n", projectName)
		fmt.Printf("  ID: %s\n", project.ID)
		fmt.Printf("  Path: %s\n", project.Path)
		fmt.Printf("  Package Manager: %s\n", project.PackageManager)

		return nil
	},
}
