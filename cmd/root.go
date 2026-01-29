package cmd

import (
	"fmt"

	"github.com/genesix/pkt/internal/config"
	"github.com/genesix/pkt/internal/db"
	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "pkt",
	Short: "A cross-platform project manager for JavaScript/Node.js projects",
	Long: `pkt is a project manager and dependency tracker that sits on top of
package managers like pnpm, npm, and bun. It provides a unified interface
for managing all your JavaScript projects.`,
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		// Skip config check for start command
		if cmd.Name() == "start" {
			return nil
		}

		// Check if config exists and is initialized
		cfg, err := config.Load()
		if err != nil || !cfg.Initialized {
			return fmt.Errorf("pkt has not been initialized, please run 'pkt start' first")
		}

		// Set database configuration
		db.SetConfig(cfg)

		// Connect to database
		if err := db.Connect(); err != nil {
			return fmt.Errorf("failed to connect to database: %w", err)
		}

		return nil
	},
	PersistentPostRunE: func(cmd *cobra.Command, args []string) error {
		// Close database connection
		return db.Close()
	},
}

// Execute runs the root command
func Execute() error {
	return rootCmd.Execute()
}

func init() {
	// Add all subcommands
	rootCmd.AddCommand(startCmd)
	rootCmd.AddCommand(createCmd)
	rootCmd.AddCommand(openCmd)
	rootCmd.AddCommand(listCmd)
	rootCmd.AddCommand(deleteCmd)
	rootCmd.AddCommand(addCmd)
	rootCmd.AddCommand(removeCmd)
	rootCmd.AddCommand(depsCmd)
	rootCmd.AddCommand(pmCmd)
}
