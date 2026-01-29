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
				fmt.Println("\nTo reconfigure, delete ~/.pkt/config.json and run 'pkt start' again.")
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
		
		// Set defaults for initial DB check
		projectsRoot = defaultRoot
		defaultPM = "pnpm"
		editorCmd = "code"

		// Step 3: Check PostgreSQL
		fmt.Println("\n🗄️  Checking PostgreSQL...")
		
		// Setup default config for connection test
		// Use environment variables if set, otherwise use defaults
		cfg := &config.Config{
			ProjectsRoot:  projectsRoot,
			DefaultPM:     defaultPM,
			EditorCommand: editorCmd,
			// Default DB settings
			DBUser:     getEnv("PKT_DB_USER", "pkt_user"),
			DBPassword: getEnv("PKT_DB_PASSWORD", "yourpassword"),
			DBName:     getEnv("PKT_DB_NAME", "pkt_db"),
			DBHost:     getEnv("PKT_DB_HOST", "127.0.0.1"),
			DBPort:     getEnv("PKT_DB_PORT", "5432"),
		}
		db.SetConfig(cfg)

		if err := db.TestConnection(); err != nil {
			// Check if it's an authentication error
			if strings.Contains(err.Error(), "authentication failed") || strings.Contains(err.Error(), "password authentication failed") {
				fmt.Println("  ⚠️  PostgreSQL authentication failed.")
				
				var autoFix bool
				prompt := &survey.Confirm{
					Message: "Would you like to automatically create the 'pkt_user' and 'pkt_db'? (Requires sudo)",
					Default: true,
				}
				if err := survey.AskOne(prompt, &autoFix); err != nil {
					return fmt.Errorf("cancelled: %w", err)
				}

				if autoFix {
					fmt.Println("\n  🔧 Setting up database...")
					// Create user
					createUserCmd := exec.Command("sudo", "-u", "postgres", "psql", "-c", 
						"DO $$ BEGIN IF NOT EXISTS (SELECT FROM pg_catalog.pg_roles WHERE rolname = 'pkt_user') THEN CREATE ROLE pkt_user WITH LOGIN PASSWORD 'yourpassword'; END IF; END $$;")
					createUserCmd.Stdout = os.Stdout
					createUserCmd.Stderr = os.Stderr
					if err := createUserCmd.Run(); err != nil {
						return fmt.Errorf("failed to create database user: %w", err)
					}

					// Create database
					// First check if it exists
					checkDbCmd := exec.Command("sudo", "-u", "postgres", "psql", "-tAc", "SELECT 1 FROM pg_database WHERE datname='pkt_db'")
					output, _ := checkDbCmd.Output()
					if strings.TrimSpace(string(output)) != "1" {
						createDbCmd := exec.Command("sudo", "-u", "postgres", "psql", "-c", "CREATE DATABASE pkt_db OWNER pkt_user")
						createDbCmd.Stdout = os.Stdout
						createDbCmd.Stderr = os.Stderr
						if err := createDbCmd.Run(); err != nil {
							return fmt.Errorf("failed to create database: %w", err)
						}
					}

					fmt.Println("  ✅ Database user and database created successfully")
					
					// Retry connection
					if err := db.TestConnection(); err != nil {
						return fmt.Errorf("still unable to connect after setup: %w", err)
					}
				} else {
					// User declined auto-fix, show manual instructions
					if !strings.Contains(err.Error(), "does not exist") {
						return fmt.Errorf("PostgreSQL connection failed: %v - please ensure PostgreSQL is installed and running, then run 'pkt start' again", err)
					}
				}
			} else if !strings.Contains(err.Error(), "does not exist") {
				return fmt.Errorf("PostgreSQL connection failed: %v - please ensure PostgreSQL is installed and running, then run 'pkt start' again", err)
			}
			fmt.Println("  ⚠️  Database will be created")
		}
		fmt.Println("  ✅ PostgreSQL available")

		// Step 4: Get configuration from user
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
			fmt.Printf("  ⚠️  Warning: '%s' command not found. You can update this later in ~/.pkt/config.json\n", editorCmd)
		} else {
			fmt.Printf("  ✅ Editor command: %s\n", editorCmd)
		}

		// Step 5: Save configuration
		fmt.Println("\n💾 Saving configuration...")
		
		// Update config with user inputs and DB settings
		cfg.ProjectsRoot = projectsRoot
		cfg.DefaultPM = defaultPM
		cfg.EditorCommand = editorCmd
		cfg.Initialized = true
		
		// Ensure DB settings are set (if they weren't already by auto-fix)
		if cfg.DBUser == "" {
			cfg.DBUser = "pkt_user"
			cfg.DBPassword = "yourpassword"
			cfg.DBName = "pkt_db"
			cfg.DBHost = "127.0.0.1"
			cfg.DBPort = "5432"
		}

		if err := config.Save(cfg); err != nil {
			return fmt.Errorf("failed to save config: %w", err)
		}
		fmt.Println("  ✅ Configuration saved to ~/.pkt/config.json")

		// Step 6: Initialize database
		fmt.Println("\n🗄️  Initializing database...")
		if err := db.InitDB(); err != nil {
			return fmt.Errorf("failed to initialize database: %w", err)
		}
		fmt.Println("  ✅ Database initialized")

		// Final summary
		fmt.Println("\n" + strings.Repeat("=", 60))
		fmt.Println("🎉 pkt is ready to use!")
		fmt.Println(strings.Repeat("=", 60))
		fmt.Println("\n✅ Summary:")
		fmt.Println("  ✅ npm detected")
		if pnpmAvailable {
			fmt.Println("  ✅ pnpm detected")
		}
		fmt.Println("  ✅ PostgreSQL available")
		fmt.Printf("  ✅ Projects folder: %s\n", projectsRoot)
		fmt.Printf("  ✅ Default package manager: %s\n", defaultPM)
		fmt.Printf("  ✅ Editor command: %s\n", editorCmd)
		
		fmt.Println("\n📚 Next steps:")
		fmt.Println("  • Create a project: pkt create <project-name>")
		fmt.Println("  • List projects: pkt list")
		fmt.Println("  • Open a project: pkt open <project-name>")

		return nil
	},
}

// checkTool checks if a command/tool is available in PATH
func checkTool(name string) bool {
	_, err := exec.LookPath(name)
	return err == nil
}

// getEnv gets an environment variable or returns a default value
func getEnv(key, defaultValue string) string {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	return value
}
