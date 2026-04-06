package cmd

import (
	"os"

	"github.com/genesix/pkt/internal/ai"
	"github.com/genesix/pkt/internal/db"
	"github.com/spf13/cobra"
)

var chatProvider string

var chatCmd = &cobra.Command{
	Use:   "chat",
	Short: "Launch the interactive Autonomous Coding Agent loop",
	RunE: func(cmd *cobra.Command, args []string) error {
		cwd, _ := os.Getwd()
		var info string
		if project, err := db.GetProjectByPath(cwd); err == nil {
			info = extractProjectInfo(project.Path)
			if len(info) > 1500 {
				info = info[:1500] + "..."
			}
		}
		return ai.StartChatSession(chatProvider, info)
	},
}

func init() {
	chatCmd.Flags().StringVarP(&chatProvider, "provider", "p", "", "Specific AI Provider to use (openai, gemini, groq)")
	rootCmd.AddCommand(chatCmd)
}
