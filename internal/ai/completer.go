package ai

import (
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/genesix/pkt/internal/config"
)

var defaultSlashCommands = []string{
	"/help",
	"/provider",
	"/providers",
	"/model",
	"/models",
	"/clear",
	"/reset",
	"/approve",
	"/skills",
	"/skill",
	"/learn",
	"/info",
	"/tools",
	"/context",
	"/exit",
	"/quit",
}

// ChatCompleter provides dynamic auto-completion for pkt chat commands, AI providers/models, @file attachments, and #skills.
type ChatCompleter struct {
	getProvider func() string
}

// NewChatCompleter creates an AutoCompleter instance aware of current provider and models.
func NewChatCompleter(getProvider func() string) *ChatCompleter {
	return &ChatCompleter{
		getProvider: getProvider,
	}
}

// Do implements readline.AutoCompleter.
func (c *ChatCompleter) Do(line []rune, pos int) (newLine [][]rune, length int) {
	if pos < 0 || pos > len(line) {
		return nil, 0
	}

	prefix := string(line[:pos])

	// Case 1: Slash command completion
	if strings.HasPrefix(prefix, "/") {
		return c.completeSlashCommand(prefix)
	}

	// Case 2: @file / @dir context completion
	if atIdx := strings.LastIndex(prefix, "@"); atIdx != -1 {
		afterAt := prefix[atIdx+1:]
		if !strings.ContainsAny(afterAt, " \t\r\n") {
			return c.completeFilePath(afterAt)
		}
	}

	// Case 3: #skill or $skill completion
	if idx := strings.LastIndexAny(prefix, "#$"); idx != -1 {
		afterSign := prefix[idx+1:]
		if !strings.ContainsAny(afterSign, " \t\r\n") {
			return c.completeSkill(afterSign)
		}
	}

	return nil, 0
}

func (c *ChatCompleter) completeSlashCommand(prefix string) ([][]rune, int) {
	// If still typing the command itself (no space yet)
	if !strings.Contains(prefix, " ") {
		var matches [][]rune
		for _, cmd := range defaultSlashCommands {
			if strings.HasPrefix(cmd, prefix) {
				suffix := cmd[len(prefix):] + " "
				matches = append(matches, []rune(suffix))
			}
		}
		return matches, len([]rune(prefix))
	}

	// Subcommand argument completion
	parts := strings.SplitN(prefix, " ", 2)
	cmd := parts[0]
	argPrefix := ""
	if len(parts) > 1 {
		argPrefix = parts[1]
	}

	switch cmd {
	case "/provider":
		cfg, _ := config.Load()
		providerSet := make(map[string]bool)
		var providers []string

		for p := range builtinDefaults {
			providerSet[p] = true
			providers = append(providers, p)
		}
		if cfg != nil {
			for _, p := range cfg.ProviderOrder {
				if !providerSet[p] {
					providerSet[p] = true
					providers = append(providers, p)
				}
			}
			for p := range cfg.AIProviders {
				if !providerSet[p] {
					providerSet[p] = true
					providers = append(providers, p)
				}
			}
		}

		var matches [][]rune
		for _, p := range providers {
			if strings.HasPrefix(p, argPrefix) {
				suffix := p[len(argPrefix):] + " "
				matches = append(matches, []rune(suffix))
			}
		}
		return matches, len([]rune(argPrefix))

	case "/model":
		activeProv := ""
		if c.getProvider != nil {
			activeProv = c.getProvider()
		}
		cfg, _ := config.Load()
		modelSet := make(map[string]bool)
		var models []string

		if cfg != nil && activeProv != "" {
			pc := cfg.AIProviders[activeProv]
			if pc.DefaultModel != "" && !modelSet[pc.DefaultModel] {
				modelSet[pc.DefaultModel] = true
				models = append(models, pc.DefaultModel)
			}
			for _, m := range pc.Models {
				if !modelSet[m] {
					modelSet[m] = true
					models = append(models, m)
				}
			}
		}
		if d, ok := builtinDefaults[activeProv]; ok && !modelSet[d[1]] {
			models = append(models, d[1])
		}

		var matches [][]rune
		for _, m := range models {
			if strings.HasPrefix(m, argPrefix) {
				suffix := m[len(argPrefix):] + " "
				matches = append(matches, []rune(suffix))
			}
		}
		return matches, len([]rune(argPrefix))

	case "/skill":
		cleanPrefix := strings.TrimPrefix(strings.TrimPrefix(argPrefix, "#"), "$")
		skills := LoadAllSkills()
		var matches [][]rune
		for name := range skills {
			if strings.HasPrefix(name, cleanPrefix) {
				suffix := name[len(cleanPrefix):] + " "
				matches = append(matches, []rune(suffix))
			}
		}
		return matches, len([]rune(argPrefix))
	}

	return nil, 0
}

func (c *ChatCompleter) completeFilePath(pathPrefix string) ([][]rune, int) {
	dir := "."
	base := pathPrefix
	if lastSlash := strings.LastIndex(pathPrefix, "/"); lastSlash != -1 {
		dir = pathPrefix[:lastSlash]
		if dir == "" {
			dir = "/"
		}
		base = pathPrefix[lastSlash+1:]
	}

	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, 0
	}

	var matches [][]rune
	for _, entry := range entries {
		name := entry.Name()
		if strings.HasPrefix(name, ".") && !strings.HasPrefix(base, ".") {
			continue
		}
		if entry.IsDir() && (name == "node_modules" || name == ".git" || name == "vendor") {
			continue
		}

		if strings.HasPrefix(name, base) {
			var fullMatch string
			switch dir {
			case ".":
				fullMatch = name
			case "/":
				fullMatch = "/" + name
			default:
				fullMatch = filepath.ToSlash(filepath.Join(dir, name))
			}

			suffix := fullMatch[len(pathPrefix):]
			if entry.IsDir() {
				suffix += "/"
			} else {
				suffix += " "
			}
			matches = append(matches, []rune(suffix))
		}
	}

	return matches, len([]rune(pathPrefix))
}

func (c *ChatCompleter) completeSkill(skillPrefix string) ([][]rune, int) {
	lowerPrefix := strings.ToLower(skillPrefix)
	skills := LoadAllSkills()
	var names []string
	for name := range skills {
		if strings.HasPrefix(strings.ToLower(name), lowerPrefix) {
			names = append(names, name)
		}
	}
	sort.Strings(names)

	var matches [][]rune
	for _, name := range names {
		suffix := name[len(skillPrefix):] + " "
		matches = append(matches, []rune(suffix))
	}
	return matches, len([]rune(skillPrefix))
}
