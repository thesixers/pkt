package cmd

import (
	"fmt"
	"os"
	"regexp"

	"github.com/dustin/go-humanize"
	"github.com/genesix/pkt/internal/db"
	"github.com/genesix/pkt/internal/lang"
	"github.com/genesix/pkt/internal/utils"
	"github.com/jedib0t/go-pretty/v6/table"
	"github.com/spf13/cobra"
)

var (
	listLangFilter string
	listNameFilter string
	listAllFlag    bool
)

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

		if listNameFilter != "" {
			nameRegex, err := regexp.Compile("(?i)" + listNameFilter)
			if err != nil {
				return fmt.Errorf("invalid filter regex: %w", err)
			}
			var filtered []*db.Project
			for _, p := range projects {
				if nameRegex.MatchString(p.Name) {
					filtered = append(filtered, p)
				}
			}
			projects = filtered
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

		t := table.NewWriter()
		t.SetOutputMirror(os.Stdout)
		t.SetStyle(table.StyleRounded)

		if listAllFlag {
			t.AppendHeader(table.Row{"NAME", "LANG", "PM", "ID", "SIZE", "PATH"})
		} else {
			t.AppendHeader(table.Row{"NAME", "LANG", "PATH"})
		}

		for _, project := range projects {
			shortLang := langToShort(project.Language)

			if listAllFlag {
				sizeBytes, _ := utils.GetDirSize(project.Path)
				sizeStr := humanize.Bytes(uint64(sizeBytes))
				t.AppendRow(table.Row{
					project.Name,
					shortLang,
					project.PackageManager,
					project.ID,
					sizeStr,
					utils.ShortPath(project.Path),
				})
			} else {
				t.AppendRow(table.Row{
					project.Name,
					shortLang,
					utils.ShortPath(project.Path),
				})
			}
		}

		t.Render()

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
	listCmd.Flags().StringVarP(&listNameFilter, "filter", "f", "", "Filter projects by regex on their name")
	listCmd.Flags().BoolVarP(&listAllFlag, "all", "a", false, "Show all details including ID and Package Manager")
	_ = listCmd.RegisterFlagCompletionFunc("lang", completeLangFlag)
}
