package ai

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/charmbracelet/glamour"
	"github.com/chzyer/readline"
	"github.com/genesix/pkt/internal/db"
)

var definedTools = []Tool{
	{
		Type: "function",
		Function: ToolFunction{
			Name:        "get_project_info",
			Description: "Get framework, language, and manager for the active project.",
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
					"path": map[string]interface{}{"type": "string", "description": "Absolute or relative path"},
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
					"path": map[string]interface{}{"type": "string"},
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
					"path":    map[string]interface{}{"type": "string"},
					"content": map[string]interface{}{"type": "string"},
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
					"path": map[string]interface{}{"type": "string"},
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
					"path": map[string]interface{}{"type": "string"},
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
					"command": map[string]interface{}{"type": "string"},
				},
				Required: []string{"command"},
			},
		},
	},
}

// StartChatSession enters the interactive REPL.
func StartChatSession(provider string, projectContext string) error {
	cwd, _ := os.Getwd()
	project, err := db.GetProjectByPath(cwd)

	sysPrompt := "You are a powerful Autonomous Coding Agent built inherently into the pkt CLI. You have direct access to the user's terminal and file system. IMPORTANT: If the user asks a question about the project, tools, or environment (like 'list the commands' or 'what is this codebase'), you MUST autonomously use your tools (like run_command or read_file) to find the answer and then report back. However, if the user simply says a generic greeting like 'hello', 'hi', or makes conversational small-talk, reply normally WITHOUT invoking any tools.\n\nCRITICAL CONSTRAINTS:\n1. If the user asks a hypothetical question like 'can you create a file?', NEVER invoke the tool to demonstrate. Just answer 'Yes, I can'.\n2. NEVER invoke state-changing tools ('write_file', 'delete_file', 'make_dir', or destructive 'run_command') without explicitly being told to do so. Always wait for the user to say 'create a file named X' before securely executing."
	if err == nil {
		sysPrompt += fmt.Sprintf("\nEnvironment: You are inside a %s project managed natively by %s located entirely at %s.", project.Language, project.PackageManager, project.Path)
	}
	if projectContext != "" {
		sysPrompt += fmt.Sprintf("\n\nProject README Context:\n%s", projectContext)
	}

	messages := []Message{
		{Role: "system", Content: sysPrompt},
	}

	fmt.Println("\n\033[1;36m╭──────────────────────────────────────────────────────────╮\033[0m")
	fmt.Println("\033[1;36m│\033[0m  🚀 \033[1mpkt chat\033[0m - Autonomous Agent Session Active           \033[1;36m│\033[0m")
	fmt.Println("\033[1;36m│\033[0m  Type 'exit' or 'quit' to close the terminal gracefully. \033[1;36m│\033[0m")
	fmt.Println("\033[1;36m╰──────────────────────────────────────────────────────────╯\033[0m")

	rl, err := readline.NewEx(&readline.Config{
		Prompt:          "\033[1;32m╰─❯\033[0m ",
		HistoryFile:     "/tmp/pkt_chat_history.tmp",
		InterruptPrompt: "^C",
		EOFPrompt:       "exit",
	})
	if err != nil {
		fmt.Printf("Error initializing readline: %s\n", err)
		return err
	}
	defer rl.Close()

	for {
		fmt.Print("\n\033[1;32m╭─ You\033[0m\n")
		line, err := rl.Readline()
		if err != nil { // handles EOF (Ctrl+D) and Interrupt (Ctrl+C)
			break
		}

		userInput := strings.TrimSpace(line)
		if userInput == "exit" || userInput == "quit" {
			fmt.Println("Goodbye!")
			break
		}
		if userInput == "" {
			continue
		}

		messages = append(messages, Message{Role: "user", Content: userInput})

		// Trim context: keep system prompt + last 20 messages to avoid token overflow
		const maxHistory = 20
		if len(messages) > maxHistory+1 {
			messages = append(messages[:1], messages[len(messages)-maxHistory:]...)
		}

		for {
			fmt.Print("\033[1;36m╭─ 🤖 pkt-ai\033[0m \033[2m(thinking...)\033[0m\r")

			// Retry up to 3 times on rate limit (429) with exponential backoff
			var responseMsg *Message
			var err error
			for attempt := 0; attempt < 3; attempt++ {
				responseMsg, err = SendMessages(messages, provider, definedTools)
				if err == nil {
					break
				}
				// Check if it's a rate limit error — if so, wait and retry
				errStr := err.Error()
				if strings.Contains(errStr, "429") || strings.Contains(errStr, "rate_limit") || strings.Contains(errStr, "Rate limit") {
					wait := time.Duration(2<<attempt) * time.Second // 2s, 4s, 8s
					fmt.Printf("\033[33m⏳ Rate limited — retrying in %s...\033[0m\r", wait)
					time.Sleep(wait)
					continue
				}
				// Non-rate-limit error, stop retrying
				break
			}
			// Clear thinking/retry text
			fmt.Print("\033[K")

			if err != nil {
				fmt.Printf("\033[31mError: %s\033[0m\n", err.Error())
				break
			}

			// Add the assistant's reply into history
			messages = append(messages, *responseMsg)

			if len(responseMsg.ToolCalls) > 0 {
				fmt.Println("\033[1;36m╭─ 🤖 pkt-ai\033[0m \033[1;35m[Executing Core Tools]\033[0m")

				// Execute tools sequentially
				for _, call := range responseMsg.ToolCalls {
					result := executeToolCall(call.Function)

					// Print out what tool was called softly
					fmt.Printf("\033[1;36m│\033[0m \033[90m⚙️ Invoking natively: \033[1m%s\033[0m\n", call.Function.Name)

					messages = append(messages, Message{
						Role:       "tool",
						Content:    result,
						ToolCallID: call.ID,
					})
				}
				// Throttle API requests by 1.5 seconds to prevent rate-limit crashes on free-tier keys
				time.Sleep(1500 * time.Millisecond)

				// Loop back immediately sequentially to Ask the AI over again since tool completed securely!
				continue
			} else {
				// Initialize glamour markdown renderer
				r, _ := glamour.NewTermRenderer(
					glamour.WithAutoStyle(),
					glamour.WithWordWrap(100),
				)
				out, err := r.Render(responseMsg.Content)
				if err != nil {
					out = responseMsg.Content
				}

				// Output rendered markdown natively
				fmt.Printf("\033[1;36m╭─ 🤖 pkt-ai\033[0m\n%s", out)
				break
			}
		}
	}
	return nil
}

