package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/AlecAivazis/survey/v2"
	"github.com/genesix/pkt/internal/config"
	"github.com/genesix/pkt/internal/db"
	"github.com/genesix/pkt/internal/lang"
	"github.com/genesix/pkt/internal/pm"
	"github.com/genesix/pkt/internal/utils"
	"github.com/spf13/cobra"
)

var configCmd = &cobra.Command{
	Use:   "config",
	Short: "Show pkt configuration",
	Long: `Show the current pkt configuration and registered AI providers.

Use 'pkt change' to modify configuration settings.`,
	Args: cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := config.Load()
		if err != nil {
			return err
		}

		fmt.Println("Current configuration:")
		fmt.Printf("  projects_root: %s\n", cfg.ProjectsRoot)
		fmt.Printf("  editor:        %s\n", cfg.EditorCommand)
		fmt.Println("  default package managers:")
		fmt.Printf("    JavaScript: %s\n", cfg.DefaultPMForLanguage("javascript"))
		fmt.Printf("    Python:     %s\n", cfg.DefaultPMForLanguage("python"))
		fmt.Printf("    Go:         %s\n", cfg.DefaultPMForLanguage("go"))
		fmt.Printf("    Rust:       %s\n", cfg.DefaultPMForLanguage("rust"))
		fmt.Printf("  ai (active):   %s\n", cfg.AIProvider)
		if len(cfg.AIProviders) > 0 {
			fmt.Println("\nRegistered AI Providers:")

			var names []string
			for _, name := range cfg.ProviderOrder {
				if _, ok := cfg.AIProviders[name]; ok {
					names = append(names, name)
				}
			}
			for name := range cfg.AIProviders {
				found := false
				for _, n := range names {
					if n == name {
						found = true
						break
					}
				}
				if !found {
					names = append(names, name)
				}
			}

			for _, name := range names {
				pc := cfg.AIProviders[name]
				active := ""
				if name == cfg.AIProvider {
					active = " ← active"
				}

				keyStr := "(no key)"
				if len(pc.APIKeys) == 1 {
					keyStr = fmt.Sprintf("key=%s", config.MaskKey(pc.APIKeys[0]))
				} else if len(pc.APIKeys) > 1 {
					var masked []string
					for _, k := range pc.APIKeys {
						masked = append(masked, config.MaskKey(k))
					}
					keyStr = fmt.Sprintf("keys=[%s]", strings.Join(masked, ", "))
				}

				modelStr := ""
				if pc.DefaultModel != "" {
					modelStr = fmt.Sprintf("model=%s", pc.DefaultModel)
				} else if len(pc.Models) > 0 {
					modelStr = fmt.Sprintf("model=%s", pc.Models[0])
				}

				if pc.BaseURL != "" {
					fmt.Printf("  %-10s  %s  url=%s  %s%s\n", name, keyStr, pc.BaseURL, modelStr, active)
				} else {
					fmt.Printf("  %-10s  %s  %s%s\n", name, keyStr, modelStr, active)
				}
			}
		}
		fmt.Printf("\nConfig file: ~/.pkt/config.json\n")
		return nil
	},
}

func promptDefaultPMLanguage() (string, error) {
	options := []string{
		"JavaScript (js)",
		"Python (py)",
		"Go (go)",
		"Rust (rs)",
	}
	var selected string
	if err := survey.AskOne(&survey.Select{
		Message: "Which language default should change?",
		Options: options,
	}, &selected); err != nil {
		return "", fmt.Errorf("select language: %w", err)
	}
	for _, option := range options {
		if selected == option {
			start := strings.LastIndex(option, "(")
			return option[start+1 : len(option)-1], nil
		}
	}
	return "", fmt.Errorf("invalid language selection")
}

