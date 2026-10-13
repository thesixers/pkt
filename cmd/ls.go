package cmd

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/genesix/pkt/internal/utils"
	"github.com/spf13/cobra"
)

type listOptions struct {
	all       bool
	long      bool
	recursive bool
	dirs      bool
	files     bool
	sortBy    string
	reverse   bool
}

type listEntry struct {
	path  string
	label string
	info  os.FileInfo
}

var lsOptions listOptions

var lsCmd = &cobra.Command{
	Use:   "ls [path]",
	Short: "List files and directories",
	Long: `List files and directories from a normal path or a tracked project reference.

Examples:
  pkt ls
  pkt ls ./src
  pkt ls -la @api
  pkt ls -r --sort modified @api/src`,
	Args:              cobra.MaximumNArgs(1),
	ValidArgsFunction: completePktPath,
	RunE: func(cmd *cobra.Command, args []string) error {
		if lsOptions.dirs && lsOptions.files {
			return fmt.Errorf("--dirs and --files cannot be used together")
		}
		if lsOptions.sortBy != "name" && lsOptions.sortBy != "size" && lsOptions.sortBy != "modified" {
			return fmt.Errorf("invalid sort field %q; use name, size, or modified", lsOptions.sortBy)
		}
		input := "."
		if len(args) == 1 {
			input = args[0]
		}
		path, err := utils.ResolvePktPath(input)
		if err != nil {
			return err
		}
		entries, err := collectListEntries(path, lsOptions)
		if err != nil {
			return err
		}
		return writeList(cmd.OutOrStdout(), entries, lsOptions)
	},
}

func collectListEntries(path string, options listOptions) ([]listEntry, error) {
	info, err := os.Stat(path)
	if err != nil {
		return nil, fmt.Errorf("cannot list %s: %w", path, err)
	}
	if !info.IsDir() {
		return filterListEntries([]listEntry{{path: path, label: filepath.Base(path), info: info}}, options), nil
	}

	var entries []listEntry
	if options.recursive {
		err := filepath.WalkDir(path, func(current string, entry os.DirEntry, walkErr error) error {
			if walkErr != nil {
				return walkErr
			}
			if current == path {
				return nil
			}
			rel, err := filepath.Rel(path, current)
			if err != nil {
				return err
			}
			if !options.all && hasHiddenComponent(rel) {
				if entry.IsDir() {
					return filepath.SkipDir
				}
				return nil
			}
			entryInfo, err := entry.Info()
			if err != nil {
				return err
			}
			entries = append(entries, listEntry{path: current, label: rel, info: entryInfo})
			return nil
		})
		if err != nil {
			return nil, fmt.Errorf("list %s: %w", path, err)
		}
	} else {
		dirEntries, err := os.ReadDir(path)
		if err != nil {
			return nil, fmt.Errorf("list %s: %w", path, err)
		}
		for _, entry := range dirEntries {
			if !options.all && strings.HasPrefix(entry.Name(), ".") {
				continue
			}
			entryInfo, err := entry.Info()
			if err != nil {
				return nil, err
			}
			entries = append(entries, listEntry{path: filepath.Join(path, entry.Name()), label: entry.Name(), info: entryInfo})
		}
	}
	return filterListEntries(entries, options), nil
}

func filterListEntries(entries []listEntry, options listOptions) []listEntry {
	filtered := make([]listEntry, 0, len(entries))
	for _, entry := range entries {
		if options.dirs && !entry.info.IsDir() {
			continue
		}
		if options.files && entry.info.IsDir() {
			continue
		}
		filtered = append(filtered, entry)
	}
	return filtered
}

func writeList(writer io.Writer, entries []listEntry, options listOptions) error {
	sort.SliceStable(entries, func(i, j int) bool {
		var less bool
		switch options.sortBy {
		case "size":
			si, sj := entries[i].info.Size(), entries[j].info.Size()
			if si != sj {
				less = si < sj
			} else {
				// tie-break by case-insensitive name
				less = strings.ToLower(entries[i].label) < strings.ToLower(entries[j].label)
			}
		case "modified":
			ti, tj := entries[i].info.ModTime(), entries[j].info.ModTime()
			if !ti.Equal(tj) {
				less = ti.Before(tj)
			} else {
				less = strings.ToLower(entries[i].label) < strings.ToLower(entries[j].label)
			}
		default:
			li, lj := strings.ToLower(entries[i].label), strings.ToLower(entries[j].label)
			if li != lj {
				less = li < lj
			} else {
				// case-sensitive tie-break (upper-case before lower-case)
				less = entries[i].label < entries[j].label
			}
		}
		if options.reverse {
			return !less
		}
		return less
	})

	for _, entry := range entries {
		label := entry.label
		if entry.info.IsDir() {
			label += "/"
		}
		if options.long {
			modified := entry.info.ModTime().Format(time.DateTime)
			if _, err := fmt.Fprintf(writer, "%s %10d %s %s\n", entry.info.Mode().String(), entry.info.Size(), modified, label); err != nil {
				return err
			}
		} else if _, err := fmt.Fprintln(writer, label); err != nil {
			return err
		}
	}
	return nil
}

func hasHiddenComponent(path string) bool {
	for _, part := range strings.Split(filepath.Clean(path), string(filepath.Separator)) {
		if strings.HasPrefix(part, ".") && part != "." && part != ".." {
			return true
		}
	}
	return false
}

func init() {
	rootCmd.AddCommand(lsCmd)
	lsCmd.Flags().BoolVarP(&lsOptions.all, "all", "a", false, "Include hidden files and directories")
	lsCmd.Flags().BoolVarP(&lsOptions.long, "long", "l", false, "Show permissions, size, and modified time")
	lsCmd.Flags().BoolVarP(&lsOptions.recursive, "recursive", "r", false, "Walk subdirectories recursively")
	lsCmd.Flags().BoolVarP(&lsOptions.dirs, "dirs", "d", false, "Show directories only")
	lsCmd.Flags().BoolVarP(&lsOptions.files, "files", "f", false, "Show files only")
	lsCmd.Flags().StringVar(&lsOptions.sortBy, "sort", "name", "Sort by name, size, or modified")
	lsCmd.Flags().BoolVar(&lsOptions.reverse, "reverse", false, "Reverse the selected sort order")
}
