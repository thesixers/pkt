package cmd

import (
	"fmt"
	"os"
	"strings"

	"github.com/genesix/pkt/internal/ai"
	"github.com/genesix/pkt/internal/db"
	"github.com/spf13/cobra"
)

var generateProvider string

var generateCmd = &cobra.Command{
	Use:   "generate <feature>",
	Short: "Generate starter code or features using AI",
	Args:  cobra.MinimumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		cwd, _ := os.Getwd()
		project, err := db.GetProjectByPath(cwd)
		if err != nil {
			return fmt.Errorf("must be in a tracked pkt project folder to generate specific code")
		}

		desc := strings.Join(args, " ")
		fmt.Println("🤖 Generating...\n")

		sysPrompt := fmt.Sprintf("You are a senior engineer scaffolding code for a %s project named '%s' using %s. Write the requested code completely and accurately. If returning files, state the filename at the top. Minimize conversational filler.", project.Language, project.Name, project.PackageManager)

		info := extractProjectInfo(project.Path)
		if info != "" {
			if len(info) > 1500 {
				info = info[:1500] + "..."
			}
			sysPrompt += fmt.Sprintf("\n\nProject Context:\n%s", info)
		}

		resp, err := ai.AskAI(sysPrompt, desc, generateProvider)
		if err != nil {
			return fmt.Errorf("failed to Ask AI: %w", err)
		}

		fmt.Printf("\033[32m%s\033[0m\n\n", resp)
		return nil
	},
}

func init() {
	generateCmd.Flags().StringVarP(&generateProvider, "provider", "p", "", "Specific AI Provider to use (openai, gemini, groq)")
	rootCmd.AddCommand(generateCmd)
}
