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

// builtinDefaults holds the default base URL and model for known providers.
var builtinDefaults = map[string][2]string{
	"openai": {"https://api.openai.com/v1/chat/completions", "gpt-4o-mini"},
	"groq":   {"https://api.groq.com/openai/v1/chat/completions", "llama-3.1-8b-instant"},
	"gemini": {"https://generativelanguage.googleapis.com/v1beta/openai/chat/completions", "gemini-1.5-flash"},
	"ollama": {"http://localhost:11434/v1/chat/completions", "llama3"},
	"local":  {"http://localhost:1234/v1/chat/completions", "local-model"},
}

// localProviders are providers that do not require an API key.
var localProviders = map[string]bool{
	"ollama": true,
	"local":  true,
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

	providerName := cfg.AIProvider
	if preferredProvider != "" {
		providerName = strings.ToLower(preferredProvider)
	}
	if providerName == "" {
		return nil, fmt.Errorf("no AI provider configured. Run 'pkt config ai <provider>'")
	}

	// Resolve the provider config from the registry
	pc := cfg.AIProviders[providerName]

	// Determine base URL and default model
	defaults, known := builtinDefaults[providerName]
	if !known && pc.BaseURL == "" {
		return nil, fmt.Errorf("unknown provider '%s'. Use openai, groq, gemini, ollama, local, or register a custom one", providerName)
	}

	baseURL := pc.BaseURL
	if baseURL == "" {
		baseURL = defaults[0]
	}

	model := pc.Model
	if model == "" && known {
		model = defaults[1]
	}
	if model == "" {
		model = "default"
	}

	// Require API key only for non-local providers
	apiKey := pc.APIKey
	if apiKey == "" && !localProviders[providerName] {
		return nil, fmt.Errorf("no API key set for '%s'. Run: pkt config set-ai %s <your-api-key>", providerName, providerName)
	}

	reqBody := ChatCompletionRequest{
		Model:       model,
		Messages:    messages,
		Tools:       tools,
		Temperature: 0.1,
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
	if apiKey != "" {
		req.Header.Set("Authorization", "Bearer "+apiKey)
	}

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		errStr := string(bodyBytes)

		// Self-heal Llama/Groq XML tool-calling hallucination
		if strings.Contains(errStr, "failed_generation") {
			type groqError struct {
				Error struct {
					FailedGeneration string `json:"failed_generation"`
				} `json:"error"`
			}
			var ge groqError
			if json.Unmarshal(bodyBytes, &ge) == nil && ge.Error.FailedGeneration != "" {
				re := regexp.MustCompile(`<function=([a-zA-Z_0-9_]+)[^\{]*(\{.*\})`)
				matches := re.FindStringSubmatch(ge.Error.FailedGeneration)
				if len(matches) == 3 {
					return &Message{
						Role: "assistant",
						ToolCalls: []ToolCall{{
							ID:   "call_synthetic_groq_" + matches[1],
							Type: "function",
							Function: CallFunction{
								Name:      matches[1],
								Arguments: strings.TrimSpace(matches[2]),
							},
						}},
					}, nil
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
