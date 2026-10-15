package ai

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/AlecAivazis/survey/v2"
	"github.com/chzyer/readline"
	"github.com/dustin/go-humanize"
	"github.com/genesix/pkt/internal/config"
	"github.com/genesix/pkt/internal/db"
	"golang.org/x/term"
)

var definedTools = []Tool{
	{
		Type: "function",
		Function: ToolFunction{
			Name:        "get_project_info",
			Description: "Get framework, language, package manager, and git branch for the active project.",
			Parameters: ToolParameters{
				Type:       "object",
				Properties: map[string]interface{}{},
			},
		},
	},
	{
		Type: "function",
		Function: ToolFunction{
			Name:        "list_dir",
			Description: "List files and folders in a directory.",
			Parameters: ToolParameters{
				Type: "object",
				Properties: map[string]interface{}{
					"path": map[string]interface{}{"type": "string", "description": "Absolute or relative path (defaults to .)"},
				},
				Required: []string{"path"},
			},
		},
	},
	{
		Type: "function",
		Function: ToolFunction{
			Name:        "read_file",
			Description: "Read text contents of a file.",
			Parameters: ToolParameters{
				Type: "object",
				Properties: map[string]interface{}{
					"path": map[string]interface{}{"type": "string", "description": "Path to file"},
				},
				Required: []string{"path"},
			},
		},
	},
	{
		Type: "function",
		Function: ToolFunction{
			Name:        "write_file",
			Description: "Write/Create/Edit a file with raw content.",
			Parameters: ToolParameters{
				Type: "object",
				Properties: map[string]interface{}{
					"path":    map[string]interface{}{"type": "string", "description": "Path to file"},
					"content": map[string]interface{}{"type": "string", "description": "Raw file contents"},
				},
				Required: []string{"path", "content"},
			},
		},
	},
	{
		Type: "function",
		Function: ToolFunction{
			Name:        "delete_file",
			Description: "Delete a specific file from the project.",
			Parameters: ToolParameters{
				Type: "object",
				Properties: map[string]interface{}{
					"path": map[string]interface{}{"type": "string", "description": "Path to file to delete"},
				},
				Required: []string{"path"},
			},
		},
	},
	{
		Type: "function",
		Function: ToolFunction{
			Name:        "make_dir",
			Description: "Create a new directory recursively.",
			Parameters: ToolParameters{
				Type: "object",
				Properties: map[string]interface{}{
					"path": map[string]interface{}{"type": "string", "description": "Directory path to create"},
				},
				Required: []string{"path"},
			},
		},
	},
	{
		Type: "function",
		Function: ToolFunction{
			Name:        "run_command",
			Description: "Execute a bash shell command inside the project environment.",
			Parameters: ToolParameters{
				Type: "object",
				Properties: map[string]interface{}{
					"command": map[string]interface{}{"type": "string", "description": "Shell command to execute"},
				},
				Required: []string{"command"},
			},
		},
	},
	{
		Type: "function",
		Function: ToolFunction{
			Name:        "learn_skill",
			Description: "Save a new reusable skill, rule, or procedure into ~/.pkt/skills (global) or .skills (project) so the agent can reuse it in future sessions.",
			Parameters: ToolParameters{
				Type: "object",
				Properties: map[string]interface{}{
					"name":         map[string]interface{}{"type": "string", "description": "Skill name in kebab-case (e.g. react-tailwind, unit-testing)"},
					"description":  map[string]interface{}{"type": "string", "description": "Brief description of the skill and when to use it"},
					"instructions": map[string]interface{}{"type": "string", "description": "Detailed guidelines, rules, patterns, or procedures for the skill"},
					"global":       map[string]interface{}{"type": "boolean", "description": "True to save globally (~/.pkt/skills), false to save in project (.skills)"},
				},
				Required: []string{"name", "description", "instructions"},
			},
		},
	},
}

