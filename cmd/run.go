package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/genesix/pkt/internal/config"
	"github.com/genesix/pkt/internal/db"
	"github.com/genesix/pkt/internal/pm"
	"github.com/spf13/cobra"
)

var runCmd = &cobra.Command{
	Use:   "run [interpreter-flags...] <file | script | keyword | .> [args...]",
	Short: "Run a script, file, or command",
	Long: `Run a file or script seamlessly. 

Any flags before the filename belong to the interpreter (e.g. node, python).
Any flags after the filename belong to your script.

Examples:
  pkt run main.py                  # Run Python file
  pkt run app.js --port 3000       # Pass flag to your script
  pkt run --inspect app.js         # Pass flag to Node
  pkt run dev                      # Run package.json script
  pkt run .                        # Run default project file/script`,
	DisableFlagParsing: true,
	ValidArgsFunction: completeScripts,
	RunE: func(cmd *cobra.Command, args []string) error {
		if len(args) == 0 {
			printNoArgsHelp()
			return nil
		}

		if args[0] == "-h" || args[0] == "--help" {
			return cmd.Help()
		}

		// Split positional args
		targetIdx := -1
		for i, arg := range args {
			if !strings.HasPrefix(arg, "-") {
				targetIdx = i
				break
			}
		}

		if targetIdx == -1 {
			return fmt.Errorf("no target specified to run. Specify a file, script, or '.'")
		}

		interpFlags := args[:targetIdx]
		target := args[targetIdx]
		scriptArgs := args[targetIdx+1:]

		cwd, err := os.Getwd()
		if err != nil {
			return fmt.Errorf("failed to get current directory: %w", err)
		}

		project, projectErr := db.GetProjectByPath(cwd)
		inTrackedProject := (projectErr == nil)

		// Check if target is a file or directory
		info, statErr := os.Stat(target)
		if statErr == nil {
			var pManager pm.PackageManager
			if inTrackedProject {
				pManager, _ = pm.Get(project.Language, project.PackageManager)
			}

			// If untracked or unknown PM, infer from extension
			if pManager == nil {
				cfg, _ := config.Load()
				ext := filepath.Ext(target)
				if ext == ".ts" || ext == ".tsx" || ext == ".js" || ext == ".mjs" {
					defaultJS := "npm"
					if cfg != nil && cfg.DefaultPM != "" {
						defaultJS = cfg.DefaultPM
					}
					pManager, _ = pm.Get("javascript", defaultJS)
				} else if ext == ".py" {
					pManager, _ = pm.Get("python", "pip")
				} else if ext == ".go" {
					pManager, _ = pm.Get("go", "go")
				} else if ext == ".rs" {
					pManager, _ = pm.Get("rust", "cargo")
				} else {
					return executeDirect(target, interpFlags, scriptArgs, cwd)
				}
			}

			if info.IsDir() {
				if pManager != nil && pManager.Language() == "go" {
					target = "./" + target
					return pManager.RunFile(cwd, target, interpFlags, scriptArgs)
				}
				return fmt.Errorf("cannot run a directory directly except for Go projects")
			}

			return pManager.RunFile(cwd, target, interpFlags, scriptArgs)
		}

		// Target is not a file on disk
		if !inTrackedProject {
			return fmt.Errorf("'%s' is not a file, and you are not in a tracked project. Run 'pkt init .' first", target)
		}

		pManager, err := pm.Get(project.Language, project.PackageManager)
		if err != nil {
			return err
		}

		if target == "." {
			target = getDefaultTarget(project.Language, cwd)
			if target == "" {
				return fmt.Errorf("no default run target found for this project")
			}
			_, err := os.Stat(target)
			if err == nil {
				return pManager.RunFile(cwd, target, interpFlags, scriptArgs)
			}
			return pManager.Run(cwd, target, scriptArgs)
		}

		// Try running as a script/keyword
		err = pManager.Run(cwd, target, append(interpFlags, scriptArgs...))
		if err != nil && project.Language == "javascript" {
			suggestJSScripts(cwd, target)
		}
		return err
	},
}

func printNoArgsHelp() {
	fmt.Println("No target specified.")
	fmt.Println("\nAvailable project scripts:")
	cwd, _ := os.Getwd()
	if cwd != "" {
		scripts := parsePackageJSONScripts(cwd)
		if len(scripts) > 0 {
			for _, s := range scripts {
				fmt.Printf("  %s\n", s)
			}
		} else {
			fmt.Println("  (none found)")
		}
	}
	fmt.Println("\nOr run a file:")
	fmt.Println("  pkt run main.py")
}

func getDefaultTarget(language string, workDir string) string {
	switch language {
	case "go":
		return "."
	case "rust":
		return "run"
	case "python":
		if _, err := os.Stat(filepath.Join(workDir, "main.py")); err == nil {
			return "main.py"
		}
		if _, err := os.Stat(filepath.Join(workDir, "app.py")); err == nil {
			return "app.py"
		}
		return "main.py"
	case "javascript":
		scripts := parsePackageJSONScripts(workDir)
		for _, s := range []string{"dev", "start"} {
			if sliceContains(scripts, s) {
				return s
			}
		}
		if _, err := os.Stat(filepath.Join(workDir, "index.js")); err == nil {
			return "index.js"
		}
		if _, err := os.Stat(filepath.Join(workDir, "src", "index.ts")); err == nil {
			return "src/index.ts"
		}
	}
	return ""
}

func suggestJSScripts(workDir, target string) {
	scripts := parsePackageJSONScripts(workDir)
	if len(scripts) > 0 {
		fmt.Printf("\nScript '%s' not found in package.json. Available scripts:\n", target)
		for _, s := range scripts {
			fmt.Printf("  %s\n", s)
		}
	}
}

func executeDirect(target string, interpFlags, args []string, cwd string) error {
	// Let the OS execute it directly
	cmdArgs := append(interpFlags, target)
	cmdArgs = append(cmdArgs, args...)
	
	// Assume the first flag is not the executable if we are direct executing
	executable := "./" + target
	execArgs := cmdArgs[1:]
	if len(interpFlags) > 0 {
		executable = interpFlags[0]
		execArgs = cmdArgs[1:]
	} else {
		executable = cmdArgs[0]
		if !strings.Contains(executable, "/") {
			executable = "./" + executable
		}
	}

	cmd := exec.Command(executable, execArgs...)
	cmd.Dir = cwd
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin
	return cmd.Run()
}

func init() {
	rootCmd.AddCommand(runCmd)
}
