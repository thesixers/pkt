package ai

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

// AttachedContext represents a project file or directory attached to a prompt.
type AttachedContext struct {
	Path      string
	IsDir     bool
	LineCount int
	ByteCount int64
	Content   string
}

// ExtractContextAttachments searches input for @path mentions, reads local files/folders if
// present, and returns the slice of attachments and an appended prompt block.
func ExtractContextAttachments(input string) ([]AttachedContext, string) {
	re := regexp.MustCompile(`@([a-zA-Z0-9_\-./]+)`)
	matches := re.FindAllStringSubmatch(input, -1)
	if len(matches) == 0 {
		return nil, ""
	}

	seen := make(map[string]bool)
	var attachments []AttachedContext
	var promptBuilder strings.Builder

	for _, match := range matches {
		targetPath := match[1]
		if targetPath == "" || targetPath == "." || targetPath == ".." || seen[targetPath] {
			continue
		}

		info, err := os.Stat(targetPath)
		if err != nil {
			continue
		}
		seen[targetPath] = true

		if info.IsDir() {
			entries, err := os.ReadDir(targetPath)
			if err != nil {
				continue
			}
			var entryNames []string
			for i, e := range entries {
				if i >= 50 {
					entryNames = append(entryNames, fmt.Sprintf("... (%d more items)", len(entries)-50))
					break
				}
				suffix := ""
				if e.IsDir() {
					suffix = "/"
				}
				entryNames = append(entryNames, e.Name()+suffix)
			}
			listing := strings.Join(entryNames, "\n")
			attachments = append(attachments, AttachedContext{
				Path:      targetPath,
				IsDir:     true,
				LineCount: len(entries),
				Content:   listing,
			})
			fmt.Fprintf(&promptBuilder, "\n\n[Attached Directory: %s]\n%s\n", targetPath, listing)
		} else {
			const maxBytes = 256 * 1024 // 256KB cap
			data, err := os.ReadFile(targetPath)
			if err != nil {
				continue
			}
			truncated := false
			if len(data) > maxBytes {
				data = data[:maxBytes]
				truncated = true
			}
			content := string(data)
			lines := strings.Split(content, "\n")
			lineCount := len(lines)
			if truncated {
				content += "\n... [truncated: file exceeds 256KB]"
			}

			attachments = append(attachments, AttachedContext{
				Path:      targetPath,
				IsDir:     false,
				LineCount: lineCount,
				ByteCount: info.Size(),
				Content:   content,
			})

			ext := strings.TrimPrefix(filepath.Ext(targetPath), ".")
			fmt.Fprintf(&promptBuilder, "\n\n[Attached File: %s (%d lines)]\n```%s\n%s\n```\n", targetPath, lineCount, ext, content)
		}
	}

	return attachments, promptBuilder.String()
}

// ExtractThinking extracts <think>...</think> or <thought>...</thought> reasoning blocks.
func ExtractThinking(content string) (thinking string, cleanResponse string) {
	thinkRe := regexp.MustCompile(`(?is)<(?:think|thought)>(.*?)</(?:think|thought)>`)
	matches := thinkRe.FindAllStringSubmatch(content, -1)
	if len(matches) == 0 {
		return "", content
	}

	var thinkParts []string
	for _, m := range matches {
		if len(m) > 1 {
			trimmed := strings.TrimSpace(m[1])
			if trimmed != "" {
				thinkParts = append(thinkParts, trimmed)
			}
		}
	}

	thinking = strings.Join(thinkParts, "\n\n")
	cleanResponse = strings.TrimSpace(thinkRe.ReplaceAllString(content, ""))
	return thinking, cleanResponse
}

// ParseFallbackToolCalls extracts embedded tool calls from models that output tool calls in plain text.
func ParseFallbackToolCalls(content string) []ToolCall {
	var calls []ToolCall

	// 1. XML style: <function=tool_name>{...}</function>
	funcRe := regexp.MustCompile(`(?is)<function=([a-zA-Z0-9_]+)[^\{]*(\{.*?\})\s*</function>`)
	for _, match := range funcRe.FindAllStringSubmatch(content, -1) {
		name := NormalizeToolName(match[1])
		calls = append(calls, ToolCall{
			ID:   fmt.Sprintf("call_xml_%s_%d", name, len(calls)+1),
			Type: "function",
			Function: CallFunction{
				Name:      name,
				Arguments: strings.TrimSpace(match[2]),
			},
		})
	}
	if len(calls) > 0 {
		return calls
	}

	// 2. <tool_call>{"name": "...", "arguments": {...}}</tool_call>
	toolCallRe := regexp.MustCompile(`(?is)<tool_call>\s*(\{.*?\})\s*</tool_call>`)
	for _, match := range toolCallRe.FindAllStringSubmatch(content, -1) {
		var obj struct {
			Name      string          `json:"name"`
			Arguments json.RawMessage `json:"arguments"`
		}
		if json.Unmarshal([]byte(match[1]), &obj) == nil && obj.Name != "" {
			argsStr := string(obj.Arguments)
			if argsStr == "" {
				argsStr = "{}"
			}
			calls = append(calls, ToolCall{
				ID:   fmt.Sprintf("call_tag_%s_%d", obj.Name, len(calls)+1),
				Type: "function",
				Function: CallFunction{
					Name:      NormalizeToolName(obj.Name),
					Arguments: argsStr,
				},
			})
		}
	}

	return calls
}

