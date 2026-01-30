package cmd

import (
	"fmt"
	"os"
	"text/tabwriter"

	"github.com/genesix/pkt/internal/db"
	"github.com/genesix/pkt/internal/lang"
	"github.com/spf13/cobra"
)

var listLangFilter string

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List all tracked projects",
	Long: `List all projects tracked by pkt with their details.

Use --lang to filter by language (js, py, go, rs).

Examples:
  pkt list            # All projects
  pkt list -l js      # JavaScript projects only
  pkt list -l py      # Python projects only`,
	RunE: func(cmd *cobra.Command, args []string) error {
		var projects []*db.Project
		var err error

		// Filter by language if specified
		if listLangFilter != "" {
			// Normalize short codes to full names (js -> javascript)
			fullName := lang.NormalizeName(listLangFilter)
			projects, err = db.GetProjectsByLanguage(fullName)
		} else {
			projects, err = db.ListAllProjects()
		}

		if err != nil {
			return fmt.Errorf("failed to list projects: %w", err)
		}

		if len(projects) == 0 {
			if listLangFilter != "" {
				fmt.Printf("No %s projects found.\n", listLangFilter)
			} else {
				fmt.Println("No projects found.")
			}
			fmt.Println("Create one with: pkt create <project-name>")
			return nil
		}

		// Create table writer
		w := tabwriter.NewWriter(os.Stdout, 0, 0, 3, ' ', 0)
		_, _ = fmt.Fprintln(w, "NAME\tLANG\tPM\tID\tPATH")
		_, _ = fmt.Fprintln(w, "----\t----\t--\t--\t----")

		for _, project := range projects {
			// Convert full language name to short code for display
			shortLang := langToShort(project.Language)
			_, _ = fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%s\n",
				project.Name,
				shortLang,
				project.PackageManager,
				project.ID,
				project.Path,
			)
		}

		_ = w.Flush()

		return nil
	},
}

// langToShort converts full language name to short code
func langToShort(language string) string {
	switch language {
	case "javascript":
		return "js"
	case "python":
		return "py"
	case "rust":
		return "rs"
	default:
		return language
	}
}

func init() {
	listCmd.Flags().StringVarP(&listLangFilter, "lang", "l", "", "Filter by language (js, py, go, rs)")
}
