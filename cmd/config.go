package cmd

import (
	"fmt"
	"os"
	"os/exec"

	"github.com/genesix/pkt/internal/config"
	"github.com/genesix/pkt/internal/db"
	"github.com/genesix/pkt/internal/pm"
	"github.com/genesix/pkt/internal/utils"
	"github.com/spf13/cobra"
)

var configCmd = &cobra.Command{
	Use:   "config [key] [value]",
	Short: "View or update pkt configuration",
	Long: `View or update pkt configuration settings.

Without arguments, displays current configuration.

This command unifies both global settings and project-specific settings.

Available keys:
  editor    - Editor command (e.g., code, cursor, vim)
  pm        - Package manager (pnpm, npm, bun, uv, pip, poetry, go, cargo)
  ai        - Switch active AI provider

For project-specific settings (like 'pm'), pkt will automatically update the current project 
if you are inside a tracked folder. Use --global (-g) to force updating the global default instead.

Examples:
  pkt config                    # Show current config
  pkt config editor cursor      # Change global editor to cursor
  pkt config pm npm             # Change current project's package manager to npm
  pkt config pm pnpm -g         # Change global default package manager to pnpm
  pkt config ai ollama          # Switch global AI provider to ollama`,
	Args:              cobra.MaximumNArgs(2),
	ValidArgsFunction: completeConfigKeys,
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
			pManager, err := pm.GetPM(value)
			if err != nil {
				return fmt.Errorf("invalid or unsupported package manager: %s", value)
			}

			isGlobal, _ := cmd.Flags().GetBool("global")

			if isGlobal {
				if _, err := exec.LookPath(value); err != nil {
					fmt.Printf("⚠️  Warning: '%s' is not installed\n", value)
				}
				cfg.DefaultPM = value
				if err := config.Save(cfg); err != nil {
					return fmt.Errorf("failed to save config: %w", err)
				}
				fmt.Printf("✓ Global default package manager set to: %s\n", value)
				return nil
			}

			// Try to set for current project
			cwd, err := os.Getwd()
			if err != nil {
				return err
			}
			project, err := db.GetProjectByPath(cwd)
			if err != nil {
				// Not in a project, default to global implicitly
				fmt.Printf("ℹ️  Not in a tracked project. Setting global default instead.\n")
				if _, err := exec.LookPath(value); err != nil {
					fmt.Printf("⚠️  Warning: '%s' is not installed\n", value)
				}
				cfg.DefaultPM = value
				if err := config.Save(cfg); err != nil {
					return fmt.Errorf("failed to save config: %w", err)
				}
				fmt.Printf("✓ Global default package manager set to: %s\n", value)
				return nil
			}

			// In a tracked project! Update the project's PM
			if project.PackageManager == value {
				fmt.Printf("Project %s is already using %s\n", project.Name, value)
				return nil
			}

			if pManager.Language() != project.Language {
				return fmt.Errorf("cannot switch project language from %s to package manager %s", project.Language, value)
			}

			if err := db.UpdateProjectPM(project.ID, value); err != nil {
				return fmt.Errorf("failed to update package manager: %w", err)
			}

			// If JavaScript, rewrite scripts
			if project.Language == "javascript" {
				if err := utils.RewriteScripts(project.Path, value); err != nil {
					fmt.Printf("Warning: failed to rewrite scripts: %v\n", err)
				}
			}

			// Re-sync dependencies
			deps, err := utils.ParseDependencies(project.Path, project.Language)
			if err != nil {
				fmt.Printf("Warning: failed to sync dependencies: %v\n", err)
			} else {
				if err := db.SyncDependencies(project.ID, deps); err != nil {
					fmt.Printf("Warning: failed to sync dependencies: %v\n", err)
				}
			}

			fmt.Printf("✓ Updated package manager for project %s: %s → %s\n", project.Name, project.PackageManager, value)

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
	Args:              cobra.RangeArgs(1, 2),
	ValidArgsFunction: completeAIProviderArg,
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
	ValidArgsFunction: func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		if len(args) == 0 {
			return completeAIProviderArg(cmd, args, toComplete)
		}
		// Model names are free-form — no static completion
		return nil, cobra.ShellCompDirectiveNoFileComp
	},
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
	configCmd.Flags().BoolP("global", "g", false, "Force setting global configuration")
	configCmd.AddCommand(configSetAICmd)
	configCmd.AddCommand(configSetModelCmd)
	configSetAICmd.Flags().String("url", "", "Custom base URL for local/self-hosted providers")
}
