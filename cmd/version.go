package cmd

import (
	"fmt"

	"github.com/genesix/pkt/internal/version"
	"github.com/spf13/cobra"
)

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print the version number of pkt",
	Long:  `All software has versions. This is pkt's.`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("pkt version %s\n", version.Version)
	},
}

func init() {
	rootCmd.AddCommand(versionCmd)
}
