package cmd

import (
	"fmt"
	"os"
	"text/tabwriter"

	"github.com/genesix/pkt/internal/db"
	"github.com/spf13/cobra"
)

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List all tracked projects",
	Long:  `List all projects tracked by pkt with their details.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		// Get all projects
		projects, err := db.ListAllProjects()
		if err != nil {
			return fmt.Errorf("failed to list projects: %w", err)
		}

		if len(projects) == 0 {
			fmt.Println("No projects found.")
			fmt.Println("Create one with: pkt create <project-name>")
			return nil
		}

		// Create table writer
		w := tabwriter.NewWriter(os.Stdout, 0, 0, 3, ' ', 0)
		_, _ = fmt.Fprintln(w, "NAME\tID\tPACKAGE MANAGER\tPATH")
		_, _ = fmt.Fprintln(w, "----\t--\t---------------\t----")

		for _, project := range projects {
			_, _ = fmt.Fprintf(w, "%s\t%s\t%s\t%s\n",
				project.Name,
				project.ID,
				project.PackageManager,
				project.Path,
			)
		}

		_ = w.Flush()

		return nil
	},
}
