package cmd

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"

	"github.com/genesix/pkt/internal/config"
	"github.com/genesix/pkt/internal/db"
	"github.com/genesix/pkt/internal/lang"
	"github.com/genesix/pkt/internal/pm"
	"github.com/spf13/cobra"
)

// completeProjects returns all tracked project names for shell completion.
func completeProjects(_ *cobra.Command, _ []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	projects, err := db.ListAllProjects()
	if err != nil {
		return nil, cobra.ShellCompDirectiveNoFileComp
	}

	var names []string
	for _, p := range projects {
		if strings.HasPrefix(strings.ToLower(p.Name), strings.ToLower(toComplete)) {
			names = append(names, p.Name)
		}
	}
	return names, cobra.ShellCompDirectiveNoFileComp
}

// completeProjectOrDot completes project names plus "." for current directory.
func completeProjectOrDot(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	names, directive := completeProjects(cmd, args, toComplete)
	if strings.HasPrefix(".", toComplete) {
		names = append([]string{"."}, names...)
	}
	return names, directive
}

// completeDepsForCurrentProject returns installed dependency names for the cwd project.
func completeDepsForCurrentProject(_ *cobra.Command, _ []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	cwd, err := os.Getwd()
	if err != nil {
		return nil, cobra.ShellCompDirectiveNoFileComp
	}

	project, err := db.GetProjectByPath(cwd)
	if err != nil {
		return nil, cobra.ShellCompDirectiveNoFileComp
	}

	deps, err := db.GetDependencies(project.ID)
	if err != nil {
		return nil, cobra.ShellCompDirectiveNoFileComp
	}

	var names []string
	for _, d := range deps {
		if strings.HasPrefix(strings.ToLower(d.Name), strings.ToLower(toComplete)) {
			names = append(names, d.Name)
		}
	}
	return names, cobra.ShellCompDirectiveNoFileComp
}

// completeScripts returns runnable script/command names for the cwd project.
func completeScripts(_ *cobra.Command, _ []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	cwd, err := os.Getwd()
	if err != nil {
		return nil, cobra.ShellCompDirectiveNoFileComp
	}

	project, err := db.GetProjectByPath(cwd)
	if err != nil {
		return nil, cobra.ShellCompDirectiveNoFileComp
	}

	var scripts []string
	switch project.Language {
	case "javascript":
		scripts = parsePackageJSONScripts(project.Path)
		if len(scripts) == 0 {
			scripts = []string{"dev", "build", "start", "test", "lint", "preview"}
		}
	case "python":
		scripts = []string{"test", "lint", "format", "run"}
	case "go":
		scripts = []string{"run", "build", "test", "vet", "fmt", "mod"}
	case "rust":
		scripts = []string{"run", "build", "test", "check", "clippy", "fmt"}
	}

	return completionFilter(scripts, toComplete), cobra.ShellCompDirectiveNoFileComp
}

// parsePackageJSONScripts reads script names from a package.json file.
func parsePackageJSONScripts(projectPath string) []string {
	data, err := os.ReadFile(filepath.Join(projectPath, "package.json"))
	if err != nil {
		return nil
	}
	var pkg struct {
		Scripts map[string]string `json:"scripts"`
	}
	if err := json.Unmarshal(data, &pkg); err != nil {
		return nil
	}
	names := make([]string, 0, len(pkg.Scripts))
	for name := range pkg.Scripts {
		names = append(names, name)
	}
	return names
}

// completeConfigKeys completes config key names and their allowed values.
func completeConfigKeys(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	if len(args) == 0 {
		return completionFilter([]string{"editor", "pm", "root", "lang", "language", "rebuild"}, toComplete), cobra.ShellCompDirectiveNoFileComp
	}
	if len(args) == 1 {
		switch args[0] {
		case "pm":
			return completionFilter(pm.ListAvailable(), toComplete), cobra.ShellCompDirectiveNoFileComp
		case "lang", "language":
			allLangs := append(lang.ListShort(), lang.List()...)
			return completionFilter(allLangs, toComplete), cobra.ShellCompDirectiveNoFileComp
		case "rebuild":
			return completeProjects(cmd, args, toComplete)
		case "editor":
			return nil, cobra.ShellCompDirectiveDefault
		case "root":
			return nil, cobra.ShellCompDirectiveDefault
		}
	}
	if len(args) == 2 {
		switch args[0] {
		case "lang", "language":
			return completeProjects(cmd, args, toComplete)
		}
	}
	return nil, cobra.ShellCompDirectiveNoFileComp
}

// completeAIProviderArg completes the first argument as an AI provider name.
func completeAIProviderArg(_ *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	if len(args) == 0 {
		return completionFilter(knownAIProviders(), toComplete), cobra.ShellCompDirectiveNoFileComp
	}
	return nil, cobra.ShellCompDirectiveNoFileComp
}

