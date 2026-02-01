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
		fmt.Println("Starting pkt initialization...")
		fmt.Println()

		// Check if already initialized
		exists, err := config.Exists()
		if err != nil {
			return err
		}

		if exists {
			cfg, err := config.Load()
			if err == nil && cfg.Initialized {
				fmt.Println("pkt is already initialized.")
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

		// Step 1: Check all language tools
		fmt.Println("Checking available languages and tools...")
		fmt.Println()

		// JavaScript
		fmt.Println("  JavaScript:")
		npmAvailable := checkTool("npm")
		pnpmAvailable := checkTool("pnpm")
		bunAvailable := checkTool("bun")
		jsAvailable := npmAvailable || pnpmAvailable || bunAvailable

		if npmAvailable {
			fmt.Println("    [+] npm")
		}
		if pnpmAvailable {
			fmt.Println("    [+] pnpm")
		}
		if bunAvailable {
			fmt.Println("    [+] bun")
		}
		if !jsAvailable {
			fmt.Println("    [-] No JavaScript tools found (install Node.js)")
		}

		// Python
		fmt.Println("  Python:")
		pythonAvailable := checkTool("python3") || checkTool("python")
		pipAvailable := checkTool("pip") || checkTool("pip3")
		poetryAvailable := checkTool("poetry")
		uvAvailable := checkTool("uv")
		pyAvailable := pythonAvailable

		if pythonAvailable {
			fmt.Println("    [+] python")
		} else {
			fmt.Println("    [-] python not found")
		}
		if pipAvailable {
			fmt.Println("    [+] pip")
		}
		if poetryAvailable {
			fmt.Println("    [+] poetry")
		}
		if uvAvailable {
			fmt.Println("    [+] uv")
		}

		// Go
		fmt.Println("  Go:")
		goAvailable := checkTool("go")
		if goAvailable {
			fmt.Println("    [+] go")
		} else {
			fmt.Println("    [-] go not found (install from go.dev)")
		}

		// Rust
		fmt.Println("  Rust:")
		cargoAvailable := checkTool("cargo")
		if cargoAvailable {
			fmt.Println("    [+] cargo")
		} else {
			fmt.Println("    [-] cargo not found (install from rustup.rs)")
		}

		// Git (required for clone)
		fmt.Println("  Git:")
		gitAvailable := checkTool("git")
		if gitAvailable {
			fmt.Println("    [+] git")
		} else {
			fmt.Println("    [!] git not found (required for 'pkt clone')")
		}

		fmt.Println()

		// Summary of available languages
		var availableLangs []string
		if jsAvailable {
			availableLangs = append(availableLangs, "JavaScript")
		}
		if pyAvailable {
			availableLangs = append(availableLangs, "Python")
		}
		if goAvailable {
			availableLangs = append(availableLangs, "Go")
		}
		if cargoAvailable {
			availableLangs = append(availableLangs, "Rust")
		}

		if len(availableLangs) == 0 {
			return fmt.Errorf("no supported languages found. Please install at least one: Node.js, Python, Go, or Rust")
		}

		fmt.Printf("Available languages: %s\n", strings.Join(availableLangs, ", "))
		fmt.Println()

		// Offer to install pnpm if npm is available but pnpm isn't
		if npmAvailable && !pnpmAvailable {
			var shouldInstall bool
			prompt := &survey.Confirm{
				Message: "pnpm is the recommended JS package manager. Install it now?",
				Default: true,
			}
			if err := survey.AskOne(prompt, &shouldInstall); err != nil {
				return fmt.Errorf("cancelled: %w", err)
			}

			if shouldInstall {
				fmt.Println("\n  Installing pnpm globally via npm...")
				if err := pm.InstallPnpm(); err != nil {
					fmt.Printf("  [!] Failed to install pnpm: %v\n", err)
					fmt.Println("  You can install it manually: npm install -g pnpm")
				} else {
					fmt.Println("  [+] pnpm installed successfully")
					pnpmAvailable = true
				}
			}
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

		// Configuration setup
		fmt.Println("\nConfiguration setup...")

		// Projects root
		promptRoot := &survey.Input{
			Message: "Projects root folder:",
			Default: defaultRoot,
		}
		if err := survey.AskOne(promptRoot, &projectsRoot); err != nil {
			return fmt.Errorf("cancelled: %w", err)
		}

		// Expand ~ if present, or prepend home/Documents directory for relative paths
		if len(projectsRoot) >= 2 && projectsRoot[:2] == "~/" {
			projectsRoot = filepath.Join(home, projectsRoot[2:])
		} else if len(projectsRoot) > 0 && projectsRoot[0] != '/' {
			// User entered a relative path (just a folder name), prepend home/Documents directory
			projectsRoot = filepath.Join(home, "Documents", projectsRoot)
		}

		// Create projects root if it doesn't exist
		if err := os.MkdirAll(projectsRoot, 0755); err != nil {
			return fmt.Errorf("failed to create projects root: %w", err)
		}
		fmt.Printf("  Projects folder: %s\n", projectsRoot)

		// Default package manager (from all available)
		available := pm.ListAvailable()
		if len(available) == 0 {
			return fmt.Errorf("no package managers found")
		}

		// Determine best default
		defaultPMOption := "npm"
		if pnpmAvailable {
			defaultPMOption = "pnpm"
		} else if bunAvailable {
			defaultPMOption = "bun"
		} else if len(available) > 0 {
			defaultPMOption = available[0]
		}

		promptPM := &survey.Select{
			Message: "Default package manager (for JavaScript):",
			Options: available,
			Default: defaultPMOption,
		}
		if err := survey.AskOne(promptPM, &defaultPM); err != nil {
			return fmt.Errorf("cancelled: %w", err)
		}
		fmt.Printf("  Default package manager: %s\n", defaultPM)

		// Editor command
		promptEditor := &survey.Input{
			Message: "Editor command (e.g., code, cursor, vim):",
			Default: "code",
		}
		if err := survey.AskOne(promptEditor, &editorCmd); err != nil {
			return fmt.Errorf("cancelled: %w", err)
		}

		// Verify editor command exists
		editorAvailable := checkTool(editorCmd)
		if !editorAvailable {
			fmt.Printf("  [!] Warning: '%s' command not found. Update later with 'pkt config editor <cmd>'\n", editorCmd)
		} else {
		fmt.Printf("  Editor command: %s\n", editorCmd)
		}

		// Save configuration
		fmt.Println("\nSaving configuration...")

		cfg := &config.Config{
			ProjectsRoot:  projectsRoot,
			DefaultPM:     defaultPM,
			EditorCommand: editorCmd,
			Initialized:   true,
		}

		if err := config.Save(cfg); err != nil {
			return fmt.Errorf("failed to save config: %w", err)
		}
		fmt.Println("  Configuration saved to ~/.pkt/config.json")

		// Initialize database
		fmt.Println("\nInitializing database...")
		if err := db.InitDB(); err != nil {
			return fmt.Errorf("failed to initialize database: %w", err)
		}
		fmt.Println("  Database initialized at ~/.pkt/pkt.db")

		// Final summary
		fmt.Println("\n" + strings.Repeat("-", 40))
		fmt.Println("pkt is ready to use!")
		fmt.Println(strings.Repeat("-", 40))

		fmt.Println("\nSupported Languages:")
		if jsAvailable {
			jsPMs := []string{}
			if npmAvailable {
				jsPMs = append(jsPMs, "npm")
			}
			if pnpmAvailable {
				jsPMs = append(jsPMs, "pnpm")
			}
			if bunAvailable {
				jsPMs = append(jsPMs, "bun")
			}
			fmt.Printf("  • JavaScript (%s)\n", strings.Join(jsPMs, ", "))
		}
		if pyAvailable {
			pyPMs := []string{}
			if pipAvailable {
				pyPMs = append(pyPMs, "pip")
			}
			if poetryAvailable {
				pyPMs = append(pyPMs, "poetry")
			}
			if uvAvailable {
				pyPMs = append(pyPMs, "uv")
			}
			if len(pyPMs) == 0 {
				pyPMs = append(pyPMs, "pip")
			}
			fmt.Printf("  • Python (%s)\n", strings.Join(pyPMs, ", "))
		}
		if goAvailable {
			fmt.Println("  • Go (go mod)")
		}
		if cargoAvailable {
			fmt.Println("  • Rust (cargo)")
		}

		fmt.Println("\nNext steps:")
		fmt.Println("  • Create a project: pkt create <name>")
		fmt.Println("  • Initialize existing: pkt init <path>")
		fmt.Println("  • Clone a repo: pkt clone <url>")
		fmt.Println("  • List projects: pkt list")

		return nil
	},
}

// checkTool checks if a command/tool is available in PATH
func checkTool(name string) bool {
	_, err := exec.LookPath(name)
	return err == nil
}