func executeToolCall(call CallFunction) string {
	var args map[string]interface{}
	if err := json.Unmarshal([]byte(call.Arguments), &args); err != nil {
		return fmt.Sprintf("Error parsing arguments: %s", err.Error())
	}

	switch call.Name {
	case "get_project_info":
		cwd, _ := os.Getwd()
		project, err := db.GetProjectByPath(cwd)
		if err != nil {
			return "Not inside a tracked pkt project folder."
		}
		return fmt.Sprintf("Project Name: %s\nLanguage: %s\nPackage Manager: %s\nPath: %s", project.Name, project.Language, project.PackageManager, project.Path)

	case "list_dir":
		path, _ := args["path"].(string)
		if path == "" {
			path = "."
		}
		entries, err := os.ReadDir(path)
		if err != nil {
			return err.Error()
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
			return "Directory is empty."
		}
		return strings.Join(out, "\n")

	case "read_file":
		path, _ := args["path"].(string)
		b, err := os.ReadFile(path)
		if err != nil {
			return err.Error()
		}
		return string(b)

	case "write_file":
		path, _ := args["path"].(string)
		content, _ := args["content"].(string)
		dir := filepath.Dir(path)
		if dir != "" && dir != "." {
			_ = os.MkdirAll(dir, 0755)
		}
		err := os.WriteFile(path, []byte(content), 0644)
		if err != nil {
			return err.Error()
		}
		return "File successfully written to " + path

	case "delete_file":
		path, _ := args["path"].(string)
		if err := os.Remove(path); err != nil {
			return err.Error()
		}
		return "File successfully deleted: " + path

	case "make_dir":
		path, _ := args["path"].(string)
		if err := os.MkdirAll(path, 0755); err != nil {
			return err.Error()
		}
		return "Directory successfully created: " + path

	case "run_command":
		cmdStr, _ := args["command"].(string)
		cmd := exec.Command("sh", "-c", cmdStr)
		out, err := cmd.CombinedOutput()
		if err != nil {
			return fmt.Sprintf("Command failed: %s\nOutput:\n%s", err.Error(), string(out))
		}
		if len(out) == 0 {
			return "Command executed silently (no output)."
		}
		return string(out)

	default:
		return fmt.Sprintf("Unknown tool invoked: %s", call.Name)
	}
}
