package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/dustin/go-humanize"
	"github.com/genesix/pkt/internal/utils"
	"github.com/spf13/cobra"
)

var infoCmd = &cobra.Command{
	Use:   "info <project | id>",
	Short: "Display project information from its README or Release Notes",
	Long:  "Parses the first few describing lines from a project's documentation file.",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		input := args[0]

		project, err := utils.ResolveProject(input)
		if err != nil {
			return err
		}

		infoText := extractProjectInfo(project.Path)
		if infoText == "" {
			fmt.Println("No info for this project.")
			return nil
		}

		sizeBytes, _ := utils.GetDirSize(project.Path)
		sizeStr := humanize.Bytes(uint64(sizeBytes))

		fmt.Printf("\nProject Info: %s\n", project.Name)
		fmt.Printf("ID: %s\n", project.ID)
		fmt.Printf("Size: %s\n", sizeStr)
		fmt.Printf("Path: %s\n", utils.ShortPath(project.Path))
		fmt.Println(strings.Repeat("-", 40))
		fmt.Println(infoText)
		fmt.Println(strings.Repeat("-", 40))

		return nil
	},
}

func extractProjectInfo(projectPath string) string {
	entries, err := os.ReadDir(projectPath)
	if err != nil {
		return ""
	}

	var readmePath, releasePath, firstMdPath string
	for _, e := range entries {
		if e.IsDir() {
			continue
		}
		name := strings.ToLower(e.Name())
		if name == "readme" || name == "readme.md" {
			readmePath = filepath.Join(projectPath, e.Name())
		} else if strings.Contains(name, "release note") || strings.Contains(name, "release_note") || strings.Contains(name, "release-note") {
			if releasePath == "" {
				releasePath = filepath.Join(projectPath, e.Name())
			}
		} else if strings.HasSuffix(name, ".md") {
			if firstMdPath == "" {
				firstMdPath = filepath.Join(projectPath, e.Name())
			}
		}
	}

	targetFile := readmePath
	if targetFile == "" {
		targetFile = releasePath
	}
	if targetFile == "" {
		targetFile = firstMdPath
	}

	if targetFile == "" {
		return ""
	}

	content, err := os.ReadFile(targetFile)
	if err != nil {
		return ""
	}

	lines := strings.Split(string(content), "\n")
	var capturedLines []string

	textParaCount := 0
	inCodeBlock := false
	var currentParaLength int

	for _, line := range lines {
		trimmedLine := strings.TrimSpace(line)

		// Always capture the original line to preserve formatting
		capturedLines = append(capturedLines, line)

		if strings.HasPrefix(trimmedLine, "```") {
			inCodeBlock = !inCodeBlock
			continue
		}

		if inCodeBlock {
			continue
		}

		isNormalText := trimmedLine != "" &&
			!strings.HasPrefix(trimmedLine, "#") &&
			!strings.HasPrefix(trimmedLine, "=") &&
			!strings.HasPrefix(trimmedLine, "<") &&
			trimmedLine != "---" &&
			trimmedLine != "***" &&
			trimmedLine != "___" &&
			!strings.HasPrefix(trimmedLine, "![") &&
			!strings.HasPrefix(trimmedLine, "[![")

		if isNormalText {
			currentParaLength += len(trimmedLine)
		} else if trimmedLine == "" {
			if currentParaLength > 0 {
				textParaCount++
				currentParaLength = 0
			}

			// Stop if we've accumulated 2 normal text paragraphs,
			// or 1 paragraph that is fairly long
			if textParaCount >= 2 {
				break
			}
		}
	}

	return prettifyMarkdown(strings.TrimSpace(strings.Join(capturedLines, "\n")))
}

func prettifyMarkdown(text string) string {
	lines := strings.Split(text, "\n")
	for i, line := range lines {
		if strings.HasPrefix(line, "# ") {
			lines[i] = "\033[1;36m" + strings.TrimPrefix(line, "# ") + "\033[0m"
		} else if strings.HasPrefix(line, "## ") {
			lines[i] = "\033[1;34m" + strings.TrimPrefix(line, "## ") + "\033[0m"
		} else if strings.HasPrefix(line, "### ") {
			lines[i] = "\033[1m" + strings.TrimPrefix(line, "### ") + "\033[0m"
		} else if strings.HasPrefix(line, "---") || strings.HasPrefix(line, "***") {
			lines[i] = "\033[90m" + line + "\033[0m"
		} else if strings.HasPrefix(line, "> ") {
			lines[i] = "\033[3;90m" + line + "\033[0m"
		}
	}
	text = strings.Join(lines, "\n")

	reBold := regexp.MustCompile(`\*\*(.*?)\*\*`)
	text = reBold.ReplaceAllString(text, "\033[1m$1\033[0m")

	reItalicStar := regexp.MustCompile(`\*([^*]+)\*`)
	text = reItalicStar.ReplaceAllString(text, "\033[3m$1\033[0m")

	reItalicUnderscore := regexp.MustCompile(`(^|\s)_([^_]+)_(\s|$)`)
	text = reItalicUnderscore.ReplaceAllString(text, "${1}\033[3m${2}\033[0m${3}")

	reCode := regexp.MustCompile("`([^`]+)`")
	text = reCode.ReplaceAllString(text, "\033[33m$1\033[0m")

	return text
}

func init() {
	rootCmd.AddCommand(infoCmd)
}