// StartChatSession enters the interactive REPL.
func StartChatSession(provider string, model string, projectContext string) error {
	cwd, _ := os.Getwd()
	project, projectErr := db.GetProjectByPath(cwd)

	cfg, _ := config.Load()

	// Resolve active provider
	currentProvider := provider
	if currentProvider == "" && cfg != nil {
		currentProvider = cfg.AIProvider
		if currentProvider == "" && len(cfg.ProviderOrder) > 0 {
			currentProvider = cfg.ProviderOrder[0]
		}
	}
	if currentProvider == "" {
		currentProvider = "groq"
	}

	// Resolve active model
	currentModel := model
	if currentModel == "" && cfg != nil {
		if pc, ok := cfg.AIProviders[currentProvider]; ok {
			currentModel = pc.DefaultModel
			if currentModel == "" && len(pc.Models) > 0 {
				currentModel = pc.Models[0]
			}
		}
	}
	if currentModel == "" {
		if d, ok := builtinDefaults[currentProvider]; ok {
			currentModel = d[1]
		} else {
			currentModel = "default"
		}
	}

	sysPrompt := "You are a powerful Autonomous Coding Agent built inherently into the pkt CLI. You have direct access to the user's terminal and file system. IMPORTANT: If the user asks a question about the project, tools, or environment (like 'list the commands' or 'what is this codebase'), you MUST autonomously use your tools (like run_command, list_dir, or read_file) to find the answer and then report back. However, if the user simply says a generic greeting like 'hello', 'hi', or makes conversational small-talk, reply normally WITHOUT invoking any tools.\n\nCRITICAL CONSTRAINTS:\n1. If the user asks a hypothetical question like 'can you create a file?', NEVER invoke the tool to demonstrate. Just answer 'Yes, I can'.\n2. NEVER invoke state-changing tools ('write_file', 'delete_file', 'make_dir', or destructive 'run_command') without explicitly being told to do so. Always wait for the user to instruct before executing.\n3. SKILLS & LEARNING: You have the ability to learn, remember, and persist new skills, conventions, workflows, or rules using the 'learn_skill' tool. When a user asks you to learn a skill or save reusable instructions, use 'learn_skill'. When a prompt contains an activated skill (#skill-name), you must strictly follow its guidelines."
	if projectErr == nil {
		sysPrompt += fmt.Sprintf("\nEnvironment: You are inside a %s project managed natively by %s located entirely at %s.", project.Language, project.PackageManager, project.Path)
	}
	if projectContext != "" {
		sysPrompt += fmt.Sprintf("\n\nProject Context:\n%s", projectContext)
	}

	messages := []Message{
		{Role: "system", Content: sysPrompt},
	}

	printBanner(currentProvider, currentModel, project)

	completer := NewChatCompleter(func() string {
		return currentProvider
	})

	rl, err := readline.NewEx(&readline.Config{
		Prompt:          formatPrompt(currentProvider, currentModel),
		HistoryFile:     chatHistoryPath(),
		AutoComplete:    completer,
		InterruptPrompt: "^C",
		EOFPrompt:       "/exit",
	})
	if err != nil {
		fmt.Printf("Error initializing readline: %s\n", err)
		return err
	}
	defer func() { _ = rl.Close() }()

	autoApproveCommands := false

	for {
		line, err := rl.Readline()
		if err != nil { // handles EOF (Ctrl+D) and Interrupt (Ctrl+C)
			fmt.Println("\nGoodbye!")
			break
		}

		userInput := strings.TrimSpace(line)
		if userInput == "" {
			continue
		}

		if strings.HasPrefix(userInput, "/") {
			parts := strings.Fields(userInput)
			cmdName := parts[0]
			switch cmdName {
			case "/exit", "/quit":
				fmt.Println("Goodbye!")
				return nil

			case "/help":
				printHelp()
				continue

			case "/clear":
				fmt.Print("\033[H\033[2J")
				printBanner(currentProvider, currentModel, project)
				continue

			case "/reset":
				messages = []Message{
					{Role: "system", Content: sysPrompt},
				}
				fmt.Println("\033[1;32m✓ Conversation memory reset. Starting fresh session.\033[0m")
				continue

			case "/approve":
				autoApproveCommands = !autoApproveCommands
				if autoApproveCommands {
					fmt.Println("\033[33m⚡ Auto-approval enabled: commands will execute without prompt.\033[0m")
				} else {
					fmt.Println("\033[1;32m🔒 Command confirmation enabled: commands require user permission.\033[0m")
				}
				continue

			case "/info":
				printSessionInfo(currentProvider, currentModel, project, len(messages), autoApproveCommands)
				continue

			case "/tools":
				printToolsList()
				continue

			case "/skills":
				allSkills := LoadAllSkills()
				if len(allSkills) == 0 {
					fmt.Println("\033[33mNo skills found. Teach the agent with /learn <name> <instructions> or learn_skill.\033[0m")
				} else {
					fmt.Printf("\033[1;38;5;141m⚡ Available Skills (%d):\033[0m\n", len(allSkills))
					for _, s := range allSkills {
						scope := "\033[38;5;75m[project]\033[0m"
						if s.IsGlobal {
							scope = "\033[38;5;177m[global]\033[0m"
						}
						fmt.Printf("  • \033[1;38;5;141m#%-22s\033[0m %-10s \033[38;5;250m%s\033[0m\n", s.Name, scope, s.Description)
					}
					fmt.Println("\n\033[2mTip: Mention \033[1;38;5;141m#<skill-name>\033[0m\033[2m (or \033[1;38;5;141m$<name>\033[0m\033[2m) in any prompt, or \033[38;5;75m/skill <name>\033[0m\033[2m to view details.\033[0m")
				}
				continue

			case "/skill":
				if len(parts) < 2 {
					fmt.Println("\033[33mUsage: /skill <name>\033[0m")
					continue
				}
				target := strings.ToLower(parts[1])
				target = strings.TrimPrefix(target, "#")
				target = strings.TrimPrefix(target, "$")
				allSkills := LoadAllSkills()
				s, ok := allSkills[target]
				if !ok {
					fmt.Printf("\033[31mSkill '%s' not found. Type /skills to list available skills.\033[0m\n", target)
					continue
				}
				scope := "project"
				if s.IsGlobal {
					scope = "global"
				}
				fmt.Printf("\033[1;38;5;141m🎯 Skill: #%s\033[0m \033[2m(%s at %s)\033[0m\n", s.Name, scope, s.Path)
				fmt.Printf("\033[38;5;244mDescription:\033[0m %s\n\n", s.Description)
				rendered := renderMarkdown(s.Content)
				fmt.Println(rendered)
				continue

			case "/learn":
				if len(parts) < 3 {
					fmt.Println("\033[33mUsage: /learn [--global|--project] <name> <instructions...>\033[0m")
					fmt.Println("\033[2mExample: /learn react-rules Always use functional components and Tailwind classes\033[0m")
					continue
				}
				args := parts[1:]
				isGlobal := (project == nil)
				switch args[0] {
				case "--global", "-g":
					isGlobal = true
					args = args[1:]
				case "--project", "-p":
					isGlobal = false
					args = args[1:]
				}
				if len(args) < 2 {
					fmt.Println("\033[33mUsage: /learn [--global|--project] <name> <instructions...>\033[0m")
					continue
				}
				name := args[0]
				name = strings.TrimPrefix(name, "#")
				name = strings.TrimPrefix(name, "$")
				instructions := strings.Join(args[1:], " ")
				path, err := SaveSkill(name, instructions, instructions, isGlobal)
				if err != nil {
					fmt.Printf("\033[31mFailed to save skill: %s\033[0m\n", err)
				} else {
					scopeStr := "project"
					if isGlobal {
						scopeStr = "global"
					}
					fmt.Printf("\033[1;32m✓ Learned %s skill '#%s' (saved to %s)\033[0m\n", scopeStr, name, path)
					fmt.Printf("\033[2mActivate it in any prompt by mentioning #%s or $%s\033[0m\n", name, name)
				}
				continue

			case "/context":
				printContextSummary(messages)
				continue

			case "/providers":
				cfg, err := config.Load()
				if err != nil {
					fmt.Printf("\033[31mError loading config: %s\033[0m\n", err)
					continue
				}
				fmt.Println("\033[1;36mRegistered Providers:\033[0m")
				allProviders := make(map[string]bool)
				for p := range builtinDefaults {
					allProviders[p] = true
				}
				for _, p := range cfg.ProviderOrder {
					allProviders[p] = true
				}
				for p := range cfg.AIProviders {
					allProviders[p] = true
				}

				for name := range allProviders {
					active := ""
					if name == currentProvider {
						active = " \033[1;32m← active\033[0m"
					}
					pc := cfg.AIProviders[name]
					details := fmt.Sprintf("(keys: %d, models: %d)", len(pc.APIKeys), len(pc.Models))
					fmt.Printf("  • \033[1m%-12s\033[0m %-18s%s\n", name, details, active)
				}
				continue

			case "/provider":
				if len(parts) < 2 {
					fmt.Printf("\033[33mActive provider: '%s'. Usage: /provider <name>\033[0m\n", currentProvider)
					continue
				}
				target := strings.ToLower(parts[1])
				cfg, err := config.Load()
				if err != nil {
					fmt.Printf("\033[31mError loading config: %s\033[0m\n", err)
					continue
				}
				if _, exists := cfg.AIProviders[target]; !exists {
					if _, isBuiltin := builtinDefaults[target]; !isBuiltin {
						fmt.Printf("\033[31mProvider '%s' is not configured. Add it with: pkt change ai add provider %s\033[0m\n", target, target)
						continue
					}
				}
				currentProvider = target
				cfg.AIProvider = target

				// Update current model for newly switched provider
				if pc, ok := cfg.AIProviders[target]; ok && pc.DefaultModel != "" {
					currentModel = pc.DefaultModel
				} else if d, ok := builtinDefaults[target]; ok {
					currentModel = d[1]
				} else {
					currentModel = "default"
				}

				_ = config.Save(cfg)
				rl.SetPrompt(formatPrompt(currentProvider, currentModel))
				fmt.Printf("\033[1;32m✓ Switched active provider to '%s' [%s] (saved as default)\033[0m\n", target, currentModel)
				continue

			case "/models":
				cfg, err := config.Load()
				if err != nil {
					fmt.Printf("\033[31mError loading config: %s\033[0m\n", err)
					continue
				}
				pc := cfg.AIProviders[currentProvider]
				fmt.Printf("\033[1;36mModels for '%s':\033[0m\n", currentProvider)
				if len(pc.Models) == 0 {
					defModel := pc.DefaultModel
					if defModel == "" {
						if d, ok := builtinDefaults[currentProvider]; ok {
							defModel = d[1]
						} else {
							defModel = "default"
						}
					}
					fmt.Printf("  • %s \033[1;32m← active\033[0m\n", defModel)
				} else {
					for _, m := range pc.Models {
						active := ""
						if m == currentModel {
							active = " \033[1;32m← active\033[0m"
						}
						fmt.Printf("  • %s%s\n", m, active)
					}
				}
				continue

			case "/model":
				if len(parts) < 2 {
					fmt.Printf("\033[33mActive model: '%s'. Usage: /model <name>\033[0m\n", currentModel)
					continue
				}
				targetModel := parts[1]
				cfg, err := config.Load()
				if err != nil {
					fmt.Printf("\033[31mError loading config: %s\033[0m\n", err)
					continue
				}
				if currentProvider == "" {
					fmt.Println("\033[31mNo active AI provider. Switch provider first with /provider <name>\033[0m")
					continue
				}
				if err := cfg.SetDefaultModel(currentProvider, targetModel); err != nil {
					fmt.Printf("\033[31mFailed to set model: %s\033[0m\n", err)
				} else {
					currentModel = targetModel
					rl.SetPrompt(formatPrompt(currentProvider, currentModel))
					fmt.Printf("\033[1;32m✓ Switched model for '%s' to '%s' (saved as default)\033[0m\n", currentProvider, targetModel)
				}
				continue

			default:
				fmt.Printf("\033[33mUnknown command '%s'. Type /help for available commands.\033[0m\n", cmdName)
				continue
			}
		}

		// Parse @file attachments from user input
		attachments, contextBlock := ExtractContextAttachments(userInput)
		for _, att := range attachments {
			if att.IsDir {
				fmt.Printf("\033[1;38;5;45m📎 Attached directory:\033[0m \033[1;38;5;117m@%s\033[0m \033[38;5;244m(%d entries)\033[0m\n", att.Path, att.LineCount)
			} else {
				fmt.Printf("\033[1;38;5;45m📎 Attached file:\033[0m \033[1;38;5;117m@%s\033[0m \033[38;5;244m(%d lines, %s)\033[0m\n", att.Path, att.LineCount, humanize.Bytes(uint64(att.ByteCount)))
			}
		}

		// Detect and load #skill or $skill mentions
		allSkills := LoadAllSkills()
		activeSkills, skillBlock := ExtractSkillMentions(userInput, allSkills)
		for _, s := range activeSkills {
			fmt.Printf("\033[1;38;5;141m🎯 Skill loaded:\033[0m \033[1;38;5;177m#%s\033[0m \033[38;5;246m— %s\033[0m\n", s.Name, s.Description)
		}

		fullUserPrompt := userInput + contextBlock + skillBlock
		messages = append(messages, Message{Role: "user", Content: fullUserPrompt})

		// Trim context: keep system prompt + last 20 messages to avoid token overflow
		const maxHistory = 20
		if len(messages) > maxHistory+1 {
			messages = append(messages[:1], messages[len(messages)-maxHistory:]...)
		}

		for {
			fmt.Print("\033[38;5;39m●\033[0m \033[2mThinking...\033[0m\r")

			var responseMsg *Message
			var err error
			for attempt := 0; attempt < 3; attempt++ {
				responseMsg, err = SendMessages(messages, currentProvider, definedTools)
				if err == nil {
					break
				}
				errStr := err.Error()
				if strings.Contains(errStr, "429") || strings.Contains(errStr, "rate_limit") || strings.Contains(errStr, "Rate limit") {
					wait := time.Duration(2<<attempt) * time.Second
					fmt.Printf("\033[33m⏳ Rate limited — retrying in %s...\033[0m\r", wait)
					time.Sleep(wait)
					continue
				}
				break
			}
			// Clear thinking line
			fmt.Print("\033[K")

			if err != nil {
				fmt.Printf("\033[31mError: %s\033[0m\n\n", err.Error())
				break
			}

			// Parse reasoning <think> blocks if present
			thinking, cleanContent := ExtractThinking(responseMsg.Content)
			if thinking != "" {
				fmt.Println("\033[2m\033[3m💭 Thinking:\033[0m")
				for _, line := range strings.Split(thinking, "\n") {
					fmt.Printf("\033[2m  %s\033[0m\n", line)
				}
				fmt.Println()
				responseMsg.Content = cleanContent
			}

			// Fallback tool calling parsing for models that emit tool calls in plain text
			if len(responseMsg.ToolCalls) == 0 {
				fallbackCalls := ParseFallbackToolCalls(responseMsg.Content)
				if len(fallbackCalls) > 0 {
					responseMsg.ToolCalls = fallbackCalls
				}
			}

			// Add assistant reply into history
			messages = append(messages, *responseMsg)

			if len(responseMsg.ToolCalls) > 0 {
				if strings.TrimSpace(responseMsg.Content) != "" {
					rendered := RenderMarkdown(responseMsg.Content)
					fmt.Printf("\n%s\n\n", rendered)
				}
				for _, call := range responseMsg.ToolCalls {
					normName := NormalizeToolName(call.Function.Name)
					rawArgs, _ := ParseToolArguments(call.Function.Arguments)
					normArgs := NormalizeToolArgs(normName, rawArgs)

					displayToolInvocation(normName, normArgs)

					// Require explicit user confirmation before executing any tool
					if !autoApproveCommands {
						promptMsg := toolPermissionPrompt(normName, normArgs)
						options := []string{
							"Yes, allow",
							"No, deny",
							"Always allow for this session",
						}
						var selected string
						prompt := &survey.Select{
							Message: promptMsg,
							Options: options,
							Default: options[0],
						}
						err := survey.AskOne(prompt, &selected)
						if err != nil || selected == "No, deny" {
							fmt.Println("  \033[38;5;203m✗ Denied by user\033[0m")
							messages = append(messages, Message{
								Role:       "tool",
								Content:    fmt.Sprintf("Permission denied: user rejected executing %s.", normName),
								ToolCallID: call.ID,
							})
							continue
						}

						if selected == "Always allow for this session" {
							autoApproveCommands = true
							fmt.Println("  \033[38;5;244m(Auto-approval enabled for this session)\033[0m")
						}
					}

					result, execErr := executeTool(normName, normArgs)
					displayToolResult(normName, result, execErr)

					toolOutput := result
					if execErr != nil {
						toolOutput = fmt.Sprintf("Error: %s", execErr.Error())
					}

					messages = append(messages, Message{
						Role:       "tool",
						Content:    toolOutput,
						ToolCallID: call.ID,
					})
				}

				time.Sleep(1000 * time.Millisecond)
				continue
			} else {
				rendered := renderMarkdown(responseMsg.Content)
				fmt.Printf("\n%s\n\n", rendered)
				break
			}
		}
	}
	return nil
}