var changeCmd = &cobra.Command{
	Use:   "change <key> <value>",
	Short: "Change pkt configuration",
	Long: `Change a pkt configuration setting.

Available keys:
  editor    - Editor command (e.g., code, cursor, vim)
  pm        - Package manager for the current project or a language default
  lang      - Change language for a project (e.g., go, js, py, rs)
  rebuild   - Re-detect and rebuild language & package manager for a project
  root      - Projects root for newly created or cloned projects

To manage AI providers, models, and keys use the 'ai' subcommand:
  pkt change ai add provider groq
  pkt change ai add key gsk-... --provider groq
  pkt change ai add model llama-3.3-70b-versatile --provider groq
  pkt change ai groq                 # Switch active provider

For project-specific settings (like 'pm' and 'lang'), pkt updates the current project when
you are inside a tracked folder, or you can specify the target project.

For 'pm', use -g to change a language default while inside a project. Outside
a tracked project, pkt asks which language default to change when -l is omitted.

Examples:
  pkt change editor cursor
  pkt change pm npm
  pkt change pm pnpm -l js -g
  pkt change lang go
  pkt change lang js my-project
  pkt change rebuild
  pkt change rebuild my-project
  pkt change root ~/Projects`,
	Args:              cobra.ArbitraryArgs,
	ValidArgsFunction: completeConfigKeys,
	RunE: func(cmd *cobra.Command, args []string) error {
		validKeys := []string{"editor", "pm", "root", "lang", "language", "rebuild", "ai"}

		if len(args) == 0 {
			return cmd.Help()
		}

		// Check if the first arg looks like an unrecognised or removed subcommand
		key := args[0]
		isValidKey := false
		for _, k := range validKeys {
			if k == key {
				isValidKey = true
				break
			}
		}
		if !isValidKey {
			return fmt.Errorf("unknown key or command: %q\n\nValid keys: editor, pm, lang, rebuild, root\nTo manage AI providers use: pkt change ai add/remove", key)
		}

		if key != "rebuild" && len(args) < 2 {
			return fmt.Errorf("missing value for key %q\n\nUsage: pkt change %s <value>", key, key)
		}

		cfg, err := config.Load()
		if err != nil {
			return err
		}

		value := ""
		if len(args) > 1 {
			value = args[1]
		}
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
			languageFlag, _ := cmd.Flags().GetString("language")
			languageFlag = lang.NormalizeName(languageFlag)

			cwd, err := os.Getwd()
			if err != nil {
				return fmt.Errorf("failed to get current directory: %w", err)
			}

			project, projectErr := db.GetProjectByPath(cwd)
			if isGlobal || projectErr != nil {
				targetLanguage := languageFlag
				if targetLanguage == "" && projectErr == nil {
					targetLanguage = project.Language
				}
				if targetLanguage == "" {
					targetLanguage, err = promptDefaultPMLanguage()
					if err != nil {
						return err
					}
				}
				if _, err := lang.Get(targetLanguage); err != nil {
					return err
				}
				if pManager.Language() != targetLanguage {
					return fmt.Errorf("%s is not a package manager for %s", value, targetLanguage)
				}
				cfg.SetDefaultPMForLanguage(targetLanguage, pManager.Name())
				if err := config.Save(cfg); err != nil {
					return fmt.Errorf("failed to save config: %w", err)
				}
				fmt.Printf("✓ Default package manager for %s set to: %s\n", targetLanguage, value)
				return nil
			}

			if languageFlag != "" && languageFlag != project.Language {
				return fmt.Errorf("--language %s does not match the current %s project; use --global to change a language default", languageFlag, project.Language)
			}
			if pManager.Language() != project.Language {
				return fmt.Errorf("%s is not a package manager for this %s project", value, project.Language)
			}

			if err := db.UpdateProjectPM(project.ID, pManager.Name()); err != nil {
				return fmt.Errorf("failed to update project package manager: %w", err)
			}

			fmt.Printf("✓ Updated package manager for project %s: %s → %s\n", project.Name, project.PackageManager, value)

		case "root":
			projectsRoot, err := utils.ExpandPath(value)
			if err != nil {
				return fmt.Errorf("failed to resolve projects root: %w", err)
			}
			if projectsRoot == "" {
				return fmt.Errorf("projects root cannot be empty")
			}
			if projectsRoot == string(os.PathSeparator) {
				return fmt.Errorf("refusing to use the filesystem root as the projects root")
			}
			if err := os.MkdirAll(projectsRoot, 0755); err != nil {
				return fmt.Errorf("failed to create projects root: %w", err)
			}
			info, err := os.Stat(projectsRoot)
			if err != nil {
				return fmt.Errorf("failed to inspect projects root: %w", err)
			}
			if !info.IsDir() {
				return fmt.Errorf("projects root must be a directory: %s", projectsRoot)
			}

			cfg.ProjectsRoot = projectsRoot
			if err := config.Save(cfg); err != nil {
				return fmt.Errorf("failed to save config: %w", err)
			}
			fmt.Printf("✓ Projects root set to: %s\n", projectsRoot)
			fmt.Println("  Existing tracked projects were not moved.")

		case "lang", "language":
			newLangObj, err := lang.Get(value)
			if err != nil {
				return fmt.Errorf("unsupported language %q. Valid options: %s", value, strings.Join(lang.ListShort(), ", "))
			}

			var project *db.Project
			if len(args) >= 3 {
				project, err = utils.ResolveProject(args[2])
				if err != nil {
					return err
				}
			} else {
				cwd, err := os.Getwd()
				if err != nil {
					return fmt.Errorf("failed to get current directory: %w", err)
				}
				project, err = db.GetProjectByPath(cwd)
				if err != nil {
					return fmt.Errorf("not inside a tracked project. Specify project: pkt change lang %s <project>", value)
				}
			}

			newPM := newLangObj.DetectPackageManager(project.Path)
			if newPM == "" {
				newPM = newLangObj.DefaultPackageManager()
			}

			if err := db.UpdateProjectLang(project.ID, newLangObj.Name(), newPM); err != nil {
				return fmt.Errorf("failed to update project language: %w", err)
			}

			fmt.Printf("✓ Updated project %s language: %s → %s\n", project.Name, project.Language, newLangObj.Name())
			fmt.Printf("  Package Manager: %s → %s\n", project.PackageManager, newPM)

		case "rebuild":
			var project *db.Project
			if len(args) >= 2 {
				project, err = utils.ResolveProject(args[1])
				if err != nil {
					return err
				}
			} else {
				cwd, err := os.Getwd()
				if err != nil {
					return fmt.Errorf("failed to get current directory: %w", err)
				}
				project, err = db.GetProjectByPath(cwd)
				if err != nil {
					return fmt.Errorf("not inside a tracked project. Specify project: pkt change rebuild <project>")
				}
			}

			detectedLang, err := lang.Detect(project.Path)
			if err != nil {
				return fmt.Errorf("failed to detect project language in %s: %w", project.Path, err)
			}
			detectedPM := detectedLang.DetectPackageManager(project.Path)
			if detectedPM == "" {
				detectedPM = detectedLang.DefaultPackageManager()
			}

			if err := db.UpdateProjectLang(project.ID, detectedLang.Name(), detectedPM); err != nil {
				return fmt.Errorf("failed to update project configuration: %w", err)
			}

			fmt.Printf("✓ Rebuilt configuration for %s:\n", project.Name)
			fmt.Printf("  Language:        %s → %s\n", project.Language, detectedLang.Name())
			fmt.Printf("  Package Manager: %s → %s\n", project.PackageManager, detectedPM)

		case "ai":
			if err := cfg.SetActiveProvider(value); err != nil {
				return err
			}
			fmt.Printf("✓ Active AI provider set to: %s\n", value)

		default:
			return fmt.Errorf("unknown config key: %s\nAvailable keys: editor, pm, lang, rebuild, ai, root", key)
		}

		return nil
	},
}

