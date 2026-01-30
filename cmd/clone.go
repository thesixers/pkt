package cmd

import (
	"fmt"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/genesix/pkt/internal/config"
	"github.com/genesix/pkt/internal/db"
	"github.com/genesix/pkt/internal/lang"
	"github.com/genesix/pkt/internal/pm"
	"github.com/genesix/pkt/internal/utils"
	"github.com/spf13/cobra"
)

var (
	cloneName    string
	cloneInstall bool
)

var cloneCmd = &cobra.Command{
	Use:   "clone <repository>",
	Short: "Clone a git repository and track it",
	Long: `Clone a git repository into the pkt workspace and automatically track it.

The repository will be cloned to the workspace directory, language will be
auto-detected, and the project will be registered in the database.

Examples:
  pkt clone https://github.com/user/repo
  pkt clone git@github.com:user/repo.git
  pkt clone https://github.com/user/repo --name my-project
  pkt clone https://github.com/user/repo --install`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		repoURL := args[0]

		// Get workspace directory
		cfg, err := config.Load()
		if err != nil {
			return fmt.Errorf("failed to load config: %w", err)
		}

		// Determine project name
		projectName := cloneName
		if projectName == "" {
			projectName = extractRepoName(repoURL)
		}

		// Target directory
		targetPath := filepath.Join(cfg.ProjectsRoot, projectName)

		// Check if directory already exists
		if _, err := os.Stat(targetPath); err == nil {
			return fmt.Errorf("directory already exists: %s", targetPath)
		}

		// Clone the repository
		fmt.Printf("📥 Cloning %s...\n", repoURL)
		cloneCmd := exec.Command("git", "clone", repoURL, targetPath)
		cloneCmd.Stdout = os.Stdout
		cloneCmd.Stderr = os.Stderr
		if err := cloneCmd.Run(); err != nil {
			return fmt.Errorf("failed to clone repository: %w", err)
		}

		// Auto-detect language
		detectedLang, err := lang.Detect(targetPath)
		if err != nil {
			fmt.Printf("⚠️  Could not detect language, defaulting to JavaScript\n")
			detectedLang, _ = lang.Get("javascript")
		}

		// Detect package manager
		packageManager := detectedLang.DetectPackageManager(targetPath)

		// Generate project ID
		projectID := utils.GenerateID()

		// Create project in database
		project, err := db.CreateProject(projectID, projectName, targetPath, detectedLang.Name(), packageManager)
		if err != nil {
			return fmt.Errorf("failed to register project: %w", err)
		}

		// Sync dependencies
		deps, parseErr := utils.ParseDependencies(targetPath, detectedLang.Name())
		if parseErr != nil {
			fmt.Printf("⚠️  Warning: failed to parse dependencies: %v\n", parseErr)
		} else if len(deps) > 0 {
			if syncErr := db.SyncDependencies(project.ID, deps); syncErr != nil {
				fmt.Printf("⚠️  Warning: failed to sync dependencies: %v\n", syncErr)
			}
		}

		fmt.Println()
		fmt.Printf("✓ Cloned and registered: %s\n", projectName)
		fmt.Printf("  ID: %s\n", project.ID)
		fmt.Printf("  Path: %s\n", project.Path)
		fmt.Printf("  Language: %s\n", detectedLang.DisplayName())
		fmt.Printf("  Package Manager: %s\n", project.PackageManager)

		// Optionally run install
		if cloneInstall {
			fmt.Println()
			fmt.Println("📦 Installing dependencies...")

			pkgMgr, err := pm.Get(detectedLang.Name(), packageManager)
			if err != nil {
				return fmt.Errorf("failed to get package manager: %w", err)
			}

			if err := pkgMgr.Install(targetPath); err != nil {
				return fmt.Errorf("failed to install dependencies: %w", err)
			}

			fmt.Println("✓ Dependencies installed")
		}

		return nil
	},
}

// extractRepoName extracts the repository name from a git URL
func extractRepoName(repoURL string) string {
	// Handle HTTPS URLs
	if strings.HasPrefix(repoURL, "https://") || strings.HasPrefix(repoURL, "http://") {
		u, err := url.Parse(repoURL)
		if err == nil {
			path := strings.TrimSuffix(u.Path, ".git")
			parts := strings.Split(path, "/")
			if len(parts) > 0 {
				return parts[len(parts)-1]
			}
		}
	}

	// Handle SSH URLs (git@github.com:user/repo.git)
	if strings.HasPrefix(repoURL, "git@") {
		parts := strings.Split(repoURL, ":")
		if len(parts) == 2 {
			path := strings.TrimSuffix(parts[1], ".git")
			pathParts := strings.Split(path, "/")
			if len(pathParts) > 0 {
				return pathParts[len(pathParts)-1]
			}
		}
	}

	// Fallback: just use the last path segment
	parts := strings.Split(strings.TrimSuffix(repoURL, ".git"), "/")
	return parts[len(parts)-1]
}

func init() {
	cloneCmd.Flags().StringVarP(&cloneName, "name", "n", "", "Custom name for the project")
	cloneCmd.Flags().BoolVarP(&cloneInstall, "install", "i", false, "Run install after cloning")
	rootCmd.AddCommand(cloneCmd)
}
