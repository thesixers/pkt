package ai

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

// Skill represents a learned agent capability or instruction set.
type Skill struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	Content     string `json:"content"`
	Path        string `json:"path"`
	IsGlobal    bool   `json:"is_global"`
}

// GetSkillDirectories returns the priority-ordered locations to discover skills.
func GetSkillDirectories() []string {
	var dirs []string
	cwd, err := os.Getwd()
	if err == nil {
		// Project local skills (hidden directories preferred)
		dirs = append(dirs, filepath.Join(cwd, ".skills"))
		dirs = append(dirs, filepath.Join(cwd, ".pkt", "skills"))
	}

	home, err := os.UserHomeDir()
	if err == nil {
		// Global pkt skills
		dirs = append(dirs, filepath.Join(home, ".pkt", "skills"))
		// Shared global skills from Gemini / Antigravity config if present (read-only sharing)
		geminiSkills := filepath.Join(home, ".gemini", "config", "skills")
		if _, err := os.Stat(geminiSkills); err == nil {
			dirs = append(dirs, geminiSkills)
		}
	}

	return dirs
}

// LoadAllSkills scans all skill directories and returns a map of skill name -> Skill.
func LoadAllSkills() map[string]Skill {
	skills := make(map[string]Skill)
	dirs := GetSkillDirectories()

	home, _ := os.UserHomeDir()
	globalPkt := filepath.Join(home, ".pkt", "skills")
	globalGemini := filepath.Join(home, ".gemini", "config", "skills")

	// Reverse iterate so project-local skills override global ones
	for i := len(dirs) - 1; i >= 0; i-- {
		dir := dirs[i]
		entries, err := os.ReadDir(dir)
		if err != nil {
			continue
		}

		isGlobal := (dir == globalPkt || dir == globalGemini)

		for _, entry := range entries {
			entryPath := filepath.Join(dir, entry.Name())

			if entry.IsDir() {
				// Directory with SKILL.md or README.md
				skillFile := filepath.Join(entryPath, "SKILL.md")
				if _, err := os.Stat(skillFile); os.IsNotExist(err) {
					skillFile = filepath.Join(entryPath, "README.md")
				}
				if info, err := os.Stat(skillFile); err == nil && !info.IsDir() {
					skill, err := parseSkillFile(skillFile, entry.Name(), isGlobal)
					if err == nil {
						skills[strings.ToLower(skill.Name)] = skill
					}
				}
			} else if strings.HasSuffix(entry.Name(), ".md") {
				// Standalone <name>.md file
				baseName := strings.TrimSuffix(entry.Name(), ".md")
				skill, err := parseSkillFile(entryPath, baseName, isGlobal)
				if err == nil {
					skills[strings.ToLower(skill.Name)] = skill
				}
			}
		}
	}

	return skills
}

func parseSkillFile(filePath, fallbackName string, isGlobal bool) (Skill, error) {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return Skill{}, err
	}

	raw := string(data)
	name := fallbackName
	description := ""
	content := raw

	// Parse YAML frontmatter if present
	if strings.HasPrefix(raw, "---") {
		parts := strings.SplitN(raw[3:], "---", 2)
		if len(parts) == 2 {
			frontmatter := parts[0]
			content = strings.TrimSpace(parts[1])

			fLines := strings.Split(frontmatter, "\n")
			for i := 0; i < len(fLines); i++ {
				fLine := strings.TrimSpace(fLines[i])
				if strings.HasPrefix(fLine, "name:") {
					val := strings.TrimSpace(strings.TrimPrefix(fLine, "name:"))
					val = strings.Trim(val, `"'`)
					if val != "" {
						name = val
					}
				} else if strings.HasPrefix(fLine, "description:") {
					val := strings.TrimSpace(strings.TrimPrefix(fLine, "description:"))
					val = strings.Trim(val, `"'`)
					if val == ">" || val == "|" || val == "" {
						var descParts []string
						for j := i + 1; j < len(fLines); j++ {
							nextLine := fLines[j]
							if len(nextLine) > 0 && (nextLine[0] == ' ' || nextLine[0] == '\t') {
								t := strings.TrimSpace(nextLine)
								if t != "" {
									descParts = append(descParts, t)
								}
								i = j
							} else {
								break
							}
						}
						description = strings.Join(descParts, " ")
					} else {
						description = val
					}
				}
			}
		}
	}

	if description == "" || description == ">" || description == "|" {
		// Use first non-heading line as description
		for _, line := range strings.Split(content, "\n") {
			trimmed := strings.TrimSpace(line)
			if trimmed != "" && !strings.HasPrefix(trimmed, "#") {
				description = trimmed
				if len(description) > 100 {
					description = description[:97] + "..."
				}
				break
			}
		}
	}
	if description == "" || description == ">" || description == "|" {
		description = "Custom skill guidelines and procedures."
	}

	return Skill{
		Name:        name,
		Description: description,
		Content:     content,
		Path:        filePath,
		IsGlobal:    isGlobal,
	}, nil
}