var aiCmd = &cobra.Command{
	Use:   "ai [provider]",
	Short: "Manage AI providers, API keys, and models",
	Long: `Manage AI providers, API keys, and models.

Examples:
  pkt change ai groq
  pkt change ai add provider groq
  pkt change ai add provider local --url http://localhost:1234/v1
  pkt change ai add key gsk-... --provider groq
  pkt change ai add model llama-3.3-70b-versatile --provider groq
  pkt change ai remove key gsk-... --provider groq
  pkt change ai remove model llama-3.3-70b-versatile --provider groq
  pkt change ai remove provider groq`,
	Args: cobra.MaximumNArgs(1),
	ValidArgsFunction: func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		if len(args) == 0 {
			return completionFilter(knownAIProviders(), toComplete), cobra.ShellCompDirectiveNoFileComp
		}
		return nil, cobra.ShellCompDirectiveNoFileComp
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := config.Load()
		if err != nil {
			return err
		}

		if len(args) == 1 {
			if err := cfg.SetActiveProvider(args[0]); err != nil {
				return err
			}
			fmt.Printf("✓ Active AI provider set to: %s\n", args[0])
			return nil
		}

		fmt.Printf("Active AI provider: %s\n", cfg.AIProvider)
		if len(cfg.AIProviders) > 0 {
			fmt.Println("\nRegistered Providers:")
			for _, name := range cfg.ProviderOrder {
				active := ""
				if name == cfg.AIProvider {
					active = " ← active"
				}
				fmt.Printf("  • %s%s\n", name, active)
			}
		}
		return nil
	},
}