func toolPermissionPrompt(name string, args map[string]interface{}) string {
	switch name {
	case "run_command":
		cmd, _ := args["command"].(string)
		if cmd == "" {
			return "\033[1;38;5;214m⚙ run_command:\033[0m Run shell command?"
		}
		return fmt.Sprintf("\033[1;38;5;214m⚙ run_command:\033[0m Run '%s'?", cmd)
	case "read_file":
		path, _ := args["path"].(string)
		if path == "" {
			return "\033[1;38;5;75m📖 read_file:\033[0m Read file?"
		}
		return fmt.Sprintf("\033[1;38;5;75m📖 read_file:\033[0m Read '%s'?", path)
	case "write_file":
		path, _ := args["path"].(string)
		if path == "" {
			return "\033[1;38;5;77m✏️ write_file:\033[0m Write to file?"
		}
		return fmt.Sprintf("\033[1;38;5;77m✏️ write_file:\033[0m Write to '%s'?", path)
	case "list_dir":
		path, _ := args["path"].(string)
		if path == "" {
			path = "."
		}
		return fmt.Sprintf("\033[1;38;5;178m📂 list_dir:\033[0m List directory '%s'?", path)
	case "delete_file":
		path, _ := args["path"].(string)
		if path == "" {
			return "\033[1;38;5;203m🗑️ delete_file:\033[0m Delete file?"
		}
		return fmt.Sprintf("\033[1;38;5;203m🗑️ delete_file:\033[0m Delete '%s'?", path)
	case "make_dir":
		path, _ := args["path"].(string)
		if path == "" {
			return "\033[1;38;5;111m📁 make_dir:\033[0m Create directory?"
		}
		return fmt.Sprintf("\033[1;38;5;111m📁 make_dir:\033[0m Create directory '%s'?", path)
	case "get_project_info":
		return "\033[1;38;5;141mℹ️ get_project_info:\033[0m Inspect project metadata?"
	case "learn_skill":
		name, _ := args["name"].(string)
		if name == "" {
			return "\033[1;38;5;141m🧠 learn_skill:\033[0m Learn new skill?"
		}
		return fmt.Sprintf("\033[1;38;5;141m🧠 learn_skill:\033[0m Learn skill '#%s'?", name)
	default:
		return fmt.Sprintf("\033[1;38;5;214m⚙ Execute tool '%s'?\033[0m", name)
	}
}

