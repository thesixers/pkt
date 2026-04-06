package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"text/tabwriter"

	"github.com/genesix/pkt/internal/db"
	"github.com/spf13/cobra"
)

var statusCmd = &cobra.Command{
	Use:   "status",
	Short: "Check global git status",
	Long:  "Check the git status of all tracked projects to find uncommitted changes or missing pushes.",
	RunE: func(cmd *cobra.Command, args []string) error {
		projects, err := db.ListAllProjects()
		if err != nil {
			return err
		}

		if len(projects) == 0 {
			fmt.Println("No projects tracked.")
			return nil
		}

		fmt.Println("🔍 Scanning repositories...")
		fmt.Println()

		w := tabwriter.NewWriter(os.Stdout, 0, 0, 3, ' ', 0)
		fmt.Fprintln(w, "PROJECT\tSTATUS\tBRANCH")
		fmt.Fprintln(w, "-------\t------\t------")

		var hasRepos bool
		for _, p := range projects {
			// Check if .git exists
			gitDir := filepath.Join(p.Path, ".git")
			if _, err := os.Stat(gitDir); os.IsNotExist(err) {
				continue
			}
			hasRepos = true

			// Check git status
			gitStatusCmd := exec.Command("git", "status", "--porcelain")
			gitStatusCmd.Dir = p.Path
			out, err := gitStatusCmd.Output()
			statusStr := "\033[32mClean\033[0m"
			if err != nil {
				statusStr = "\033[31mError\033[0m"
			} else if len(strings.TrimSpace(string(out))) > 0 {
				statusStr = "\033[33mModified\033[0m"
			}

			// Check git branch
			gitBranchCmd := exec.Command("git", "branch", "--show-current")
			gitBranchCmd.Dir = p.Path
			branchOut, _ := gitBranchCmd.Output()
			branchStr := strings.TrimSpace(string(branchOut))
			if branchStr == "" {
				branchStr = "unknown"
			}

			fmt.Fprintf(w, "%s\t%s\t%s\n", p.Name, statusStr, branchStr)
		}

		if !hasRepos {
			fmt.Println("No git repositories found in tracked projects.")
			return nil
		}

		w.Flush()
		return nil
	},
}

func init() {
	rootCmd.AddCommand(statusCmd)
}
