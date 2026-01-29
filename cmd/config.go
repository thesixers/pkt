package cmd

import (
	"fmt"
	"os/exec"

	"github.com/genesix/pkt/internal/config"
	"github.com/spf13/cobra"
)

var configCmd = &cobra.Command{
	Use:   "config [key] [value]",
	Short: "View or update pkt configuration",
	Long: `View or update pkt configuration settings.

Without arguments, displays current configuration.
With a key and value, updates that setting.

Available keys:
  editor    - Editor command (e.g., code, cursor, vim)
  pm        - Default package manager (pnpm, npm, bun)

Examples:
  pkt config                    # Show current config
  pkt config editor cursor      # Change editor to cursor
  pkt config pm npm             # Change default PM to npm`,
	Args: cobra.MaximumNArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := config.Load()
		if err != nil {
			return err
		}

		// No args: show current config
		if len(args) == 0 {
			fmt.Println("Current configuration:")
			fmt.Printf("  projects_root: %s\n", cfg.ProjectsRoot)
			fmt.Printf("  editor:        %s\n", cfg.EditorCommand)
			fmt.Printf("  pm:            %s\n", cfg.DefaultPM)
			fmt.Printf("\nConfig file: ~/.pkt/config.json\n")
			return nil
		}

		// Need both key and value to update
		if len(args) == 1 {
			key := args[0]
			switch key {
			case "editor":
				fmt.Printf("editor: %s\n", cfg.EditorCommand)
			case "pm":
				fmt.Printf("pm: %s\n", cfg.DefaultPM)
			case "projects_root":
				fmt.Printf("projects_root: %s\n", cfg.ProjectsRoot)
			default:
				return fmt.Errorf("unknown config key: %s\nAvailable keys: editor, pm, projects_root", key)
			}
			return nil
		}

		// Update config
		key := args[0]
		value := args[1]

		switch key {
		case "editor":
			// Verify editor command exists
			if _, err := exec.LookPath(value); err != nil {
				fmt.Printf("⚠️  Warning: '%s' command not found in PATH\n", value)
			}
			cfg.EditorCommand = value
			if err := config.Save(cfg); err != nil {
				return fmt.Errorf("failed to save config: %w", err)
			}
			fmt.Printf("✓ Editor set to: %s\n", value)

		case "pm":
			// Validate package manager
			validPMs := map[string]bool{"pnpm": true, "npm": true, "bun": true}
			if !validPMs[value] {
				return fmt.Errorf("invalid package manager: %s\nSupported: pnpm, npm, bun", value)
			}
			// Check if available
			if _, err := exec.LookPath(value); err != nil {
				fmt.Printf("⚠️  Warning: '%s' is not installed\n", value)
			}
			cfg.DefaultPM = value
			if err := config.Save(cfg); err != nil {
				return fmt.Errorf("failed to save config: %w", err)
			}
			fmt.Printf("✓ Default package manager set to: %s\n", value)

		case "projects_root":
			return fmt.Errorf("projects_root cannot be changed after setup\nRecreate config with 'pkt start' if needed")

		default:
			return fmt.Errorf("unknown config key: %s\nAvailable keys: editor, pm", key)
		}

		return nil
	},
}

func init() {
	rootCmd.AddCommand(configCmd)
}
