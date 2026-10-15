package cmd

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/genesix/pkt/internal/utils"
	"github.com/spf13/cobra"
)

// AutocompleteResult represents the result of setting up autocomplete.
type AutocompleteResult struct {
	OS         string
	Shell      string
	ScriptPath string
	RcPath     string
	Message    string
	UpdatedRc  bool
}

// DetectShell identifies the user's active shell from environment or OS defaults.
func DetectShell() string {
	shellEnv := os.Getenv("SHELL")
	if shellEnv != "" {
		base := strings.ToLower(filepath.Base(shellEnv))
		switch {
		case strings.Contains(base, "bash"):
			return "bash"
		case strings.Contains(base, "zsh"):
			return "zsh"
		case strings.Contains(base, "fish"):
			return "fish"
		}
	}

	if os.Getenv("ZSH_NAME") != "" || os.Getenv("ZSH_VERSION") != "" {
		return "zsh"
	}
	if os.Getenv("BASH_VERSION") != "" {
		return "bash"
	}
	if os.Getenv("FISH_VERSION") != "" {
		return "fish"
	}

	if runtime.GOOS == "windows" {
		return "powershell"
	}
	if runtime.GOOS == "darwin" {
		return "zsh"
	}
	return "bash"
}

// InstallAutocomplete generates and installs the completion script for the target shell.
// If targetShell is empty, it detects the shell automatically.
func InstallAutocomplete(root *cobra.Command, targetShell string) (*AutocompleteResult, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return nil, fmt.Errorf("failed to get user home directory: %w", err)
	}
	return InstallAutocompleteTo(root, targetShell, home)
}

// InstallAutocompleteTo installs autocomplete for a given shell and home directory.
func InstallAutocompleteTo(root *cobra.Command, targetShell, home string) (*AutocompleteResult, error) {
	if targetShell == "" {
		targetShell = DetectShell()
	}
	targetShell = strings.ToLower(strings.TrimSpace(targetShell))

	switch targetShell {
	case "bash":
		return installBash(root, home)
	case "zsh":
		return installZsh(root, home)
	case "fish":
		return installFish(root, home)
	case "powershell", "pwsh":
		return installPowerShell(root, home)
	default:
		return nil, fmt.Errorf("unsupported shell %q (supported: bash, zsh, fish, powershell)", targetShell)
	}
}

func installBash(root *cobra.Command, home string) (*AutocompleteResult, error) {
	var buf bytes.Buffer
	if err := root.GenBashCompletionV2(&buf, true); err != nil {
		return nil, fmt.Errorf("failed to generate bash completion: %w", err)
	}
	script := buf.Bytes()

	// 1. Try standard bash-completion user directory:
	// ~/.local/share/bash-completion/completions/pkt
	compDir := filepath.Join(home, ".local", "share", "bash-completion", "completions")
	if xdg := os.Getenv("XDG_DATA_HOME"); xdg != "" {
		compDir = filepath.Join(xdg, "bash-completion", "completions")
	}

	standardPath := filepath.Join(compDir, "pkt")
	wroteStandard := false
	if err := os.MkdirAll(compDir, 0755); err == nil {
		if err := os.WriteFile(standardPath, script, 0644); err == nil {
			wroteStandard = true
		}
	}

	// 2. Also save in ~/.pkt/completions/pkt.bash as standard / fallback location
	pktDir := filepath.Join(home, ".pkt", "completions")
	pktPath := filepath.Join(pktDir, "pkt.bash")
	wrotePkt := false
	if err := os.MkdirAll(pktDir, 0755); err == nil {
		if err := os.WriteFile(pktPath, script, 0644); err == nil {
			wrotePkt = true
		}
	}

	if !wroteStandard && !wrotePkt {
		return nil, fmt.Errorf("failed to write completion script to %s or %s", compDir, pktDir)
	}

	activeScript := standardPath
	if !wroteStandard {
		activeScript = pktPath
	}

	// 3. Ensure ~/.bashrc (or ~/.bash_profile on macOS) sources the completion
	rcFile := filepath.Join(home, ".bashrc")
	if runtime.GOOS == "darwin" && !pathExists(rcFile) {
		rcFile = filepath.Join(home, ".bash_profile")
	}

	updatedRc := false
	snippet := fmt.Sprintf("\n# pkt shell completion\nif [ -f %q ]; then\n    source %q\nelif [ -f %q ]; then\n    source %q\nfi\n", standardPath, standardPath, pktPath, pktPath)

	if shouldAppend(rcFile, "completions/pkt") {
		if err := appendString(rcFile, snippet); err == nil {
			updatedRc = true
		}
	}

	msg := fmt.Sprintf("Saved bash completion to %s", utils.ShortPath(activeScript))
	if updatedRc {
		msg += fmt.Sprintf(" and added source to %s", utils.ShortPath(rcFile))
	}

	return &AutocompleteResult{
		OS:         runtime.GOOS,
		Shell:      "bash",
		ScriptPath: activeScript,
		RcPath:     rcFile,
		Message:    msg,
		UpdatedRc:  updatedRc,
	}, nil
}