// SaveSkill saves a new skill into ~/.pkt/skills (for global) or .skills/ (for project).
// Note: Global skills are ALWAYS saved to ~/.pkt/skills (never ~/.gemini/config/skills).
func SaveSkill(name, description, instructions string, global bool) (string, error) {
	cleanName := strings.ToLower(strings.TrimSpace(name))
	cleanName = regexp.MustCompile(`[^a-zA-Z0-9_\-]`).ReplaceAllString(cleanName, "-")
	cleanName = strings.Trim(cleanName, "-")
	if cleanName == "" {
		return "", fmt.Errorf("invalid skill name")
	}

	var baseDir string
	if global {
		home, err := os.UserHomeDir()
		if err != nil {
			return "", err
		}
		// ALWAYS save global skills to ~/.pkt/skills
		baseDir = filepath.Join(home, ".pkt", "skills")
	} else {
		cwd, err := os.Getwd()
		if err != nil {
			return "", err
		}
		// Save project skills in hidden .skills/ folder
		baseDir = filepath.Join(cwd, ".skills")
	}

	skillDir := filepath.Join(baseDir, cleanName)
	if err := os.MkdirAll(skillDir, 0755); err != nil {
		return "", err
	}

	skillFile := filepath.Join(skillDir, "SKILL.md")
	doc := fmt.Sprintf("---\nname: %s\ndescription: %s\n---\n\n# %s\n\n%s\n",
		cleanName, strings.TrimSpace(description), cleanName, strings.TrimSpace(instructions))

	if err := os.WriteFile(skillFile, []byte(doc), 0644); err != nil {
		return "", err
	}

	return skillFile, nil
}

// ExtractSkillMentions detects #skill or $skill in user input, matches against loaded skills,
// and returns the activated skills and formatted prompt context.
func ExtractSkillMentions(input string, availableSkills map[string]Skill) ([]Skill, string) {
	if len(availableSkills) == 0 {
		return nil, ""
	}

	// Match #skill-name or $skill-name
	re := regexp.MustCompile(`[#$]([a-zA-Z0-9_\-]+)`)
	matches := re.FindAllStringSubmatch(input, -1)
	if len(matches) == 0 {
		return nil, ""
	}

	var activeSkills []Skill
	seen := make(map[string]bool)
	var promptBuilder strings.Builder

	for _, match := range matches {
		skillKey := strings.ToLower(match[1])
		if seen[skillKey] {
			continue
		}

		if skill, exists := availableSkills[skillKey]; exists {
			seen[skillKey] = true
			activeSkills = append(activeSkills, skill)

			promptBuilder.WriteString("\n\n")
			fmt.Fprintf(&promptBuilder, "[Active Learned Skill: %s]\nDescription: %s\nGuidelines & Instructions:\n%s\n",
				skill.Name, skill.Description, skill.Content)
		}
	}

	return activeSkills, promptBuilder.String()
}