var aiAddCmd = &cobra.Command{
	Use:   "add",
	Short: "Add an AI provider, API key, or model",
}

var aiAddProviderCmd = &cobra.Command{
	Use:   "provider <name>",
	Short: "Add or update an AI provider",
	Args:  cobra.ExactArgs(1),
	ValidArgsFunction: func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		if len(args) == 0 {
			return completionFilter(knownAIProviders(), toComplete), cobra.ShellCompDirectiveNoFileComp
		}
		return nil, cobra.ShellCompDirectiveNoFileComp
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := config.Load()
		if err != nil {
			return err
		}
		url, _ := cmd.Flags().GetString("url")
		name := args[0]
		isNew, err := cfg.AddProvider(name, url)
		if err != nil {
			return err
		}
		if isNew {
			fmt.Printf("✓ Added AI provider '%s'\n", name)
		} else {
			fmt.Printf("✓ Updated AI provider '%s'\n", name)
		}
		return nil
	},
}

var aiAddKeyCmd = &cobra.Command{
	Use:   "key <api_key>",
	Short: "Add an API key to an existing provider",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		provider, _ := cmd.Flags().GetString("provider")
		provider = strings.TrimSpace(provider)
		if provider == "" {
			return fmt.Errorf("--provider is required (e.g. --provider groq)")
		}

		cfg, err := config.Load()
		if err != nil {
			return err
		}

		key := args[0]
		if err := cfg.AddKey(provider, key); err != nil {
			return err
		}
		fmt.Printf("✓ Added API key (%s) to provider '%s'\n", config.MaskKey(key), provider)
		return nil
	},
}

var aiAddModelCmd = &cobra.Command{
	Use:   "model <model_name>",
	Short: "Add a model to an existing provider",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		provider, _ := cmd.Flags().GetString("provider")
		provider = strings.TrimSpace(provider)
		if provider == "" {
			return fmt.Errorf("--provider is required (e.g. --provider groq)")
		}

		cfg, err := config.Load()
		if err != nil {
			return err
		}

		model := args[0]
		if err := cfg.AddModel(provider, model); err != nil {
			return err
		}
		fmt.Printf("✓ Added model '%s' to provider '%s'\n", model, provider)
		return nil
	},
}

var aiRemoveCmd = &cobra.Command{
	Use:   "remove",
	Short: "Remove an AI provider, API key, or model",
}

var aiRemoveProviderCmd = &cobra.Command{
	Use:   "provider <name>",
	Short: "Remove an AI provider",
	Args:  cobra.ExactArgs(1),
	ValidArgsFunction: func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		if len(args) == 0 {
			return completionFilter(knownAIProviders(), toComplete), cobra.ShellCompDirectiveNoFileComp
		}
		return nil, cobra.ShellCompDirectiveNoFileComp
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := config.Load()
		if err != nil {
			return err
		}
		name := args[0]
		if err := cfg.RemoveProvider(name); err != nil {
			return err
		}
		fmt.Printf("✓ Removed AI provider '%s'\n", name)
		return nil
	},
}