func installZsh(root *cobra.Command, home string) (*AutocompleteResult, error) {
	var buf bytes.Buffer
	if err := root.GenZshCompletion(&buf); err != nil {
		return nil, fmt.Errorf("failed to generate zsh completion: %w", err)
	}
	script := buf.Bytes()

	// 1. Write to ~/.pkt/completions/pkt.zsh
	compDir := filepath.Join(home, ".pkt", "completions")
	if err := os.MkdirAll(compDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create directory %s: %w", compDir, err)
	}
	compFile := filepath.Join(compDir, "pkt.zsh")
	if err := os.WriteFile(compFile, script, 0644); err != nil {
		return nil, fmt.Errorf("failed to write completion script to %s: %w", compFile, err)
	}

	// 2. If Oh-My-Zsh or ~/.zfunc exists, write _pkt there too
	omzDir := filepath.Join(home, ".oh-my-zsh", "completions")
	if isDir(omzDir) {
		_ = os.WriteFile(filepath.Join(omzDir, "_pkt"), script, 0644)
	}
	zfuncDir := filepath.Join(home, ".zfunc")
	if isDir(zfuncDir) {
		_ = os.WriteFile(filepath.Join(zfuncDir, "_pkt"), script, 0644)
	}

	// 3. Ensure ~/.zshrc sources it
	rcFile := filepath.Join(home, ".zshrc")
	updatedRc := false
	snippet := fmt.Sprintf("\n# pkt shell completion\nif [ -f %q ]; then\n    autoload -Uz compinit 2>/dev/null && compinit 2>/dev/null\n    source %q\nfi\n", compFile, compFile)

	if shouldAppend(rcFile, "completions/pkt.zsh") {
		if err := appendString(rcFile, snippet); err == nil {
			updatedRc = true
		}
	}

	msg := fmt.Sprintf("Saved zsh completion to %s", utils.ShortPath(compFile))
	if updatedRc {
		msg += fmt.Sprintf(" and added source to %s", utils.ShortPath(rcFile))
	}

	return &AutocompleteResult{
		OS:         runtime.GOOS,
		Shell:      "zsh",
		ScriptPath: compFile,
		RcPath:     rcFile,
		Message:    msg,
		UpdatedRc:  updatedRc,
	}, nil
}

func installFish(root *cobra.Command, home string) (*AutocompleteResult, error) {
	var buf bytes.Buffer
	if err := root.GenFishCompletion(&buf, true); err != nil {
		return nil, fmt.Errorf("failed to generate fish completion: %w", err)
	}
	script := buf.Bytes()

	compDir := filepath.Join(home, ".config", "fish", "completions")
	if xdg := os.Getenv("XDG_CONFIG_HOME"); xdg != "" {
		compDir = filepath.Join(xdg, "fish", "completions")
	}
	if err := os.MkdirAll(compDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create directory %s: %w", compDir, err)
	}
	compFile := filepath.Join(compDir, "pkt.fish")
	if err := os.WriteFile(compFile, script, 0644); err != nil {
		return nil, fmt.Errorf("failed to write completion script to %s: %w", compFile, err)
	}

	return &AutocompleteResult{
		OS:         runtime.GOOS,
		Shell:      "fish",
		ScriptPath: compFile,
		Message:    fmt.Sprintf("Saved fish completion to %s", utils.ShortPath(compFile)),
	}, nil
}

