package cmd

import (
	"fmt"
	"os"
	"strings"

	"github.com/genesix/pkt/internal/ai"
	"github.com/genesix/pkt/internal/db"
	"github.com/genesix/pkt/internal/pm"
	"github.com/genesix/pkt/internal/utils"
	"github.com/spf13/cobra"
)

var (
	devFlag    bool
	allFlag    bool
	aiFlag     bool
	aiProvider string
)

var addCmd = &cobra.Command{
	Use:   "add [packages...] | .",
	Short: "Add or install dependencies for the current project",
	Long: `Add one or more dependencies or install all dependencies using its package manager.
Must be run inside a tracked project folder.

Supports: JavaScript (npm/pnpm/bun), Python (uv/pip/poetry), Go, Rust

Examples:
  pkt add axios                    # JavaScript
  pkt add requests flask           # Python
  pkt add -D typescript eslint     # Dev dependencies
  pkt add .                        # Install all dependencies
  pkt add -a                       # Install all dependencies`,
	Args: func(cmd *cobra.Command, args []string) error {
		isAll, _ := cmd.Flags().GetBool("all")
		if !isAll && len(args) == 0 {
			return fmt.Errorf("requires at least 1 package to add, or use --all / . to install all dependencies")
		}
		return nil
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		packages := args
		isAll := allFlag
		if len(packages) > 0 && packages[0] == "." {
			isAll = true
			packages = packages[1:]
		}

		if isAll && len(packages) > 0 {
			return fmt.Errorf("cannot specify packages when using --all or .")
		}

		// Get current directory
		cwd, err := os.Getwd()
		if err != nil {
			return fmt.Errorf("failed to get current directory: %w", err)
		}

		// Get project from current directory
		project, err := db.GetProjectByPath(cwd)
		if err != nil {
			return fmt.Errorf("not in a tracked project. Run this command inside a project folder")
		}

		// Get package manager for this language
		packageManager, err := pm.Get(project.Language, project.PackageManager)
		if err != nil {
			return err
		}

		if aiFlag {
			if len(packages) == 0 {
				return fmt.Errorf("please provide a description of the package you want when using --ai")
			}
			desc := strings.Join(packages, " ")

			fmt.Println("🤖 Thinking...")
			sysPrompt := fmt.Sprintf("You are a package-manager assistant for a %s project. The user wants to add a dependency based on their description. Return ONLY the exact space-separated module names to install. No explanation, no markdown backticks, no code blocks.", project.Language)
			resp, err := ai.AskAI(sysPrompt, desc, aiProvider)
			if err != nil {
				return fmt.Errorf("ai error: %w", err)
			}

			resp = strings.TrimSpace(strings.ReplaceAll(resp, "`", ""))
			if resp == "" {
				return fmt.Errorf("AI could not determine packages to add")
			}

			packages = strings.Split(resp, " ")
			fmt.Printf("🤖 AI suggests: \033[36m%s\033[0m\n\n", strings.Join(packages, " "))
		}

		if isAll {
			fmt.Printf("📦 Installing dependencies using %s for %s project...\n", project.PackageManager, project.Language)
			if err := packageManager.Install(cwd); err != nil {
				return fmt.Errorf("failed to install dependencies: %w", err)
			}
		} else {
			// Add dependencies
			packagesStr := strings.Join(packages, " ")
			fmt.Printf("Adding %s...\n", packagesStr)

			if err := packageManager.Add(cwd, packages, devFlag); err != nil {
				return fmt.Errorf("failed to add dependencies: %w", err)
			}
		}

		// Sync dependencies to database for all languages
		deps, err := utils.ParseDependencies(cwd, project.Language)
		if err != nil {
			fmt.Printf("⚠️  Warning: failed to parse dependencies: %v\n", err)
		} else if len(deps) > 0 {
			if err := db.SyncDependencies(project.ID, deps); err != nil {
				fmt.Printf("⚠️  Warning: failed to sync dependencies: %v\n", err)
			} else if isAll {
				fmt.Printf("✓ Synced %d dependencies to database\n", len(deps))
			}
		}

		if isAll {
			fmt.Println("✓ Dependencies installed")
		} else {
			if len(packages) == 1 {
				fmt.Printf("✓ Added %s\n", packages[0])
			} else {
				packagesStr := strings.Join(packages, " ")
				fmt.Printf("✓ Added %d packages: %s\n", len(packages), packagesStr)
			}
		}

		return nil
	},
}

func init() {
	addCmd.Flags().BoolVarP(&devFlag, "dev", "D", false, "Add as dev dependency")
	addCmd.Flags().BoolVarP(&allFlag, "all", "a", false, "Install all dependencies")
	addCmd.Flags().BoolVarP(&aiFlag, "ai", "", false, "Use AI to determine and add dependencies based on description")
	addCmd.Flags().StringVarP(&aiProvider, "provider", "p", "", "Specific AI Provider to use with --ai (openai, gemini, groq)")
}
