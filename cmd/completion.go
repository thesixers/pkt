package cmd

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"

	"github.com/genesix/pkt/internal/config"
	"github.com/genesix/pkt/internal/db"
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
func completeConfigKeys(_ *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	if len(args) == 0 {
		return completionFilter([]string{"editor", "pm", "ai"}, toComplete), cobra.ShellCompDirectiveNoFileComp
	}
	if len(args) == 1 {
		switch args[0] {
		case "pm":
			return completionFilter(pm.ListAvailable(), toComplete), cobra.ShellCompDirectiveNoFileComp
		case "ai":
			return completionFilter(knownAIProviders(), toComplete), cobra.ShellCompDirectiveNoFileComp
		case "editor":
			// Fall back to PATH-based completion for editor binary names
			return nil, cobra.ShellCompDirectiveDefault
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

// completePMThenProject completes PM name for first arg, project name for second.
// DEPRECATED: This was used by pm set which is now removed. Keeping function signature to avoid breaking other files if imported.
func completePMThenProject(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	if len(args) == 0 {
		return completionFilter(pm.ListAvailable(), toComplete), cobra.ShellCompDirectiveNoFileComp
	}
	if len(args) == 1 {
		return completeProjectOrDot(cmd, args, toComplete)
	}
	return nil, cobra.ShellCompDirectiveNoFileComp
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
