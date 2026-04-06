package cmd

import (
	"fmt"
	"os/exec"

	"github.com/genesix/pkt/internal/config"
	"github.com/spf13/cobra"
)

var configCmd = &cobra.Command{
	Use:   "config [key] [value]",
	Short: "View or update pkt configuration",
	Long: `View or update pkt configuration settings.

Without arguments, displays current configuration.

Available keys:
  editor    - Editor command (e.g., code, cursor, vim)
  pm        - Default package manager (pnpm, npm, bun)
  ai        - Switch active AI provider

Examples:
  pkt config                    # Show current config
  pkt config editor cursor      # Change editor to cursor
  pkt config pm npm             # Change default PM to npm
  pkt config ai ollama          # Switch to Ollama (local)`,
	Args: cobra.MaximumNArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := config.Load()
		if err != nil {
			return err
		}

		// No args: show current config
		if len(args) == 0 {
			fmt.Println("Current configuration:")
			fmt.Printf("  projects_root: %s\n", cfg.ProjectsRoot)
			fmt.Printf("  editor:        %s\n", cfg.EditorCommand)
			fmt.Printf("  pm:            %s\n", cfg.DefaultPM)
			fmt.Printf("  ai (active):   %s\n", cfg.AIProvider)
			if len(cfg.AIProviders) > 0 {
				fmt.Println("\nRegistered AI Providers:")
				for name, pc := range cfg.AIProviders {
					active := ""
					if name == cfg.AIProvider {
						active = " ← active"
					}
					if pc.APIKey != "" {
						fmt.Printf("  %-10s  key=****  model=%s%s\n", name, pc.Model, active)
					} else if pc.BaseURL != "" {
						fmt.Printf("  %-10s  url=%s  model=%s%s\n", name, pc.BaseURL, pc.Model, active)
					} else {
						fmt.Printf("  %-10s  (built-in local)  model=%s%s\n", name, pc.Model, active)
					}
				}
			}
			fmt.Printf("\nConfig file: ~/.pkt/config.json\n")
			return nil
		}

		if len(args) == 1 {
			switch args[0] {
			case "editor":
				fmt.Printf("editor: %s\n", cfg.EditorCommand)
			case "pm":
				fmt.Printf("pm: %s\n", cfg.DefaultPM)
			case "ai":
				fmt.Printf("ai: %s\n", cfg.AIProvider)
			default:
				return fmt.Errorf("unknown config key: %s", args[0])
			}
			return nil
		}

		key, value := args[0], args[1]
		switch key {
		case "editor":
			if _, err := exec.LookPath(value); err != nil {
				fmt.Printf("⚠️  Warning: '%s' command not found in PATH\n", value)
			}
			cfg.EditorCommand = value
			if err := config.Save(cfg); err != nil {
				return fmt.Errorf("failed to save config: %w", err)
			}
			fmt.Printf("✓ Editor set to: %s\n", value)

		case "pm":
			validPMs := map[string]bool{"pnpm": true, "npm": true, "bun": true}
			if !validPMs[value] {
				return fmt.Errorf("invalid package manager: %s\nSupported: pnpm, npm, bun", value)
			}
			if _, err := exec.LookPath(value); err != nil {
				fmt.Printf("⚠️  Warning: '%s' is not installed\n", value)
			}
			cfg.DefaultPM = value
			if err := config.Save(cfg); err != nil {
				return fmt.Errorf("failed to save config: %w", err)
			}
			fmt.Printf("✓ Default package manager set to: %s\n", value)

		case "projects_root":
			return fmt.Errorf("projects_root cannot be changed after setup\nRecreate config with 'pkt start' if needed")

		case "ai":
			cfg.AIProvider = value
			// Register built-in local providers on first switch even if no explicit set-ai was run
			if _, exists := cfg.AIProviders[value]; !exists {
				cfg.AIProviders[value] = config.ProviderConfig{}
			}
			if err := config.Save(cfg); err != nil {
				return fmt.Errorf("failed to save config: %w", err)
			}
			fmt.Printf("✓ Active AI provider set to: %s\n", value)

		default:
			return fmt.Errorf("unknown config key: %s\nAvailable keys: editor, pm, ai", key)
		}

		return nil
	},
}

// configSetAICmd registers a provider with an API key (cloud) or a custom URL (local).
var configSetAICmd = &cobra.Command{
	Use:   "set-ai <provider> [api_key]",
	Short: "Register an AI provider with an API key or custom URL",
	Long: `Register a cloud or local AI provider.

Cloud providers (require API key):
  pkt config set-ai groq   sk-...
  pkt config set-ai openai sk-...
  pkt config set-ai gemini AI...

Local providers (no key needed, use --url to override default):
  pkt config set-ai ollama                        # uses http://localhost:11434
  pkt config set-ai ollama --url http://myserver:11434
  pkt config set-ai local  --url http://localhost:1234`,
	Args: cobra.RangeArgs(1, 2),
	RunE: func(cmd *cobra.Command, args []string) error {
		provider := args[0]
		apiKey := ""
		if len(args) == 2 {
			apiKey = args[1]
		}

		customURL, _ := cmd.Flags().GetString("url")

		cfg, err := config.Load()
		if err != nil {
			return err
		}

		pc := cfg.AIProviders[provider]
		if apiKey != "" {
			pc.APIKey = apiKey
		}
		if customURL != "" {
			pc.BaseURL = customURL
		}
		cfg.AIProviders[provider] = pc

		if cfg.AIProvider == "" {
			cfg.AIProvider = provider
		}

		if err := config.Save(cfg); err != nil {
			return err
		}

		if apiKey != "" {
			fmt.Printf("✓ Registered provider '%s' with API key\n", provider)
		} else if customURL != "" {
			fmt.Printf("✓ Registered provider '%s' with URL: %s\n", provider, customURL)
		} else {
			fmt.Printf("✓ Registered provider '%s' (using built-in default URL)\n", provider)
		}
		return nil
	},
}

// configSetModelCmd pins a model for a specific provider.
var configSetModelCmd = &cobra.Command{
	Use:   "set-model <provider> <model_name>",
	Short: "Pin a specific AI model for a provider",
	Long: `Pin a model for any registered provider.

Examples:
  pkt config set-model groq   llama-3.3-70b-versatile
  pkt config set-model gemini gemini-1.5-pro
  pkt config set-model ollama mistral
  pkt config set-model local  phi3`,
	Args: cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		provider, model := args[0], args[1]

		cfg, err := config.Load()
		if err != nil {
			return err
		}

		pc := cfg.AIProviders[provider]
		pc.Model = model
		cfg.AIProviders[provider] = pc

		if err := config.Save(cfg); err != nil {
			return err
		}

		fmt.Printf("✓ Model for '%s' set to: %s\n", provider, model)
		return nil
	},
}

func init() {
	rootCmd.AddCommand(configCmd)
	configCmd.AddCommand(configSetAICmd)
	configCmd.AddCommand(configSetModelCmd)
	configSetAICmd.Flags().String("url", "", "Custom base URL for local/self-hosted providers")
}