func installPowerShell(root *cobra.Command, home string) (*AutocompleteResult, error) {
	var buf bytes.Buffer
	if err := root.GenPowerShellCompletionWithDesc(&buf); err != nil {
		return nil, fmt.Errorf("failed to generate powershell completion: %w", err)
	}
	script := buf.Bytes()

	compDir := filepath.Join(home, ".pkt", "completions")
	if err := os.MkdirAll(compDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create directory %s: %w", compDir, err)
	}
	compFile := filepath.Join(compDir, "pkt.ps1")
	if err := os.WriteFile(compFile, script, 0644); err != nil {
		return nil, fmt.Errorf("failed to write completion script to %s: %w", compFile, err)
	}

	profileFile := filepath.Join(home, "Documents", "PowerShell", "Microsoft.PowerShell_profile.ps1")
	if runtime.GOOS != "windows" {
		profileFile = filepath.Join(home, ".config", "powershell", "Microsoft.PowerShell_profile.ps1")
	}

	updatedRc := false
	snippet := fmt.Sprintf("\n# pkt shell completion\nif (Test-Path %q) {\n    . %q\n}\n", compFile, compFile)

	if shouldAppend(profileFile, "completions/pkt.ps1") {
		_ = os.MkdirAll(filepath.Dir(profileFile), 0755)
		if err := appendString(profileFile, snippet); err == nil {
			updatedRc = true
		}
	}

	msg := fmt.Sprintf("Saved powershell completion to %s", utils.ShortPath(compFile))
	if updatedRc {
		msg += fmt.Sprintf(" and added source to %s", utils.ShortPath(profileFile))
	}

	return &AutocompleteResult{
		OS:         runtime.GOOS,
		Shell:      "powershell",
		ScriptPath: compFile,
		RcPath:     profileFile,
		Message:    msg,
		UpdatedRc:  updatedRc,
	}, nil
}

func shouldAppend(filePath, marker string) bool {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return true
	}
	content := string(data)
	return !strings.Contains(content, marker) && !strings.Contains(content, "pkt completion")
}

func appendString(filePath, text string) error {
	f, err := os.OpenFile(filePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}
	defer f.Close()
	_, err = f.WriteString(text)
	return err
}

func pathExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

func isDir(path string) bool {
	info, err := os.Stat(path)
	return err == nil && info.IsDir()
}

var completionCmd = &cobra.Command{
	Use:   "completion [command]",
	Short: "Generate or install autocompletion scripts",
	Long: `Generate or install autocompletion scripts for pkt.

To automatically detect your shell and install completion:
  pkt completion install

To print the completion script for manual sourcing:
  pkt completion bash
  pkt completion zsh
  pkt completion fish
  pkt completion powershell`,
}

var completionBashCmd = &cobra.Command{
	Use:   "bash",
	Short: "Generate bash autocompletion script",
	RunE: func(cmd *cobra.Command, args []string) error {
		return rootCmd.GenBashCompletionV2(os.Stdout, true)
	},
}

var completionZshCmd = &cobra.Command{
	Use:   "zsh",
	Short: "Generate zsh autocompletion script",
	RunE: func(cmd *cobra.Command, args []string) error {
		return rootCmd.GenZshCompletion(os.Stdout)
	},
}

var completionFishCmd = &cobra.Command{
	Use:   "fish",
	Short: "Generate fish autocompletion script",
	RunE: func(cmd *cobra.Command, args []string) error {
		return rootCmd.GenFishCompletion(os.Stdout, true)
	},
}

var completionPowerShellCmd = &cobra.Command{
	Use:     "powershell",
	Aliases: []string{"pwsh"},
	Short:   "Generate powershell autocompletion script",
	RunE: func(cmd *cobra.Command, args []string) error {
		return rootCmd.GenPowerShellCompletionWithDesc(os.Stdout)
	},
}

var completionInstallCmd = &cobra.Command{
	Use:   "install [bash|zsh|fish|powershell]",
	Short: "Automatically detect and install shell autocomplete",
	Long: `Detect operating system and shell, generate the autocomplete script,
and install it to the appropriate user location so autocomplete is ready for use.`,
	Args:      cobra.MaximumNArgs(1),
	ValidArgs: []string{"bash", "zsh", "fish", "powershell"},
	RunE: func(cmd *cobra.Command, args []string) error {
		target := ""
		if len(args) > 0 {
			target = args[0]
		}
		res, err := InstallAutocomplete(rootCmd, target)
		if err != nil {
			return err
		}
		fmt.Printf("✓ %s\n", res.Message)
		fmt.Println("Restart your terminal or run 'exec $SHELL' to activate autocomplete.")
		return nil
	},
}

func init() {
	completionCmd.AddCommand(completionBashCmd)
	completionCmd.AddCommand(completionZshCmd)
	completionCmd.AddCommand(completionFishCmd)
	completionCmd.AddCommand(completionPowerShellCmd)
	completionCmd.AddCommand(completionInstallCmd)
	rootCmd.AddCommand(completionCmd)
}
