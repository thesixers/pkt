package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/genesix/pkt/internal/config"
	"github.com/genesix/pkt/internal/utils"
	"github.com/spf13/cobra"
)

var (
	openEditorFlag string
	openDirFlag    bool
)

var openCmd = &cobra.Command{
	Use:   "open [project | id | .]",
	Short: "Open a project in your configured editor or terminal",
	Long: `Open a project in your configured editor or navigate into its directory.

By default, opens the project in your configured editor.
Use --dir (-d) to cd into the project directory in your terminal shell.
Use --editor <editor> (-e) to specify a different editor than the default.

Examples:
  pkt open my-project
  pkt open my-project --editor cursor
  pkt open my-project -d
  pkt open . -d`,
	Args:              cobra.MaximumNArgs(1),
	ValidArgsFunction: completeProjectOrDot,
	RunE: func(cmd *cobra.Command, args []string) error {
		input := "."
		if len(args) > 0 {
			input = args[0]
		}

		// Resolve project
		project, err := utils.ResolveProject(input)
		if err != nil {
			return err
		}

		// Load config for editor command
		cfg, err := config.Load()
		if err != nil {
			return err
		}

		// Determine editor command
		editor := cfg.EditorCommand
		if openEditorFlag != "" {
			editor = openEditorFlag
		}

		// If --dir (-d) flag is passed, cd / enter project directory
		if openDirFlag {
			// Record target for potential shell integration wrappers
			if home, err := os.UserHomeDir(); err == nil {
				pktDir := filepath.Join(home, ".pkt")
				_ = os.MkdirAll(pktDir, 0755)
				_ = os.WriteFile(filepath.Join(pktDir, ".cd_target"), []byte(project.Path), 0644)
			}

			// If an editor was also explicitly specified, launch editor first
			if openEditorFlag != "" {
				editorCmd := exec.Command(editor, project.Path)
				_ = editorCmd.Start()
				fmt.Printf("✓ Opening %s in %s\n", project.Name, editor)
			}

			fmt.Printf("✓ Entering project directory: %s\n", project.Path)
			shell := os.Getenv("SHELL")
			if shell == "" {
				shell = "bash"
			}
			subShell := exec.Command(shell)
			subShell.Dir = project.Path
			subShell.Stdin = os.Stdin
			subShell.Stdout = os.Stdout
			subShell.Stderr = os.Stderr
			subShell.Env = append(os.Environ(), "PKT_PROJECT="+project.Name)
			if err := subShell.Run(); err != nil {
				if exitErr, ok := err.(*exec.ExitError); ok && exitErr.ExitCode() == 0 {
					return nil
				}
			}
			return nil
		}

		// Default behavior: execute editor command
		editorCmd := exec.Command(editor, project.Path)

		// If terminal-based editor (vim, nvim, nano), attach stdin/stdout
		isTerminalEditor := editor == "vim" || editor == "nvim" || editor == "nano" || editor == "vi" || editor == "emacs"
		if isTerminalEditor {
			editorCmd.Stdin = os.Stdin
			editorCmd.Stdout = os.Stdout
			editorCmd.Stderr = os.Stderr
			fmt.Printf("✓ Opening %s in %s...\n", project.Name, editor)
			return editorCmd.Run()
		}

		if err := editorCmd.Start(); err != nil {
			return fmt.Errorf("failed to open editor %q: %w", editor, err)
		}

		fmt.Printf("✓ Opening %s in %s\n", project.Name, editor)
		fmt.Printf("  Path: %s\n", project.Path)

		return nil
	},
}

func init() {
	openCmd.Flags().StringVarP(&openEditorFlag, "editor", "e", "", "Editor to open with (overrides default editor)")
	openCmd.Flags().BoolVarP(&openDirFlag, "dir", "d", false, "Navigate into the project directory in your terminal shell")
}