var aiRemoveKeyCmd = &cobra.Command{
	Use:   "key <api_key>",
	Short: "Remove an API key from an existing provider",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		provider, _ := cmd.Flags().GetString("provider")
		provider = strings.TrimSpace(provider)
		if provider == "" {
			return fmt.Errorf("--provider is required (e.g. --provider groq)")
		}

		cfg, err := config.Load()
		if err != nil {
			return err
		}

		key := args[0]
		if err := cfg.RemoveKey(provider, key); err != nil {
			return err
		}
		fmt.Printf("✓ Removed API key from provider '%s'\n", provider)
		return nil
	},
}

var aiRemoveModelCmd = &cobra.Command{
	Use:   "model <model_name>",
	Short: "Remove a model from an existing provider",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		provider, _ := cmd.Flags().GetString("provider")
		provider = strings.TrimSpace(provider)
		if provider == "" {
			return fmt.Errorf("--provider is required (e.g. --provider groq)")
		}

		cfg, err := config.Load()
		if err != nil {
			return err
		}

		model := args[0]
		if err := cfg.RemoveModel(provider, model); err != nil {
			return err
		}
		fmt.Printf("✓ Removed model '%s' from provider '%s'\n", model, provider)
		return nil
	},
}

var langCmd = &cobra.Command{
	Use:     "lang <language> [project]",
	Aliases: []string{"language"},
	Short:   "Change language and package manager for a project",
	Long: `Change the language and default package manager for a tracked project.
pkt will automatically update the language and select an appropriate package manager.

Examples:
  pkt change lang go
  pkt change lang js my-project
  pkt change lang py .`,
	Args: cobra.RangeArgs(1, 2),
	ValidArgsFunction: func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		if len(args) == 0 {
			allLangs := append(lang.ListShort(), lang.List()...)
			return completionFilter(allLangs, toComplete), cobra.ShellCompDirectiveNoFileComp
		}
		if len(args) == 1 {
			return completeProjects(cmd, args, toComplete)
		}
		return nil, cobra.ShellCompDirectiveNoFileComp
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		value := args[0]
		newLangObj, err := lang.Get(value)
		if err != nil {
			return fmt.Errorf("unsupported language %q. Valid options: %s", value, strings.Join(lang.ListShort(), ", "))
		}

		var project *db.Project
		if len(args) >= 2 {
			project, err = utils.ResolveProject(args[1])
			if err != nil {
				return err
			}
		} else {
			cwd, err := os.Getwd()
			if err != nil {
				return fmt.Errorf("failed to get current directory: %w", err)
			}
			project, err = db.GetProjectByPath(cwd)
			if err != nil {
				return fmt.Errorf("not inside a tracked project. Specify project: pkt change lang %s <project>", value)
			}
		}

		newPM := newLangObj.DetectPackageManager(project.Path)
		if newPM == "" {
			newPM = newLangObj.DefaultPackageManager()
		}

		if err := db.UpdateProjectLang(project.ID, newLangObj.Name(), newPM); err != nil {
			return fmt.Errorf("failed to update project language: %w", err)
		}

		fmt.Printf("✓ Updated project %s language: %s → %s\n", project.Name, project.Language, newLangObj.Name())
		fmt.Printf("  Package Manager: %s → %s\n", project.PackageManager, newPM)
		return nil
	},
}

