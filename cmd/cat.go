package cmd

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"strings"
	"unicode/utf8"

	"github.com/genesix/pkt/internal/utils"
	"github.com/spf13/cobra"
)

type catOptions struct {
	number   bool
	start    int
	end      int
	head     int
	tail     int
	noHeader bool
	force    bool
}

var catFlags catOptions

var catCmd = &cobra.Command{
	Use:   "cat <file...>",
	Short: "Print text files",
	Long: `Print text files from normal paths or tracked project references.

Examples:
  pkt cat README.md
  pkt cat -n @api/src/main.go
  pkt cat --start 20 --end 50 @api/README.md`,
	Args:              cobra.MinimumNArgs(1),
	ValidArgsFunction: completePktPath,
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := validateCatOptions(catFlags); err != nil {
			return err
		}
		var failures []string
		for index, input := range args {
			path, err := utils.ResolvePktPath(input)
			if err == nil {
				err = writeCatFile(cmd.OutOrStdout(), path, input, len(args) > 1 && !catFlags.noHeader, catFlags)
			}
			if err != nil {
				failures = append(failures, fmt.Sprintf("%s: %v", input, err))
			}
			if index < len(args)-1 && len(args) > 1 && !catFlags.noHeader {
				_, _ = fmt.Fprintln(cmd.OutOrStdout())
			}
		}
		if len(failures) > 0 {
			return fmt.Errorf("could not read %d file(s):\n%s", len(failures), strings.Join(failures, "\n"))
		}
		return nil
	},
}

func validateCatOptions(options catOptions) error {
	if options.start < 0 || options.end < 0 || options.head < 0 || options.tail < 0 {
		return fmt.Errorf("line and count options must be positive")
	}
	if options.head > 0 && options.tail > 0 {
		return fmt.Errorf("--head and --tail cannot be used together")
	}
	if options.tail > 0 && (options.start > 0 || options.end > 0) {
		return fmt.Errorf("--tail cannot be combined with --start or --end")
	}
	if options.start > 0 && options.end > 0 && options.end < options.start {
		return fmt.Errorf("--end must be greater than or equal to --start")
	}
	return nil
}

func writeCatFile(writer io.Writer, path, displayName string, header bool, options catOptions) error {
	info, err := os.Stat(path)
	if err != nil {
		return err
	}
	if info.IsDir() {
		return fmt.Errorf("is a directory; use pkt ls")
	}
	data, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	if !options.force && isBinary(data) {
		return fmt.Errorf("appears to be binary; use --force to print it")
	}
	if header {
		if _, err := fmt.Fprintf(writer, "==> %s <==\n", displayName); err != nil {
			return err
		}
	}
	if options.number || options.start > 0 || options.end > 0 || options.head > 0 || options.tail > 0 {
		return writeSelectedLines(writer, data, options)
	}
	_, err = writer.Write(data)
	return err
}

func isBinary(data []byte) bool {
	return bytes.IndexByte(data, 0) >= 0 || !utf8.Valid(data)
}

func writeSelectedLines(writer io.Writer, data []byte, options catOptions) error {
	lines := strings.Split(string(data), "\n")
	if len(lines) > 0 && lines[len(lines)-1] == "" {
		lines = lines[:len(lines)-1]
	}
	start := 1
	end := len(lines)
	if options.start > 0 {
		start = options.start
	}
	if options.end > 0 && options.end < end {
		end = options.end
	}
	if options.head > 0 && start+options.head-1 < end {
		end = start + options.head - 1
	}
	if options.tail > 0 {
		start = len(lines) - options.tail + 1
		if start < 1 {
			start = 1
		}
	}
	for lineNumber := start; lineNumber <= end && lineNumber <= len(lines); lineNumber++ {
		if options.number {
			if _, err := fmt.Fprintf(writer, "%6d\t%s\n", lineNumber, lines[lineNumber-1]); err != nil {
				return err
			}
		} else if _, err := fmt.Fprintln(writer, lines[lineNumber-1]); err != nil {
			return err
		}
	}
	return nil
}

func init() {
	rootCmd.AddCommand(catCmd)
	catCmd.Flags().BoolVarP(&catFlags.number, "number", "n", false, "Show line numbers")
	catCmd.Flags().IntVar(&catFlags.start, "start", 0, "Start at a one-based line number")
	catCmd.Flags().IntVar(&catFlags.end, "end", 0, "Stop at a one-based line number")
	catCmd.Flags().IntVar(&catFlags.head, "head", 0, "Show only the first N lines")
	catCmd.Flags().IntVar(&catFlags.tail, "tail", 0, "Show only the final N lines")
	catCmd.Flags().BoolVar(&catFlags.noHeader, "no-header", false, "Do not print file headers for multiple files")
	catCmd.Flags().BoolVar(&catFlags.force, "force", false, "Print binary content")
}
