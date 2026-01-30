package cmd

import (
	"fmt"

	"github.com/AlecAivazis/survey/v2"
	"github.com/genesix/pkt/internal/config"
	"github.com/genesix/pkt/internal/db"
	"github.com/genesix/pkt/internal/lang"
	"github.com/genesix/pkt/internal/pm"
	"github.com/genesix/pkt/internal/utils"
	"github.com/spf13/cobra"
)

var createLang string

var createCmd = &cobra.Command{
	Use:   "create <project-name>",
	Short: "Create a new project",
	Long: `Create a new project folder inside the projects root and track it.
The project will be assigned a unique ID and tracked in the database.

Supported languages:
  js   - JavaScript/Node.js (npm, pnpm, bun)
  py   - Python (pip, poetry, uv)
  go   - Go (go mod)
  rs   - Rust (cargo)

Examples:
  pkt create my-app           # Prompts for language
  pkt create my-api -l js     # JavaScript project
  pkt create my-cli -l py     # Python project
  pkt create my-tool -l go    # Go project
  pkt create my-lib -l rs     # Rust project`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		projectName := args[0]

		// Load config
		cfg, err := config.Load()
		if err != nil {
			return err
		}

		// Determine language
		language := createLang
		if language == "" {
			// Prompt user to select language
			langOptions := []string{
				"js - JavaScript/Node.js",
				"py - Python",
				"go - Go",
				"rs - Rust",
			}
			var selected string
			prompt := &survey.Select{
				Message: "Select project language:",
				Options: langOptions,
				Default: langOptions[0],
			}
			if err := survey.AskOne(prompt, &selected); err != nil {
				return fmt.Errorf("cancelled: %w", err)
			}
			// Extract language code from selection
			for i := 0; i < len(selected); i++ {
				if selected[i] == ' ' {
					language = selected[:i]
					break
				}
			}
		}

		// Validate and normalize language (js -> javascript, py -> python, etc.)
		langImpl, err := lang.Get(language)
		if err != nil {
			return err
		}

		// Use the full language name for database storage
		fullLangName := langImpl.Name()

		// Determine package manager
		packageManager := cfg.DefaultPM
		availablePMs := pm.ListForLanguage(fullLangName)
		if len(availablePMs) == 0 {
			return fmt.Errorf("no package managers available for %s", langImpl.DisplayName())
		}

		// Check if default PM is valid for this language
		validPM := false
		for _, p := range availablePMs {
			if p == packageManager {
				validPM = true
				break
			}
		}
		if !validPM {
			// Use language's default PM
			packageManager = langImpl.DefaultPackageManager()
			// Check if it's available
			if !pm.CheckAvailability(packageManager) {
				// Use first available
				packageManager = availablePMs[0]
			}
		}

		// Create project directory
		projectPath, err := utils.CreateProjectDir(cfg.ProjectsRoot, projectName)
		if err != nil {
			return fmt.Errorf("failed to create project directory: %w", err)
		}

		// Initialize project with package manager
		pmImpl, err := pm.Get(fullLangName, packageManager)
		if err != nil {
			_ = utils.DeleteProjectDir(projectPath)
			return fmt.Errorf("failed to get package manager: %w", err)
		}

		if err := pmImpl.Init(projectPath); err != nil {
			_ = utils.DeleteProjectDir(projectPath)
			return fmt.Errorf("failed to initialize project: %w", err)
		}

		// Generate project ID
		projectID := utils.GenerateID()

		// Insert into database with full language name
		project, err := db.CreateProject(projectID, projectName, projectPath, fullLangName, packageManager)
		if err != nil {
			// Clean up directory if database insert fails
			_ = utils.DeleteProjectDir(projectPath)
			return fmt.Errorf("failed to create project in database: %w", err)
		}

		fmt.Printf("✓ Created %s project: %s\n", langImpl.DisplayName(), projectName)
		fmt.Printf("  ID: %s\n", project.ID)
		fmt.Printf("  Path: %s\n", project.Path)
		fmt.Printf("  Language: %s\n", langImpl.DisplayName())
		fmt.Printf("  Package Manager: %s\n", project.PackageManager)

		return nil
	},
}

func init() {
	createCmd.Flags().StringVarP(&createLang, "lang", "l", "", "Project language (js, py, go, rs)")
}