// completePktPath completes @project/path arguments for pkt cp and pkt mv.
//
// NOTE ON BASH WORD-SPLITTING: bash treats '@' as a word boundary, so when
// the user types "pkt cp @<TAB>" the shell passes toComplete="" (empty string),
// not "@". We handle all three cases:
//  1. toComplete is "" or doesn't start with "@" → offer all @project/ names
//     (the @ prefix itself is included in each suggestion) AND allow file comp.
//  2. toComplete starts with "@" but has no "/" → filter project names by prefix.
//  3. toComplete is "@project/sub…" → enumerate files inside the project root.
func completePktPath(_ *cobra.Command, _ []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	// Case 1: empty token or plain path — bash gave us "" after the "@" word break.
	// Emit all @project/ names so they appear alongside normal file completions.
	if toComplete == "" || (!strings.HasPrefix(toComplete, "@") && !strings.Contains(toComplete, "/")) {
		projects, err := db.ListAllProjects()
		if err != nil {
			return nil, cobra.ShellCompDirectiveDefault
		}
		var completions []string
		for _, p := range projects {
			if strings.HasPrefix(strings.ToLower(p.Name), strings.ToLower(toComplete)) {
				completions = append(completions, "@"+p.Name+"/")
			}
		}
		// Return @project/ suggestions alongside normal filesystem completion.
		return completions, cobra.ShellCompDirectiveNoSpace
	}

	// Strip the leading "@".
	withoutAt := toComplete[1:]

	// Case 2: "@partial" — still completing the project name portion.
	sepIdx := strings.Index(withoutAt, "/")
	if sepIdx == -1 {
		projects, err := db.ListAllProjects()
		if err != nil {
			return nil, cobra.ShellCompDirectiveNoFileComp
		}
		var completions []string
		for _, p := range projects {
			if strings.HasPrefix(strings.ToLower(p.Name), strings.ToLower(withoutAt)) {
				completions = append(completions, "@"+p.Name+"/")
			}
		}
		return completions, cobra.ShellCompDirectiveNoSpace | cobra.ShellCompDirectiveNoFileComp
	}

	// Case 3: "@project/sub…" — enumerate files inside the resolved project root.
	projectName := withoutAt[:sepIdx]
	subPath := withoutAt[sepIdx+1:]

	projects, err := db.ListAllProjects()
	if err != nil {
		return nil, cobra.ShellCompDirectiveNoFileComp
	}

	var projectRoot string
	for _, p := range projects {
		if strings.EqualFold(p.Name, projectName) {
			projectRoot = p.Path
			break
		}
	}
	if projectRoot == "" {
		return nil, cobra.ShellCompDirectiveNoFileComp
	}

	// Determine which directory to list and which basename prefix to match.
	searchDir := filepath.Join(projectRoot, filepath.Dir(subPath))
	basePart := filepath.Base(subPath)
	if subPath == "" || strings.HasSuffix(subPath, "/") {
		basePart = ""
		searchDir = filepath.Join(projectRoot, subPath)
	}

	entries, err := os.ReadDir(searchDir)
	if err != nil {
		return nil, cobra.ShellCompDirectiveNoFileComp
	}

	var completions []string
	for _, e := range entries {
		if basePart != "" && !strings.HasPrefix(strings.ToLower(e.Name()), strings.ToLower(basePart)) {
			continue
		}
		relToRoot, _ := filepath.Rel(projectRoot, filepath.Join(searchDir, e.Name()))
		token := "@" + projectName + "/" + relToRoot
		if e.IsDir() {
			token += "/"
		}
		completions = append(completions, token)
	}

	return completions, cobra.ShellCompDirectiveNoSpace | cobra.ShellCompDirectiveNoFileComp
}

// completeLangFlag completes language short-code values for --lang flags.
func completeLangFlag(_ *cobra.Command, _ []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	return completionFilter([]string{"js", "py", "go", "rs"}, toComplete), cobra.ShellCompDirectiveNoFileComp
}

// knownAIProviders returns well-known provider names plus any registered in config.
func knownAIProviders() []string {
	known := []string{"openai", "gemini", "groq", "ollama", "local"}
	if cfg, err := config.Load(); err == nil {
		for name := range cfg.AIProviders {
			if !sliceContains(known, name) {
				known = append(known, name)
			}
		}
	}
	return known
}

// completionFilter filters a list to items that start with the given prefix.
func completionFilter(items []string, prefix string) []string {
	if prefix == "" {
		return items
	}
	var result []string
	for _, item := range items {
		if strings.HasPrefix(item, prefix) {
			result = append(result, item)
		}
	}
	return result
}

// sliceContains reports whether s is in the slice.
func sliceContains(slice []string, s string) bool {
	for _, v := range slice {
		if v == s {
			return true
		}
	}
	return false
}