func formatPrompt(prov, mod string) string {
	return fmt.Sprintf("\033[1;38;5;39mpkt\033[0m \033[2m[\033[0m\033[38;5;75m%s\033[0m\033[2m:\033[0m\033[38;5;111m%s\033[0m\033[2m]\033[0m \033[1;32m❯\033[0m ", prov, mod)
}

func printBanner(provider, model string, project *db.Project) {
	projName := "none"
	projLang := ""
	if project != nil {
		projName = project.Name
		projLang = fmt.Sprintf(" (%s/%s)", project.Language, project.PackageManager)
	}

	fmt.Println("\n\033[38;5;39m┌────────────────────────────────────────────────────────────┐\033[0m")
	fmt.Printf("\033[38;5;39m│\033[0m  \033[1m⚡ pkt chat\033[0m — Autonomous AI Coding Agent               \033[38;5;39m│\033[0m\n")
	fmt.Printf("\033[38;5;39m│\033[0m  Provider: \033[1;32m%-10s\033[0m Model: \033[1;36m%-24s\033[0m   \033[38;5;39m│\033[0m\n", provider, truncate(model, 24))
	fmt.Printf("\033[38;5;39m│\033[0m  Project:  \033[1m%-44s\033[0m   \033[38;5;39m│\033[0m\n", truncate(projName+projLang, 44))
	fmt.Println("\033[38;5;39m│\033[0m                                                            \033[38;5;39m│\033[0m")
	fmt.Println("│  Commands: \033[1;38;5;75m/help\033[0m, \033[1;38;5;141m/skills\033[0m, \033[1;38;5;141m/learn\033[0m, \033[1;38;5;214m/tools\033[0m, \033[1;38;5;75m/clear\033[0m, \033[1;38;5;203m/exit\033[0m   │")
	fmt.Println("│  Shortcuts: \033[1;38;5;141m#<skill>\033[0m \033[2muse skill,\033[0m \033[1;38;5;45m@<file>\033[0m \033[2mattach,\033[0m \033[1;38;5;220mTab\033[0m \033[2mauto\033[0m   │")
	fmt.Println("└────────────────────────────────────────────────────────────┘")
}

