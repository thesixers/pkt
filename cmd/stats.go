package cmd

import (
	"fmt"
	"os"
	"text/tabwriter"

	"github.com/dustin/go-humanize"
	"github.com/genesix/pkt/internal/db"
	"github.com/genesix/pkt/internal/utils"
	"github.com/spf13/cobra"
)

var statsCmd = &cobra.Command{
	Use:   "stats",
	Short: "Show project statistics",
	Long:  "Calculate and display statistics for all tracked projects, such as disk space usage and language distribution.",
	RunE: func(cmd *cobra.Command, args []string) error {
		projects, err := db.ListAllProjects()
		if err != nil {
			return err
		}

		if len(projects) == 0 {
			fmt.Println("No projects tracked. Run 'pkt create' or 'pkt init' to get started.")
			return nil
		}

		fmt.Println("📊 Calculating statistics...")

		var totalSize int64
		langCounts := make(map[string]int)
		langSize := make(map[string]int64)

		for _, p := range projects {
			langCounts[p.Language]++
			size, _ := utils.GetDirSize(p.Path)
			totalSize += size
			langSize[p.Language] += size
		}

		fmt.Println("\n\033[1mWorkspace Summary:\033[0m")
		fmt.Printf("  Total Projects: %d\n", len(projects))
		fmt.Printf("  Total Space: \033[36m%s\033[0m\n", humanize.Bytes(uint64(totalSize)))

		fmt.Println("\n\033[1mLanguage Breakdown:\033[0m")
		w := tabwriter.NewWriter(os.Stdout, 0, 0, 3, ' ', 0)
		fmt.Fprintln(w, "LANGUAGE\tCOUNT\tSIZE")
		fmt.Fprintln(w, "--------\t-----\t----")
		for lang, count := range langCounts {
			fmt.Fprintf(w, "%s\t%d\t%s\n", lang, count, humanize.Bytes(uint64(langSize[lang])))
		}
		w.Flush()

		return nil
	},
}

func init() {
	rootCmd.AddCommand(statsCmd)
}
