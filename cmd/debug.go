package cmd

import (
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/genesix/pkt/internal/ai"
	"github.com/genesix/pkt/internal/db"
	"github.com/spf13/cobra"
)

var debugProvider string

var debugCmd = &cobra.Command{
	Use:   "debug [error_logs...]",
	Short: "Debug error logs with AI (supports stdin piping)",
	RunE: func(cmd *cobra.Command, args []string) error {
		var errorLog string

		stat, _ := os.Stdin.Stat()
		if (stat.Mode() & os.ModeCharDevice) == 0 {
			bytes, err := io.ReadAll(os.Stdin)
			if err != nil {
				return err
			}
			errorLog = string(bytes)
		} else {
			errorLog = strings.Join(args, " ")
		}

		if strings.TrimSpace(errorLog) == "" {
			return fmt.Errorf("provide an error log as arguments or via stdin pipe (e.g. 'cat error.log | pkt debug')")
		}

		cwd, _ := os.Getwd()
		project, err := db.GetProjectByPath(cwd)

		sysPrompt := "You are a debugging assistant."
		if err == nil {
			sysPrompt = fmt.Sprintf("You are deeply analyzing stack traces for a %s project managed by %s. Identify the bug precisely and provide exactly how to fix it with the correct code or command.", project.Language, project.PackageManager)
		}

		fmt.Println("🤖 Analyzing stack trace...\n")

		resp, err := ai.AskAI(sysPrompt, errorLog, debugProvider)
		if err != nil {
			return fmt.Errorf("failed to Ask AI: %w", err)
		}

		fmt.Printf("\033[33m%s\033[0m\n\n", resp)
		return nil
	},
}

func init() {
	debugCmd.Flags().StringVarP(&debugProvider, "provider", "p", "", "Specific AI Provider to use (openai, gemini, groq)")
	rootCmd.AddCommand(debugCmd)
}