func printHelp() {
	fmt.Println("\033[1;38;5;75mChat Commands:\033[0m")
	fmt.Println("  \033[1;38;5;75m/provider <name>\033[0m     Switch active AI provider and save as default")
	fmt.Println("  \033[1;38;5;75m/providers\033[0m           List all registered providers and keys/models")
	fmt.Println("  \033[1;38;5;75m/model <name>\033[0m        Switch active model and save as default")
	fmt.Println("  \033[1;38;5;75m/models\033[0m              List models for current provider")
	fmt.Println("  \033[1;38;5;75m/clear\033[0m               Clear the terminal screen")
	fmt.Println("  \033[1;38;5;75m/reset\033[0m               Clear conversation history and restart memory")
	fmt.Println("  \033[1;38;5;75m/approve\033[0m             Toggle requiring confirmation before running commands")
	fmt.Println("  \033[1;38;5;75m/info\033[0m                Display active session, provider, and project details")
	fmt.Println("  \033[1;38;5;75m/context\033[0m             Inspect message history and token footprint")
	fmt.Println("  \033[1;38;5;75m/help\033[0m                Display this help guide")
	fmt.Println("  \033[1;38;5;203m/exit, /quit\033[0m         Exit the interactive chat session")
	fmt.Println()
	fmt.Println("\033[1;38;5;141mSkill Commands & Workflows:\033[0m")
	fmt.Println("  \033[1;38;5;141m#<skill-name>\033[0m        Activate a skill in prompt (or $<name>)")
	fmt.Println("  \033[1;38;5;141m/skills\033[0m              List all available skills (project & global)")
	fmt.Println("  \033[1;38;5;141m/skill <name>\033[0m        Display full instructions for a specific skill")
	fmt.Println("  \033[1;38;5;141m/learn <name> <ins>\033[0m  Teach the agent a new reusable skill or rule")
	fmt.Println()
	fmt.Println("\033[1;38;5;214mAgent Tool Capabilities:\033[0m")
	fmt.Println("  \033[1;38;5;214m/tools\033[0m               Display list of autonomous tools the agent can invoke")
	fmt.Println()
	fmt.Println("\033[1;36mAuto-Complete & Context Mentions:\033[0m")
	fmt.Println("  • Press \033[1;38;5;220mTab\033[0m at any time to auto-complete commands, providers, models, skills, or files.")
	fmt.Println("  • Type \033[1;38;5;141m#<skill>\033[0m (or \033[1;38;5;141m$<skill>\033[0m) in any prompt to activate and follow a skill.")
	fmt.Println("  • Type \033[1;38;5;45m@filename\033[0m in your prompt to auto-attach file or directory contents.")
}

