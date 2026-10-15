package cmd

import (
	"fmt"

	"github.com/genesix/pkt/internal/config"
	"github.com/genesix/pkt/internal/db"
	"github.com/genesix/pkt/internal/version"
	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:     "pkt",
	Version: version.Version,
	Short:   "Project Kit - A universal project manager for JS, Python, Go, and Rust",
	Long: `pkt (Project Kit) is a universal project manager and dependency tracker.

Supported Languages:
  • JavaScript (npm, pnpm, bun)
  • Python (pip, poetry, uv) — auto venv management
  • Go (go mod)
  • Rust (cargo)

Prerequisites:
  • Git — required for 'pkt clone'
  • Node.js — for JavaScript projects
  • Python 3.8+ — for Python projects
  • Go 1.18+ — for Go projects
  • Rust/Cargo — for Rust projects

Quick Start:
  pkt start              Initialize pkt
  pkt create my-app      Create new project
  pkt init .             Track existing project
  pkt add react flask    Add dependencies
  pkt run dev            Run scripts
  pkt clone <url>        Clone and track repo`,
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		// Skip config check for commands that don't need initialization
		if isNoConfigCmd(cmd) {
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

// isNoConfigCmd returns true if the command can run without initialized pkt config or database.
func isNoConfigCmd(cmd *cobra.Command) bool {
	if cmd == nil {
		return false
	}
	name := cmd.Name()
	if name == "start" || name == "version" || name == "help" {
		return true
	}
	for c := cmd; c != nil; c = c.Parent() {
		if c.Name() == "completion" {
			return true
		}
	}
	return false
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
	rootCmd.AddCommand(renameCmd)
}

