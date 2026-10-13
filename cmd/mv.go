package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/genesix/pkt/internal/utils"
	"github.com/spf13/cobra"
)

var mvCmd = &cobra.Command{
	Use:   "mv <source> <destination>",
	Short: "Move files or directories between projects",
	Long: `Move files or directories between tracked projects seamlessly using the @ syntax.
	
Examples:
  pkt mv @api-server/old.json @frontend/new.json
  pkt mv ./local.txt @my-project/assets/`,
	Args:              cobra.ExactArgs(2),
	ValidArgsFunction: completePktPath,
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

		// Ensure destination parent directory exists
		if err := os.MkdirAll(filepath.Dir(dstPath), 0755); err != nil {
			return fmt.Errorf("failed to create destination directory: %w", err)
		}

		fmt.Printf("Moving to %s...\n", dstPath)
		
		if err := os.Rename(srcPath, dstPath); err != nil {
			// Fallback to Copy + Delete for cross-device moves
			if copyErr := utils.CopyItem(srcPath, dstPath); copyErr != nil {
				return fmt.Errorf("failed to move (cross-device copy failed): %w", copyErr)
			}
			_ = os.RemoveAll(srcPath)
		}

		fmt.Println("✓ Moved successfully")
		return nil
	},
}

func init() {
	rootCmd.AddCommand(mvCmd)
}