// ParseToolArguments safely parses tool arguments into a map, handling code fences and unescaped strings.
func ParseToolArguments(rawArgs string) (map[string]interface{}, error) {
	trimmed := strings.TrimSpace(rawArgs)
	if trimmed == "" {
		return make(map[string]interface{}), nil
	}

	// Strip markdown code fences if model wrapped args in ```json ... ```
	if strings.HasPrefix(trimmed, "```") {
		lines := strings.Split(trimmed, "\n")
		if len(lines) >= 2 {
			if strings.HasPrefix(lines[len(lines)-1], "```") {
				lines = lines[1 : len(lines)-1]
			} else {
				lines = lines[1:]
			}
			trimmed = strings.TrimSpace(strings.Join(lines, "\n"))
		}
	}

	var args map[string]interface{}
	if err := json.Unmarshal([]byte(trimmed), &args); err == nil {
		return args, nil
	}

	// Try extracting outermost JSON object
	start := strings.Index(trimmed, "{")
	end := strings.LastIndex(trimmed, "}")
	if start != -1 && end != -1 && end > start {
		sub := trimmed[start : end+1]
		if err := json.Unmarshal([]byte(sub), &args); err == nil {
			return args, nil
		}
	}

	// Handle double-escaped JSON string
	var unescaped string
	if err := json.Unmarshal([]byte(trimmed), &unescaped); err == nil {
		if err := json.Unmarshal([]byte(unescaped), &args); err == nil {
			return args, nil
		}
	}

	return nil, fmt.Errorf("unable to parse arguments as JSON: %s", rawArgs)
}

// NormalizeToolName maps alternative tool names and common variations to canonical tool names.
func NormalizeToolName(name string) string {
	lower := strings.ToLower(strings.TrimSpace(name))
	switch lower {
	case "view_file", "cat", "read", "open_file", "show_file":
		return "read_file"
	case "create_file", "write", "edit_file", "save_file", "update_file":
		return "write_file"
	case "bash", "sh", "exec", "shell", "terminal", "execute_command", "cmd":
		return "run_command"
	case "ls", "dir", "list_files", "browse_dir":
		return "list_dir"
	case "rm", "remove_file", "unlink":
		return "delete_file"
	case "mkdir", "create_dir":
		return "make_dir"
	case "project_info", "info", "get_project":
		return "get_project_info"
	case "save_skill", "create_skill", "add_skill":
		return "learn_skill"
	default:
		return lower
	}
}

// NormalizeToolArgs maps alternate argument keys (e.g. "cmd" -> "command", "file" -> "path").
func NormalizeToolArgs(toolName string, args map[string]interface{}) map[string]interface{} {
	if args == nil {
		return make(map[string]interface{})
	}

	switch toolName {
	case "read_file", "delete_file":
		if _, ok := args["path"]; !ok {
			for _, k := range []string{"file", "filename", "filepath", "target"} {
				if v, exists := args[k]; exists {
					args["path"] = v
					break
				}
			}
		}

	case "write_file":
		if _, ok := args["path"]; !ok {
			for _, k := range []string{"file", "filename", "filepath", "target"} {
				if v, exists := args[k]; exists {
					args["path"] = v
					break
				}
			}
		}
		if _, ok := args["content"]; !ok {
			for _, k := range []string{"data", "text", "body", "code", "contents"} {
				if v, exists := args[k]; exists {
					args["content"] = v
					break
				}
			}
		}

	case "run_command":
		if _, ok := args["command"]; !ok {
			for _, k := range []string{"cmd", "script", "exec"} {
				if v, exists := args[k]; exists {
					args["command"] = v
					break
				}
			}
		}

	case "list_dir", "make_dir":
		if _, ok := args["path"]; !ok {
			for _, k := range []string{"dir", "directory", "folder"} {
				if v, exists := args[k]; exists {
					args["path"] = v
					break
				}
			}
		}

	case "learn_skill":
		if _, ok := args["instructions"]; !ok {
			for _, k := range []string{"content", "prompt", "guidelines", "rules"} {
				if v, exists := args[k]; exists {
					args["instructions"] = v
					break
				}
			}
		}
		if _, ok := args["description"]; !ok {
			for _, k := range []string{"desc", "summary", "about"} {
				if v, exists := args[k]; exists {
					args["description"] = v
					break
				}
			}
		}
	}

	return args
}
