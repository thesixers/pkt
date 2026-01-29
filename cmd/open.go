package cmd

import (
	"fmt"
	"os/exec"

	"github.com/genesix/pkt/internal/config"
	"github.com/genesix/pkt/internal/utils"
	"github.com/spf13/cobra"
)

var openCmd = &cobra.Command{
	Use:   "open <project | id | .>",
	Short: "Open a project in your configured editor",
	Long: `Open a project in your configured editor.
You can specify the project by name, ID, or use "." for the current directory.`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		input := args[0]

		// Resolve project
		project, err := utils.ResolveProject(input)
		if err != nil {
			return err
		}

		// Load config for editor command
		cfg, err := config.Load()
		if err != nil {
			return err
		}

		// Execute editor command
		editorCmd := exec.Command(cfg.EditorCommand, project.Path)
		if err := editorCmd.Start(); err != nil {
			return fmt.Errorf("failed to open editor: %w", err)
		}

		fmt.Printf("✓ Opening %s in %s\n", project.Name, cfg.EditorCommand)
		fmt.Printf("  Path: %s\n", project.Path)

		return nil
	},
}
