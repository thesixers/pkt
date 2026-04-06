package ai

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"regexp"
	"strings"

	"github.com/genesix/pkt/internal/config"
)

type Tool struct {
	Type     string       `json:"type"`
	Function ToolFunction `json:"function"`
}

type ToolFunction struct {
	Name        string         `json:"name"`
	Description string         `json:"description"`
	Parameters  ToolParameters `json:"parameters"`
}

type ToolParameters struct {
	Type       string                 `json:"type"`
	Properties map[string]interface{} `json:"properties"`
	Required   []string               `json:"required,omitempty"`
}

type ToolCall struct {
	ID       string       `json:"id"`
	Type     string       `json:"type"`
	Function CallFunction `json:"function"`
}

type CallFunction struct {
	Name      string `json:"name"`
	Arguments string `json:"arguments"`
}

type ChatCompletionRequest struct {
	Model       string    `json:"model"`
	Messages    []Message `json:"messages"`
	Tools       []Tool    `json:"tools,omitempty"`
	Temperature float64   `json:"temperature"`
}

type Message struct {
	Role       string     `json:"role"`
	Content    string     `json:"content"`
	Name       string     `json:"name,omitempty"`
	ToolCalls  []ToolCall `json:"tool_calls,omitempty"`
	ToolCallID string     `json:"tool_call_id,omitempty"`
}

type ChatCompletionResponse struct {
	Choices []struct {
		Message Message `json:"message"`
	} `json:"choices"`
}

func AskAI(systemPrompt, userPrompt, preferredProvider string) (string, error) {
	messages := []Message{
		{Role: "system", Content: systemPrompt},
		{Role: "user", Content: userPrompt},
	}
	msg, err := SendMessages(messages, preferredProvider, nil)
	if err != nil {
		return "", err
	}
	return msg.Content, nil
}

func SendMessages(messages []Message, preferredProvider string, tools []Tool) (*Message, error) {
	cfg, err := config.Load()
	if err != nil {
		return nil, fmt.Errorf("failed to load config: %w", err)
	}

	provider := cfg.AIProvider
	if preferredProvider != "" {
		provider = strings.ToLower(preferredProvider)
	}

	if provider == "" {
		return nil, fmt.Errorf("no AI provider configured. Run 'pkt config set-ai <provider> <key>'")
	}

	apiKey := cfg.AIKeys[provider]
	if apiKey == "" {
		if cfg.AIKey != "" && cfg.AIProvider == provider {
			apiKey = cfg.AIKey
			cfg.AIKeys[provider] = apiKey
			config.Save(cfg)
		} else {
			return nil, fmt.Errorf("API Key not set for provider '%s'. Run 'pkt config set-ai %s <key>'", provider, provider)
		}
	}

	var baseURL, defaultModel string
	switch provider {
	case "groq":
		baseURL = "https://api.groq.com/openai/v1/chat/completions"
		defaultModel = "llama-3.1-8b-instant"
	case "gemini":
		baseURL = "https://generativelanguage.googleapis.com/v1beta/openai/chat/completions"
		defaultModel = "gemini-1.5-flash"
	case "openai":
		baseURL = "https://api.openai.com/v1/chat/completions"
		defaultModel = "gpt-4o-mini"
	default:
		return nil, fmt.Errorf("unsupported provider '%s'. Use openai, gemini, or groq", provider)
	}

	model := cfg.AIModels[provider]
	if model == "" {
		model = defaultModel
	}

	reqBody := ChatCompletionRequest{
		Model:       model,
		Messages:    messages,
		Tools:       tools,
		Temperature: 0.1, // Explicit limit natively caching strictly stable schemas
	}

	jsonBody, err := json.Marshal(reqBody)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("POST", baseURL, bytes.NewBuffer(jsonBody))
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+apiKey)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		errStr := string(bodyBytes)

		// Self-heal Llama's Groq XML Tool-Calling hallucination natively
		if strings.Contains(errStr, "failed_generation") {
			type groqError struct {
				Error struct {
					FailedGeneration string `json:"failed_generation"`
				} `json:"error"`
			}
			var ge groqError
			if json.Unmarshal(bodyBytes, &ge) == nil && ge.Error.FailedGeneration != "" {
				fg := ge.Error.FailedGeneration

				re := regexp.MustCompile(`<function=([a-zA-Z_0-9_]+)[^\{]*(\{.*\})`)
				matches := re.FindStringSubmatch(fg)

				if len(matches) == 3 {
					funcName := matches[1]
					funcArgs := strings.TrimSpace(matches[2])

					syntheticMessage := Message{
						Role: "assistant",
						ToolCalls: []ToolCall{
							{
								ID:   "call_synthetic_groq_" + funcName,
								Type: "function",
								Function: CallFunction{
									Name:      funcName,
									Arguments: funcArgs,
								},
							},
						},
					}
					return &syntheticMessage, nil
				}
			}
		}

		return nil, fmt.Errorf("API error (%d): %s", resp.StatusCode, errStr)
	}

	var parsedResp ChatCompletionResponse
	if err := json.NewDecoder(resp.Body).Decode(&parsedResp); err != nil {
		return nil, err
	}

	if len(parsedResp.Choices) == 0 {
		return nil, fmt.Errorf("no response from AI")
	}

	return &parsedResp.Choices[0].Message, nil
}
