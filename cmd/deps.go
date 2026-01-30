package cmd

import (
	"fmt"
	"os"
	"text/tabwriter"

	"github.com/genesix/pkt/internal/db"
	"github.com/genesix/pkt/internal/utils"
	"github.com/spf13/cobra"
)

var depsCmd = &cobra.Command{
	Use:   "deps [project | id | .]",
	Short: "List dependencies for a project",
	Long: `List all dependencies for a project.
If no argument is provided, uses the current directory.`,
	Args: cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		var project *db.Project
		var err error

		// Determine which project to use
		if len(args) == 0 {
			// Use current directory
			cwd, err := os.Getwd()
			if err != nil {
				return fmt.Errorf("failed to get current directory: %w", err)
			}
			project, err = db.GetProjectByPath(cwd)
			if err != nil {
				return fmt.Errorf("not in a tracked project")
			}
		} else {
			// Resolve from argument
			project, err = utils.ResolveProject(args[0])
			if err != nil {
				return err
			}
		}

		// Parse dependencies based on project language
		deps, err := utils.ParseDependencies(project.Path, project.Language)
		if err != nil {
			return fmt.Errorf("failed to parse dependencies: %w", err)
		}

		// Sync to database
		if err := db.SyncDependencies(project.ID, deps); err != nil {
			return fmt.Errorf("failed to sync dependencies: %w", err)
		}

		// Get dependencies from database
		dbDeps, err := db.GetDependencies(project.ID)
		if err != nil {
			return fmt.Errorf("failed to get dependencies: %w", err)
		}

		if len(dbDeps) == 0 {
			fmt.Printf("No dependencies found for %s\n", project.Name)
			return nil
		}

		// Display dependencies
		fmt.Printf("Dependencies for %s:\n\n", project.Name)

		w := tabwriter.NewWriter(os.Stdout, 0, 0, 3, ' ', 0)
		_, _ = fmt.Fprintln(w, "NAME\tVERSION\tTYPE")
		_, _ = fmt.Fprintln(w, "----\t-------\t----")

		for _, dep := range dbDeps {
			_, _ = fmt.Fprintf(w, "%s\t%s\t%s\n",
				dep.Name,
				dep.Version,
				dep.DepType,
			)
		}

		_ = w.Flush()

		return nil
	},
}