func printSessionInfo(provider, model string, project *db.Project, messageCount int, autoApprove bool) {
	fmt.Println("\033[1;36mSession Information:\033[0m")
	fmt.Printf("  • Provider:      \033[1m%s\033[0m\n", provider)
	fmt.Printf("  • Model:         \033[1m%s\033[0m\n", model)
	fmt.Printf("  • Context Count: \033[1m%d messages\033[0m\n", messageCount)
	if autoApprove {
		fmt.Println("  • Permissions:   \033[33mAuto-approved (runs without prompt)\033[0m")
	} else {
		fmt.Println("  • Permissions:   \033[1;32mProtected (requires confirmation before running)\033[0m")
	}
	if project != nil {
		fmt.Printf("  • Project:       \033[1m%s\033[0m (%s)\n", project.Name, project.Language)
		fmt.Printf("  • Manager:       \033[1m%s\033[0m\n", project.PackageManager)
		fmt.Printf("  • Location:      \033[1m%s\033[0m\n", project.Path)
	} else {
		cwd, _ := os.Getwd()
		fmt.Printf("  • Location:      \033[1m%s\033[0m (untracked)\n", cwd)
	}
}

func printToolsList() {
	fmt.Println("\033[1;38;5;214m⚙ Available Agent Tools:\033[0m")
	for _, t := range definedTools {
		icon := "🔧"
		switch t.Function.Name {
		case "run_command":
			icon = "⚙️"
		case "read_file":
			icon = "📖"
		case "write_file":
			icon = "✏️"
		case "list_dir":
			icon = "📂"
		case "delete_file":
			icon = "🗑️"
		case "make_dir":
			icon = "📁"
		case "learn_skill":
			icon = "🧠"
		case "get_project_info":
			icon = "ℹ️"
		}
		fmt.Printf("  • %s \033[1;38;5;214m%-18s\033[0m \033[38;5;250m%s\033[0m\n", icon, t.Function.Name, t.Function.Description)
	}
}

