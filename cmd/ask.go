package cmd

import (
	"fmt"
	"os"
	"strings"

	"github.com/genesix/pkt/internal/ai"
	"github.com/genesix/pkt/internal/db"
	"github.com/spf13/cobra"
)

var askProvider string

var askCmd = &cobra.Command{
	Use:   "ask <question>",
	Short: "Ask the AI a question about your project",
	Args:  cobra.MinimumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		cwd, _ := os.Getwd()
		project, err := db.GetProjectByPath(cwd)

		sysPrompt := "You are a helpful coding assistant. Answer concisely without extra filler, jump straight to the exact code/commands."
		if err == nil {
			sysPrompt = fmt.Sprintf("You are a helpful assistant for a developer working on a %s project named '%s' managed by %s. Give exact terminal commands and extremely concise explanations.", project.Language, project.Name, project.PackageManager)

			info := extractProjectInfo(project.Path)
			if info != "" {
				if len(info) > 1500 {
					info = info[:1500] + "..."
				}
				sysPrompt += fmt.Sprintf("\n\nProject Context:\n%s", info)
			}
		}

		question := strings.Join(args, " ")
		fmt.Println("🤖 Thinking...\n")

		resp, err := ai.AskAI(sysPrompt, question, askProvider)
		if err != nil {
			return fmt.Errorf("failed to Ask AI: %w", err)
		}

		fmt.Printf("\033[36m%s\033[0m\n\n", resp)
		return nil
	},
}

func init() {
	askCmd.Flags().StringVarP(&askProvider, "provider", "p", "", "Specific AI Provider to use (openai, gemini, groq)")
	rootCmd.AddCommand(askCmd)
}
