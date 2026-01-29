package cmd

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/genesix/pkt/internal/config"
	"github.com/genesix/pkt/internal/db"
	"github.com/genesix/pkt/internal/utils"
	"github.com/spf13/cobra"
)

var initCmd = &cobra.Command{
	Use:   "init [path]",
	Short: "Initialize an existing project for pkt management",
	Long: `Initialize an existing project for pkt management.
This command takes an existing project folder and adds it to pkt's tracking system.

If the project is outside pkt's projects folder, it will be moved there.
The project must contain a valid package.json file.

Examples:
  pkt init .                        # Initialize current directory
  pkt init /path/to/my-project      # Initialize a specific project`,
	Args: cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		// Get the path (default to current directory)
		inputPath := "."
		if len(args) > 0 {
			inputPath = args[0]
		}

		// Resolve the path to absolute
		absPath, err := filepath.Abs(inputPath)
		if err != nil {
			return fmt.Errorf("failed to resolve path: %w", err)
		}

		// Check if folder exists
		info, err := os.Stat(absPath)
		if err != nil {
			if os.IsNotExist(err) {
				return fmt.Errorf("path does not exist: %s", absPath)
			}
			return fmt.Errorf("failed to access path: %w", err)
		}
		if !info.IsDir() {
			return fmt.Errorf("path is not a directory: %s", absPath)
		}

		// Check for package.json
		pkgJSONPath := filepath.Join(absPath, "package.json")
		if _, err := os.Stat(pkgJSONPath); os.IsNotExist(err) {
			return fmt.Errorf("no package.json found in %s\nThis command requires an existing Node.js project", absPath)
		}

		// Read package.json to get project name
		pkgData, err := os.ReadFile(pkgJSONPath)
		if err != nil {
			return fmt.Errorf("failed to read package.json: %w", err)
		}

		var pkg utils.PackageJSON
		if err := json.Unmarshal(pkgData, &pkg); err != nil {
			return fmt.Errorf("failed to parse package.json: %w", err)
		}

		// Get project name from package.json or folder name
		projectName := pkg.Name
		if projectName == "" {
			projectName = filepath.Base(absPath)
		}

		// Detect package manager from lockfiles
		pm := detectPackageManager(absPath)

		// Load config
		cfg, err := config.Load()
		if err != nil {
			return err
		}

		// Check if project already exists in database
		existingProject, err := db.GetProjectByPath(absPath)
		if err == nil && existingProject != nil {
			return fmt.Errorf("project is already tracked with ID: %s", existingProject.ID)
		}

		// Determine final path
		finalPath := absPath

		// Check if project is outside pkt's projects folder
		projectsRoot, err := utils.ExpandPath(cfg.ProjectsRoot)
		if err != nil {
			return fmt.Errorf("failed to expand projects root: %w", err)
		}

		if !strings.HasPrefix(absPath, projectsRoot) {
			fmt.Printf("📦 Project is outside pkt workspace, moving to %s...\n", projectsRoot)

			// Get unique folder name
			targetPath := filepath.Join(projectsRoot, filepath.Base(absPath))
			targetPath, err = getUniqueFolder(targetPath)
			if err != nil {
				return fmt.Errorf("failed to find unique folder name: %w", err)
			}

			// Move the project
			if err := moveProject(absPath, targetPath); err != nil {
				return fmt.Errorf("failed to move project: %w", err)
			}

			finalPath = targetPath
			fmt.Printf("✓ Moved to: %s\n", finalPath)
		}

		// Generate project ID
		projectID := utils.GenerateID()

		// Insert project into database
		project, err := db.CreateProject(projectID, projectName, finalPath, pm)
		if err != nil {
			return fmt.Errorf("failed to add project to database: %w", err)
		}

		// Parse and sync dependencies
		deps, err := utils.ParsePackageJSON(finalPath)
		if err != nil {
			fmt.Printf("⚠️  Warning: failed to parse dependencies: %v\n", err)
		} else if len(deps) > 0 {
			if err := db.SyncDependencies(project.ID, deps); err != nil {
				fmt.Printf("⚠️  Warning: failed to sync dependencies: %v\n", err)
			} else {
				fmt.Printf("✓ Synced %d dependencies\n", len(deps))
			}
		}

		fmt.Println()
		fmt.Printf("✓ Initialized project: %s\n", projectName)
		fmt.Printf("  ID: %s\n", project.ID)
		fmt.Printf("  Path: %s\n", project.Path)
		fmt.Printf("  Package Manager: %s\n", project.PackageManager)

		return nil
	},
}

// detectPackageManager detects the package manager from lockfiles
func detectPackageManager(projectPath string) string {
	// Check for pnpm-lock.yaml
	if _, err := os.Stat(filepath.Join(projectPath, "pnpm-lock.yaml")); err == nil {
		return "pnpm"
	}
	// Check for bun.lockb
	if _, err := os.Stat(filepath.Join(projectPath, "bun.lockb")); err == nil {
		return "bun"
	}
	// Check for package-lock.json
	if _, err := os.Stat(filepath.Join(projectPath, "package-lock.json")); err == nil {
		return "npm"
	}
	// Check for yarn.lock and fallback to npm since we don't support yarn yet
	if _, err := os.Stat(filepath.Join(projectPath, "yarn.lock")); err == nil {
		return "npm" // Fallback to npm for yarn projects
	}
	// Default to pnpm
	return "pnpm"
}

// getUniqueFolder returns a unique folder path, appending numbers if needed
func getUniqueFolder(basePath string) (string, error) {
	path := basePath
	for i := 1; i <= 100; i++ {
		if _, err := os.Stat(path); os.IsNotExist(err) {
			return path, nil
		}
		path = fmt.Sprintf("%s-%d", basePath, i)
	}
	return "", fmt.Errorf("could not find unique folder name after 100 attempts")
}

// moveProject moves a project directory to a new location
func moveProject(src, dst string) error {
	// Try simple rename first (works if on same filesystem)
	if err := os.Rename(src, dst); err == nil {
		return nil
	}

	// If rename fails, do a copy then delete
	if err := copyDir(src, dst); err != nil {
		return fmt.Errorf("failed to copy directory: %w", err)
	}

	if err := os.RemoveAll(src); err != nil {
		return fmt.Errorf("failed to remove source directory: %w", err)
	}

	return nil
}

// copyDir recursively copies a directory
func copyDir(src, dst string) error {
	srcInfo, err := os.Stat(src)
	if err != nil {
		return err
	}

	if err := os.MkdirAll(dst, srcInfo.Mode()); err != nil {
		return err
	}

	entries, err := os.ReadDir(src)
	if err != nil {
		return err
	}

	for _, entry := range entries {
		srcPath := filepath.Join(src, entry.Name())
		dstPath := filepath.Join(dst, entry.Name())

		if entry.IsDir() {
			if err := copyDir(srcPath, dstPath); err != nil {
				return err
			}
		} else {
			if err := copyFile(srcPath, dstPath); err != nil {
				return err
			}
		}
	}

	return nil
}

// copyFile copies a single file
func copyFile(src, dst string) error {
	srcFile, err := os.Open(src)
	if err != nil {
		return err
	}
	defer func() { _ = srcFile.Close() }()

	srcInfo, err := srcFile.Stat()
	if err != nil {
		return err
	}

	dstFile, err := os.OpenFile(dst, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, srcInfo.Mode())
	if err != nil {
		return err
	}
	defer func() { _ = dstFile.Close() }()

	_, err = io.Copy(dstFile, srcFile)
	return err
}

func init() {
	rootCmd.AddCommand(initCmd)
}
