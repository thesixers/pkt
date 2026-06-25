package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/genesix/pkt/internal/utils"
	"github.com/spf13/cobra"
)

var cpCmd = &cobra.Command{
	Use:   "cp <source> <destination>",
	Short: "Copy files or directories between projects",
	Long: `Copy files or directories between tracked projects seamlessly using the @ syntax.
	
Examples:
  pkt cp @api-server/config.json @frontend/config.json
  pkt cp ./local.txt @my-project/assets/
  pkt cp @legacy-app/data ./data`,
	Args: cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		srcPath, err := utils.ResolvePktPath(args[0])
		if err != nil {
			return err
		}

		dstPath, err := utils.ResolvePktPath(args[1])
		if err != nil {
			return err
		}

		if _, err := os.Stat(srcPath); err != nil {
			return fmt.Errorf("source does not exist: %s", srcPath)
		}

		// If destination is an existing directory, append the source filename
		if dstInfo, err := os.Stat(dstPath); err == nil && dstInfo.IsDir() {
			dstPath = filepath.Join(dstPath, filepath.Base(srcPath))
		}

		fmt.Printf("Copying to %s...\n", dstPath)
		if err := utils.CopyItem(srcPath, dstPath); err != nil {
			return fmt.Errorf("failed to copy: %w", err)
		}

		fmt.Println("✓ Copied successfully")
		return nil
	},
}

func init() {
	rootCmd.AddCommand(cpCmd)
}
