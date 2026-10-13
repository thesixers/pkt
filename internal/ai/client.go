package ai

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/genesix/pkt/internal/config"
)

var (
	cooldownMu   sync.Mutex
	keyCooldowns = make(map[string]time.Time)
)

func isKeyCooledDown(key string) bool {
	cooldownMu.Lock()
	defer cooldownMu.Unlock()
	until, exists := keyCooldowns[key]
	if !exists {
		return false
	}
	if time.Now().After(until) {
		delete(keyCooldowns, key)
		return false
	}
	return true
}

func setKeyCooldown(key string, duration time.Duration) {
	cooldownMu.Lock()
	defer cooldownMu.Unlock()
	keyCooldowns[key] = time.Now().Add(duration)
}

func getEarliestCooldown(keys []string) time.Duration {
	cooldownMu.Lock()
	defer cooldownMu.Unlock()
	var earliest time.Time
	now := time.Now()
	for _, k := range keys {
		until, exists := keyCooldowns[k]
		if exists && until.After(now) {
			if earliest.IsZero() || until.Before(earliest) {
				earliest = until
			}
		}
	}
	if earliest.IsZero() {
		return 0
	}
	return earliest.Sub(now)
}

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
		return nil, fmt.Errorf("no AI provider configured. Add one with 'pkt change ai add provider <provider>'")
	}

	// Resolve the provider config from the registry
	pc, exists := cfg.AIProviders[providerName]
	defaults, known := builtinDefaults[providerName]
	if !exists && !known {
		return nil, fmt.Errorf("unknown provider '%s'. Use openai, groq, gemini, ollama, local, or register a custom one with 'pkt change ai add provider %s'", providerName, providerName)
	}

	baseURL := pc.BaseURL
	if baseURL == "" && known {
		baseURL = defaults[0]
	}
	if baseURL == "" {
		return nil, fmt.Errorf("no base URL configured for provider '%s'", providerName)
	}

	// Determine model: pc.DefaultModel -> pc.Models[0] -> builtin default -> "default"
	model := pc.DefaultModel
	if model == "" && len(pc.Models) > 0 {
		model = pc.Models[0]
	}
	if model == "" && known {
		model = defaults[1]
	}
	if model == "" {
		model = "default"
	}

	isLocal := localProviders[providerName]
	keysToTry := pc.APIKeys
	if isLocal && len(keysToTry) == 0 {
		keysToTry = []string{""}
	} else if !isLocal && len(keysToTry) == 0 {
		return nil, fmt.Errorf("no API key set for '%s'. Add one with: pkt change ai add key <key> --provider %s", providerName, providerName)
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

	// Prioritize keys not currently in cooldown
	var orderedKeys []string
	var cooledDownKeys []string
	for _, k := range keysToTry {
		if k == "" || !isKeyCooledDown(k) {
			orderedKeys = append(orderedKeys, k)
		} else {
			cooledDownKeys = append(cooledDownKeys, k)
		}
	}
	orderedKeys = append(orderedKeys, cooledDownKeys...)

	var lastErrStr string
	client := &http.Client{}

	for keyIdx, apiKey := range orderedKeys {
		req, err := http.NewRequest("POST", baseURL, bytes.NewBuffer(jsonBody))
		if err != nil {
			return nil, err
		}

		req.Header.Set("Content-Type", "application/json")
		if apiKey != "" {
			req.Header.Set("Authorization", "Bearer "+apiKey)
		}

		resp, err := client.Do(req)
		if err != nil {
			return nil, err
		}

		bodyBytes, _ := io.ReadAll(resp.Body)
		_ = resp.Body.Close()

		if resp.StatusCode == http.StatusTooManyRequests || (resp.StatusCode != http.StatusOK && (strings.Contains(string(bodyBytes), "rate_limit") || strings.Contains(string(bodyBytes), "RESOURCE_EXHAUSTED") || strings.Contains(string(bodyBytes), "Rate limit"))) {
			retrySec := 30
			if ra := resp.Header.Get("Retry-After"); ra != "" {
				if sec, err := strconv.Atoi(ra); err == nil && sec > 0 {
					retrySec = sec
				}
			}
			if apiKey != "" {
				setKeyCooldown(apiKey, time.Duration(retrySec)*time.Second)
			}
			lastErrStr = string(bodyBytes)

			if keyIdx < len(orderedKeys)-1 {
				fmt.Printf("\033[33m⚠️ Key for '%s' rate limited (429). Failing over to next key...\033[0m\n", providerName)
				continue
			}
			earliest := getEarliestCooldown(keysToTry)
			return nil, fmt.Errorf("all %d API key(s) for '%s' are rate limited (429). Earliest retry in %s", len(keysToTry), providerName, earliest.Round(time.Second))
		}

		if resp.StatusCode != http.StatusOK {
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
		if err := json.Unmarshal(bodyBytes, &parsedResp); err != nil {
			return nil, err
		}

		if len(parsedResp.Choices) == 0 {
			return nil, fmt.Errorf("no response from AI")
		}

		return &parsedResp.Choices[0].Message, nil
	}

	if lastErrStr != "" {
		return nil, fmt.Errorf("API rate limit error: %s", lastErrStr)
	}
	return nil, fmt.Errorf("no available keys")
}
