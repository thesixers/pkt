package cmd

import (
	"fmt"
	"os"
	"strings"
	"text/tabwriter"

	"github.com/genesix/pkt/internal/db"
	"github.com/spf13/cobra"
)

var searchCmd = &cobra.Command{
	Use:   "search <query>",
	Short: "Search through tracked projects",
	Long: `Search through tracked projects by name or path.
Supports partial and case-insensitive matching.`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		query := strings.ToLower(args[0])

		// Get all projects
		projects, err := db.ListAllProjects()
		if err != nil {
			return fmt.Errorf("failed to list projects: %w", err)
		}

		// Filter projects by query (case-insensitive, partial match)
		var matches []*db.Project
		for _, project := range projects {
			nameLower := strings.ToLower(project.Name)
			pathLower := strings.ToLower(project.Path)

			if strings.Contains(nameLower, query) || strings.Contains(pathLower, query) {
				matches = append(matches, project)
			}
		}

		if len(matches) == 0 {
			fmt.Printf("No projects found matching '%s'.\n", args[0])
			return nil
		}

		// Display matches in the same format as list
		fmt.Printf("Found %d project(s) matching '%s':\n\n", len(matches), args[0])

		w := tabwriter.NewWriter(os.Stdout, 0, 0, 3, ' ', 0)
		_, _ = fmt.Fprintln(w, "NAME\tID\tPACKAGE MANAGER\tPATH")
		_, _ = fmt.Fprintln(w, "----\t--\t---------------\t----")

		for _, project := range matches {
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
