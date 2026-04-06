package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/AlecAivazis/survey/v2"
	"github.com/dustin/go-humanize"
	"github.com/genesix/pkt/internal/db"
	"github.com/genesix/pkt/internal/utils"
	"github.com/spf13/cobra"
)

var cleanCmd = &cobra.Command{
	Use:   "clean",
	Short: "Prune heavy project cache folders",
	Long:  "Find and safely delete bulky cache/build folders (like node_modules, target, venv) across all tracked projects.",
	RunE: func(cmd *cobra.Command, args []string) error {
		projects, err := db.ListAllProjects()
		if err != nil {
			return err
		}

		fmt.Println("🧹 Scanning discrete cache folders... This may take a moment.")

		type target struct {
			ProjectName string
			Path        string
			Size        int64
		}

		var targets []target
		var totalSaved int64

		for _, p := range projects {
			var dirs []string
			lang := strings.ToLower(p.Language)
			switch {
			case strings.Contains(lang, "javascript") || strings.Contains(lang, "node"):
				dirs = []string{"node_modules", "dist", "build", ".next"}
			case strings.Contains(lang, "python"):
				dirs = []string{"venv", ".venv", "__pycache__"}
			case strings.Contains(lang, "rust"):
				dirs = []string{"target"}
			}

			for _, d := range dirs {
				candidate := filepath.Join(p.Path, d)
				if info, err := os.Stat(candidate); err == nil && info.IsDir() {
					size, _ := utils.GetDirSize(candidate)
					targets = append(targets, target{
						ProjectName: p.Name,
						Path:        candidate,
						Size:        size,
					})
					totalSaved += size
				}
			}
		}

		if len(targets) == 0 {
			fmt.Println("✨ Workspace is already pristine. No bloat folders found.")
			return nil
		}

		fmt.Println("\nCache directories found:")
		for _, t := range targets {
			fmt.Printf("  • %s (%s): %s\n", t.ProjectName, filepath.Base(t.Path), humanize.Bytes(uint64(t.Size)))
		}

		fmt.Printf("\nTotal recoverable space: \033[1;32m%s\033[0m\n", humanize.Bytes(uint64(totalSaved)))

		var confirm bool
		prompt := &survey.Confirm{
			Message: "Do you want to permanently delete these directories?",
			Default: false,
		}
		if err := survey.AskOne(prompt, &confirm); err != nil {
			return fmt.Errorf("cancelled: %w", err)
		}

		if !confirm {
			fmt.Println("Skipping clean.")
			return nil
		}

		fmt.Println()
		for _, t := range targets {
			fmt.Printf("Removing %s... ", t.Path)
			if err := os.RemoveAll(t.Path); err != nil {
				fmt.Printf("❌ Failed: %v\n", err)
			} else {
				fmt.Println("✓")
			}
		}

		fmt.Printf("\n✨ Successfully reclaimed \033[1;32m%s\033[0m of disk space!\n", humanize.Bytes(uint64(totalSaved)))

		return nil
	},
}

func init() {
	rootCmd.AddCommand(cleanCmd)
}
