package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/AlecAivazis/survey/v2"
	"github.com/genesix/pkt/internal/config"
	"github.com/genesix/pkt/internal/db"
	"github.com/genesix/pkt/internal/pm"
	"github.com/spf13/cobra"
)

var startCmd = &cobra.Command{
	Use:   "start",
	Short: "Initialize pkt configuration",
	Long: `Initialize pkt by setting up configuration and database.
This must be run before using any other pkt commands.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		fmt.Println("🚀 Starting pkt initialization...")
		fmt.Println()

		// Check if already initialized
		exists, err := config.Exists()
		if err != nil {
			return err
		}

		if exists {
			cfg, err := config.Load()
			if err == nil && cfg.Initialized {
				fmt.Println("✅ pkt is already initialized.")
				fmt.Println("\nConfiguration:")
				fmt.Printf("  Projects root: %s\n", cfg.ProjectsRoot)
				fmt.Printf("  Default PM: %s\n", cfg.DefaultPM)
				fmt.Printf("  Editor: %s\n", cfg.EditorCommand)
				fmt.Println("\nConfig file: ~/.pkt/config.json")
				fmt.Println("Database: ~/.pkt/pkt.db")
				fmt.Println("\nTo reconfigure, delete ~/.pkt/ and run 'pkt start' again.")
				return nil
			}
		}

		// Step 1: Check for npm
		fmt.Println("📦 Checking required tools...")
		npmAvailable := checkTool("npm")
		if !npmAvailable {
			return fmt.Errorf("npm is required to install pnpm - please install npm first from https://nodejs.org and then run 'pkt start' again")
		}
		fmt.Println("  ✅ npm detected")

		// Step 2: Check for pnpm and offer to install
		pnpmAvailable := checkTool("pnpm")
		if !pnpmAvailable {
			fmt.Println("  ⚠️  pnpm not detected")

			// Confirm installation
			var shouldInstall bool
			prompt := &survey.Confirm{
				Message: "pnpm is the recommended package manager. Install it now via npm?",
				Default: true,
			}
			if err := survey.AskOne(prompt, &shouldInstall); err != nil {
				return fmt.Errorf("cancelled: %w", err)
			}

			if shouldInstall {
				fmt.Println("\n  📦 Installing pnpm globally via npm...")
				if err := pm.InstallPnpm(); err != nil {
					fmt.Printf("  ⚠️  Failed to install pnpm: %v\n", err)
					fmt.Println("  You can install it manually: npm install -g pnpm")
				} else {
					fmt.Println("  ✅ pnpm installed successfully")
					pnpmAvailable = true
				}
			}
		} else {
			fmt.Println("  ✅ pnpm detected")
		}

		// Prompt for configuration
		var projectsRoot string
		var defaultPM string
		var editorCmd string

		// Get home directory for default
		home, err := os.UserHomeDir()
		if err != nil {
			return fmt.Errorf("failed to get home directory: %w", err)
		}
		defaultRoot := filepath.Join(home, "Documents", "workspace")

		// Step 3: Get configuration from user
		fmt.Println("\n⚙️  Configuration setup...")

		// Projects root
		promptRoot := &survey.Input{
			Message: "Projects root folder:",
			Default: defaultRoot,
		}
		if err := survey.AskOne(promptRoot, &projectsRoot); err != nil {
			return fmt.Errorf("cancelled: %w", err)
		}

		// Expand ~ if present
		if len(projectsRoot) >= 2 && projectsRoot[:2] == "~/" {
			projectsRoot = filepath.Join(home, projectsRoot[2:])
		}

		// Create projects root if it doesn't exist
		if err := os.MkdirAll(projectsRoot, 0755); err != nil {
			return fmt.Errorf("failed to create projects root: %w", err)
		}
		fmt.Printf("  ✅ Projects folder: %s\n", projectsRoot)

		// Default package manager
		available := pm.ListAvailable()
		if len(available) == 0 {
			return fmt.Errorf("no package managers found. Please install pnpm, npm, or bun")
		}

		defaultPMOption := "pnpm"
		if !pnpmAvailable && len(available) > 0 {
			defaultPMOption = available[0]
		}

		promptPM := &survey.Select{
			Message: "Default package manager:",
			Options: available,
			Default: defaultPMOption,
		}
		if err := survey.AskOne(promptPM, &defaultPM); err != nil {
			return fmt.Errorf("cancelled: %w", err)
		}
		fmt.Printf("  ✅ Default package manager: %s\n", defaultPM)

		// Editor command
		promptEditor := &survey.Input{
			Message: "Editor command (e.g., code, vim, nano):",
			Default: "code",
		}
		if err := survey.AskOne(promptEditor, &editorCmd); err != nil {
			return fmt.Errorf("cancelled: %w", err)
		}

		// Verify editor command exists
		editorAvailable := checkTool(editorCmd)
		if !editorAvailable {
			fmt.Printf("  ⚠️  Warning: '%s' command not found. You can update this later with 'pkt config editor <cmd>'\n", editorCmd)
		} else {
			fmt.Printf("  ✅ Editor command: %s\n", editorCmd)
		}

		// Step 4: Save configuration
		fmt.Println("\n💾 Saving configuration...")

		cfg := &config.Config{
			ProjectsRoot:  projectsRoot,
			DefaultPM:     defaultPM,
			EditorCommand: editorCmd,
			Initialized:   true,
		}

		if err := config.Save(cfg); err != nil {
			return fmt.Errorf("failed to save config: %w", err)
		}
		fmt.Println("  ✅ Configuration saved to ~/.pkt/config.json")

		// Step 5: Initialize database
		fmt.Println("\n🗄️  Initializing database...")
		if err := db.InitDB(); err != nil {
			return fmt.Errorf("failed to initialize database: %w", err)
		}
		fmt.Println("  ✅ Database initialized at ~/.pkt/pkt.db")

		// Final summary
		fmt.Println("\n" + strings.Repeat("=", 60))
		fmt.Println("🎉 pkt is ready to use!")
		fmt.Println(strings.Repeat("=", 60))
		fmt.Println("\n✅ Summary:")
		fmt.Println("  ✅ npm detected")
		if pnpmAvailable {
			fmt.Println("  ✅ pnpm detected")
		}
		fmt.Println("  ✅ SQLite database initialized")
		fmt.Printf("  ✅ Projects folder: %s\n", projectsRoot)
		fmt.Printf("  ✅ Default package manager: %s\n", defaultPM)
		fmt.Printf("  ✅ Editor command: %s\n", editorCmd)

		fmt.Println("\n📚 Next steps:")
		fmt.Println("  • Create a project: pkt create <project-name>")
		fmt.Println("  • Initialize existing: pkt init <path>")
		fmt.Println("  • List projects: pkt list")

		return nil
	},
}

// checkTool checks if a command/tool is available in PATH
func checkTool(name string) bool {
	_, err := exec.LookPath(name)
	return err == nil
}