func printContextSummary(messages []Message) {
	fmt.Println("\033[1;36mConversation Context:\033[0m")
	totalChars := 0
	counts := make(map[string]int)
	for _, m := range messages {
		counts[m.Role]++
		totalChars += len(m.Content)
	}
	fmt.Printf("  • Total Messages:  \033[1m%d\033[0m\n", len(messages))
	fmt.Printf("  • System / User:   \033[1m%d / %d\033[0m\n", counts["system"], counts["user"])
	fmt.Printf("  • Assistant/Tool:  \033[1m%d / %d\033[0m\n", counts["assistant"], counts["tool"])
	fmt.Printf("  • Estimated Size:  \033[1m~%d characters\033[0m (~%d tokens)\n", totalChars, totalChars/4)
}

func displayToolInvocation(name string, args map[string]interface{}) {
	switch name {
	case "run_command":
		cmd, _ := args["command"].(string)
		fmt.Printf("\033[38;5;214m⚙ run_command\033[0m \033[38;5;244m›\033[0m \033[1m%s\033[0m\n", cmd)
	case "read_file":
		path, _ := args["path"].(string)
		fmt.Printf("\033[38;5;75m📖 read_file\033[0m \033[38;5;244m›\033[0m \033[1m%s\033[0m\n", path)
	case "write_file":
		path, _ := args["path"].(string)
		fmt.Printf("\033[38;5;77m✏️ write_file\033[0m \033[38;5;244m›\033[0m \033[1m%s\033[0m\n", path)
	case "list_dir":
		path, _ := args["path"].(string)
		if path == "" {
			path = "."
		}
		fmt.Printf("\033[38;5;178m📂 list_dir\033[0m \033[38;5;244m›\033[0m \033[1m%s\033[0m\n", path)
	case "delete_file":
		path, _ := args["path"].(string)
		fmt.Printf("\033[38;5;203m🗑️ delete_file\033[0m \033[38;5;244m›\033[0m \033[1m%s\033[0m\n", path)
	case "make_dir":
		path, _ := args["path"].(string)
		fmt.Printf("\033[38;5;111m📁 make_dir\033[0m \033[38;5;244m›\033[0m \033[1m%s\033[0m\n", path)
	case "get_project_info":
		fmt.Println("\033[38;5;141mℹ️ get_project_info\033[0m")
	case "learn_skill":
		name, _ := args["name"].(string)
		fmt.Printf("\033[38;5;141m🧠 learn_skill\033[0m \033[38;5;244m›\033[0m \033[1m#%s\033[0m\n", name)
	default:
		fmt.Printf("\033[90m⚙ %s\033[0m\n", name)
	}
}

func displayToolResult(name string, result string, err error) {
	if err != nil {
		fmt.Printf("  \033[38;5;203m✗ Failed:\033[0m %s\n", err.Error())
		return
	}

	switch name {
	case "run_command":
		lines := strings.Split(strings.TrimSpace(result), "\n")
		if len(lines) == 1 && (lines[0] == "" || strings.Contains(lines[0], "silently")) {
			fmt.Println("  \033[38;5;71m✓ Done\033[0m \033[2m(silent)\033[0m")
		} else if len(lines) <= 3 {
			for _, l := range lines {
				fmt.Printf("  \033[2m› %s\033[0m\n", l)
			}
		} else {
			fmt.Printf("  \033[38;5;71m✓ Done\033[0m \033[2m(%d lines output)\033[0m\n", len(lines))
		}
	case "read_file":
		lines := strings.Split(result, "\n")
		fmt.Printf("  \033[38;5;71m✓ Read %d lines\033[0m\n", len(lines))
	case "write_file":
		fmt.Println("  \033[38;5;71m✓ Saved\033[0m")
	case "list_dir":
		lines := strings.Split(strings.TrimSpace(result), "\n")
		fmt.Printf("  \033[38;5;71m✓ %d items listed\033[0m\n", len(lines))
	case "learn_skill":
		fmt.Println("  \033[38;5;71m✓ Skill saved to library\033[0m")
	default:
		fmt.Println("  \033[38;5;71m✓ Done\033[0m")
	}
}