var rebuildCmd = &cobra.Command{
	Use:   "rebuild [project]",
	Short: "Re-detect and rebuild language & package manager for a project",
	Long: `Re-scan the project directory, automatically detect the primary language and
package manager, and update the project configuration in the pkt database.

Examples:
  pkt change rebuild
  pkt change rebuild my-project
  pkt change rebuild .`,
	Args: cobra.MaximumNArgs(1),
	ValidArgsFunction: func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		if len(args) == 0 {
			return completeProjectOrDot(cmd, args, toComplete)
		}
		return nil, cobra.ShellCompDirectiveNoFileComp
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		var project *db.Project
		var err error
		if len(args) >= 1 {
			project, err = utils.ResolveProject(args[0])
			if err != nil {
				return err
			}
		} else {
			cwd, err := os.Getwd()
			if err != nil {
				return fmt.Errorf("failed to get current directory: %w", err)
			}
			project, err = db.GetProjectByPath(cwd)
			if err != nil {
				return fmt.Errorf("not inside a tracked project. Specify project: pkt change rebuild <project>")
			}
		}

		detectedLang, err := lang.Detect(project.Path)
		if err != nil {
			return fmt.Errorf("failed to detect project language in %s: %w", project.Path, err)
		}
		detectedPM := detectedLang.DetectPackageManager(project.Path)
		if detectedPM == "" {
			detectedPM = detectedLang.DefaultPackageManager()
		}

		if err := db.UpdateProjectLang(project.ID, detectedLang.Name(), detectedPM); err != nil {
			return fmt.Errorf("failed to update project configuration: %w", err)
		}

		fmt.Printf("✓ Rebuilt configuration for %s:\n", project.Name)
		fmt.Printf("  Language:        %s → %s\n", project.Language, detectedLang.Name())
		fmt.Printf("  Package Manager: %s → %s\n", project.PackageManager, detectedPM)
		return nil
	},
}

var rebuildRootCmd = &cobra.Command{
	Use:               "rebuild [project]",
	Short:             "Re-detect and rebuild language & package manager for a project",
	Long:              rebuildCmd.Long,
	Args:              rebuildCmd.Args,
	ValidArgsFunction: rebuildCmd.ValidArgsFunction,
	RunE:              rebuildCmd.RunE,
}

func init() {
	rootCmd.AddCommand(configCmd)
	rootCmd.AddCommand(changeCmd)
	rootCmd.AddCommand(rebuildRootCmd)
	changeCmd.Flags().BoolP("global", "g", false, "Change the default package manager instead of the current project")
	changeCmd.Flags().StringP("language", "l", "", "Language for a default package manager (js, py, go, rs)")
	_ = changeCmd.RegisterFlagCompletionFunc("language", completeLangFlag)

	changeCmd.AddCommand(langCmd)
	changeCmd.AddCommand(rebuildCmd)
	changeCmd.AddCommand(aiCmd)
	aiCmd.AddCommand(aiAddCmd)
	aiCmd.AddCommand(aiRemoveCmd)

	aiAddCmd.AddCommand(aiAddProviderCmd)
	aiAddCmd.AddCommand(aiAddKeyCmd)
	aiAddCmd.AddCommand(aiAddModelCmd)

	aiRemoveCmd.AddCommand(aiRemoveProviderCmd)
	aiRemoveCmd.AddCommand(aiRemoveKeyCmd)
	aiRemoveCmd.AddCommand(aiRemoveModelCmd)

	aiAddProviderCmd.Flags().String("url", "", "Custom base URL for local/self-hosted providers")

	aiAddKeyCmd.Flags().StringP("provider", "p", "", "Target AI provider")
	_ = aiAddKeyCmd.RegisterFlagCompletionFunc("provider", completeAIProviderArg)

	aiAddModelCmd.Flags().StringP("provider", "p", "", "Target AI provider")
	_ = aiAddModelCmd.RegisterFlagCompletionFunc("provider", completeAIProviderArg)

	aiRemoveKeyCmd.Flags().StringP("provider", "p", "", "Target AI provider")
	_ = aiRemoveKeyCmd.RegisterFlagCompletionFunc("provider", completeAIProviderArg)

	aiRemoveModelCmd.Flags().StringP("provider", "p", "", "Target AI provider")
	_ = aiRemoveModelCmd.RegisterFlagCompletionFunc("provider", completeAIProviderArg)
}
