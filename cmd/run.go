package cmd

import (
	"fmt"
	"os"

	"github.com/genesix/pkt/internal/db"
	"github.com/genesix/pkt/internal/pm"
	"github.com/spf13/cobra"
)

var runCmd = &cobra.Command{
	Use:   "run <script> [args...]",
	Short: "Run a script in the current project",
	Long: `Run a script or command using the project's package manager.
Must be run inside a tracked project folder.

Supported scripts by language:
  JavaScript: Any script from package.json (e.g., dev, build, test)
  Python:     test (pytest), or any .py file
  Go:         run, build, test, or any .go file
  Rust:       run, build, test, or binary name

Examples:
  pkt run dev              # npm/pnpm run dev
  pkt run test             # Run tests for any language
  pkt run build            # Build project
  pkt run main.py          # Run Python file
  pkt run test -- -v       # Pass args to test command`,
	Args: cobra.MinimumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		script := args[0]
		scriptArgs := args[1:]

		// Get current directory
		cwd, err := os.Getwd()
		if err != nil {
			return fmt.Errorf("failed to get current directory: %w", err)
		}

		// Get project from current directory
		project, err := db.GetProjectByPath(cwd)
		if err != nil {
			return fmt.Errorf("not in a tracked project. Run 'pkt init .' first")
		}

		// Get package manager
		packageManager, err := pm.Get(project.Language, project.PackageManager)
		if err != nil {
			return err
		}

		// Run the script
		return packageManager.Run(cwd, script, scriptArgs)
	},
}

func init() {
	rootCmd.AddCommand(runCmd)
}