func executeTool(name string, args map[string]interface{}) (string, error) {
	switch name {
	case "get_project_info":
		cwd, _ := os.Getwd()
		project, err := db.GetProjectByPath(cwd)
		if err != nil {
			return "Not inside a tracked pkt project folder.", nil
		}
		branch := "unknown"
		if out, err := exec.Command("git", "rev-parse", "--abbrev-ref", "HEAD").Output(); err == nil {
			branch = strings.TrimSpace(string(out))
		}
		return fmt.Sprintf("Project Name: %s\nLanguage: %s\nPackage Manager: %s\nPath: %s\nGit Branch: %s",
			project.Name, project.Language, project.PackageManager, project.Path, branch), nil

	case "list_dir":
		path, _ := args["path"].(string)
		if path == "" {
			path = "."
		}
		entries, err := os.ReadDir(path)
		if err != nil {
			return "", err
		}
		var out []string
		for _, e := range entries {
			suffix := ""
			if e.IsDir() {
				suffix = "/"
			}
			out = append(out, e.Name()+suffix)
		}
		if len(out) == 0 {
			return "Directory is empty.", nil
		}
		return strings.Join(out, "\n"), nil

	case "read_file":
		path, _ := args["path"].(string)
		if path == "" {
			return "", fmt.Errorf("missing path parameter")
		}
		info, err := os.Stat(path)
		if err != nil {
			return "", err
		}
		if info.Size() > 1024*1024 {
			return "", fmt.Errorf("file %s is larger than 1MB", path)
		}
		b, err := os.ReadFile(path)
		if err != nil {
			return "", err
		}
		return string(b), nil

	case "write_file":
		path, _ := args["path"].(string)
		if path == "" {
			return "", fmt.Errorf("missing path parameter")
		}
		content, _ := args["content"].(string)
		dir := filepath.Dir(path)
		if dir != "" && dir != "." {
			_ = os.MkdirAll(dir, 0755)
		}
		err := os.WriteFile(path, []byte(content), 0644)
		if err != nil {
			return "", err
		}
		return fmt.Sprintf("File successfully written: %s (%d bytes)", path, len(content)), nil

	case "delete_file":
		path, _ := args["path"].(string)
		if path == "" {
			return "", fmt.Errorf("missing path parameter")
		}
		if err := os.Remove(path); err != nil {
			return "", err
		}
		return fmt.Sprintf("File successfully deleted: %s", path), nil

	case "make_dir":
		path, _ := args["path"].(string)
		if path == "" {
			return "", fmt.Errorf("missing path parameter")
		}
		if err := os.MkdirAll(path, 0755); err != nil {
			return "", err
		}
		return fmt.Sprintf("Directory successfully created: %s", path), nil

	case "run_command":
		cmdStr, _ := args["command"].(string)
		if cmdStr == "" {
			return "", fmt.Errorf("missing command parameter")
		}
		ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
		defer cancel()

		cmd := exec.CommandContext(ctx, "sh", "-c", cmdStr)
		out, err := cmd.CombinedOutput()
		if ctx.Err() == context.DeadlineExceeded {
			return "", fmt.Errorf("command timed out after 60 seconds")
		}
		if err != nil {
			return fmt.Sprintf("Command failed: %s\nOutput:\n%s", err.Error(), string(out)), nil
		}
		if len(out) == 0 {
			return "Command executed silently (no output).", nil
		}
		outStr := string(out)
		if len(outStr) > 10000 {
			outStr = outStr[:10000] + "\n... [truncated: output exceeded 10,000 characters]"
		}
		return outStr, nil

	case "learn_skill":
		name, _ := args["name"].(string)
		if name == "" {
			return "", fmt.Errorf("missing name parameter")
		}
		description, _ := args["description"].(string)
		instructions, _ := args["instructions"].(string)
		if instructions == "" {
			return "", fmt.Errorf("missing instructions parameter")
		}
		global, _ := args["global"].(bool)
		path, err := SaveSkill(name, description, instructions, global)
		if err != nil {
			return "", err
		}
		return fmt.Sprintf("Skill '#%s' successfully learned and saved to %s. Users can invoke it anytime with #%s or $%s.", name, path, name, name), nil

	default:
		return "", fmt.Errorf("unknown tool: %s", name)
	}
}

func renderMarkdown(content string) string {
	return RenderMarkdown(content)
}

func getTerminalWidth() int {
	fd := int(os.Stdout.Fd())
	if term.IsTerminal(fd) {
		w, _, err := term.GetSize(fd)
		if err == nil && w > 40 {
			if w > 120 {
				return 110
			}
			return w - 4
		}
	}
	return 100
}

func chatHistoryPath() string {
	home, err := os.UserHomeDir()
	if err != nil {
		return filepath.Join(os.TempDir(), "pkt_chat_history")
	}
	pktDir := filepath.Join(home, ".pkt")
	_ = os.MkdirAll(pktDir, 0755)
	return filepath.Join(pktDir, "chat_history")
}

func truncate(s string, max int) string {
	if len(s) <= max {
		return s
	}
	return s[:max-3] + "..."
}
