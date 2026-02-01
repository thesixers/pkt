package pm

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
)

// UV implements PackageManager for uv (fast Python package manager)
// UV handles venv automatically
type UV struct{}

func (u *UV) Name() string {
	return "uv"
}

func (u *UV) Language() string {
	return "python"
}

func (u *UV) Add(workDir string, packages []string, dev bool) error {
	args := []string{"add"}
	if dev {
		args = append(args, "--dev")
	}
	args = append(args, packages...)
	return runCommand("uv", args, workDir)
}

func (u *UV) Remove(workDir string, packages []string) error {
	args := append([]string{"remove"}, packages...)
	return runCommand("uv", args, workDir)
}

func (u *UV) Install(workDir string) error {
	return runCommand("uv", []string{"sync"}, workDir)
}

func (u *UV) Init(workDir string) error {
	return runCommand("uv", []string{"init"}, workDir)
}

func (u *UV) Run(workDir string, script string, args []string) error {
	cmdArgs := []string{"run", script}
	cmdArgs = append(cmdArgs, args...)
	return runCommandInteractive("uv", cmdArgs, workDir)
}

func (u *UV) Update(workDir string, packages []string) error {
	if len(packages) == 0 {
		return runCommand("uv", []string{"lock", "--upgrade"}, workDir)
	}
	args := []string{"lock", "--upgrade-package"}
	args = append(args, packages...)
	return runCommand("uv", args, workDir)
}

func (u *UV) IsAvailable() bool {
	_, err := exec.LookPath("uv")
	return err == nil
}

// Pip implements PackageManager for pip with venv support
type Pip struct{}

func (p *Pip) Name() string {
	return "pip"
}

func (p *Pip) Language() string {
	return "python"
}

// venvPath returns the path to the venv directory
func (p *Pip) venvPath(workDir string) string {
	return filepath.Join(workDir, "venv")
}

// venvPip returns the pip executable path inside venv
func (p *Pip) venvPip(workDir string) string {
	if runtime.GOOS == "windows" {
		return filepath.Join(p.venvPath(workDir), "Scripts", "pip.exe")
	}
	return filepath.Join(p.venvPath(workDir), "bin", "pip")
}

// venvPython returns the python executable path inside venv
func (p *Pip) venvPython(workDir string) string {
	if runtime.GOOS == "windows" {
		return filepath.Join(p.venvPath(workDir), "Scripts", "python.exe")
	}
	return filepath.Join(p.venvPath(workDir), "bin", "python")
}

// ensureVenv creates a virtual environment if it doesn't exist
func (p *Pip) ensureVenv(workDir string) error {
	venvDir := p.venvPath(workDir)

	// Check if venv already exists
	if _, err := os.Stat(venvDir); err == nil {
		return nil
	}

	// Create venv using python -m venv
	fmt.Println("Creating virtual environment...")
	cmd := exec.Command("python3", "-m", "venv", "venv")
	cmd.Dir = workDir
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		// Try python instead of python3
		cmd = exec.Command("python", "-m", "venv", "venv")
		cmd.Dir = workDir
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		if err := cmd.Run(); err != nil {
			return fmt.Errorf("failed to create venv: %w", err)
		}
	}
	fmt.Println("✓ Created venv/")
	return nil
}

func (p *Pip) Add(workDir string, packages []string, dev bool) error {
	// Ensure venv exists
	if err := p.ensureVenv(workDir); err != nil {
		return err
	}

	// Install using venv pip
	pip := p.venvPip(workDir)
	args := append([]string{"install"}, packages...)
	if err := runCommand(pip, args, workDir); err != nil {
		return err
	}

	// Update requirements.txt
	return p.freezeRequirements(workDir, dev)
}

func (p *Pip) freezeRequirements(workDir string, dev bool) error {
	filename := "requirements.txt"
	if dev {
		filename = "requirements-dev.txt"
	}
	reqPath := filepath.Join(workDir, filename)

	pip := p.venvPip(workDir)
	cmd := exec.Command(pip, "freeze")
	cmd.Dir = workDir
	output, err := cmd.Output()
	if err != nil {
		return err
	}

	return os.WriteFile(reqPath, output, 0644)
}

func (p *Pip) Remove(workDir string, packages []string) error {
	// Ensure venv exists
	if err := p.ensureVenv(workDir); err != nil {
		return err
	}

	pip := p.venvPip(workDir)
	args := append([]string{"uninstall", "-y"}, packages...)
	if err := runCommand(pip, args, workDir); err != nil {
		return err
	}

	// Update requirements.txt
	return p.freezeRequirements(workDir, false)
}

func (p *Pip) Install(workDir string) error {
	// Ensure venv exists
	if err := p.ensureVenv(workDir); err != nil {
		return err
	}

	reqPath := filepath.Join(workDir, "requirements.txt")
	if _, err := os.Stat(reqPath); err == nil {
		pip := p.venvPip(workDir)
		return runCommand(pip, []string{"install", "-r", "requirements.txt"}, workDir)
	}
	return nil
}

func (p *Pip) Init(workDir string) error {
	// Create venv
	if err := p.ensureVenv(workDir); err != nil {
		return err
	}

	// Create empty requirements.txt
	reqPath := filepath.Join(workDir, "requirements.txt")
	return os.WriteFile(reqPath, []byte("# Python dependencies\n"), 0644)
}

func (p *Pip) Run(workDir string, script string, args []string) error {
	// Ensure venv exists
	if err := p.ensureVenv(workDir); err != nil {
		return err
	}

	python := p.venvPython(workDir)

	// Handle common scripts
	switch script {
	case "test":
		// Run pytest if available, else python -m pytest
		cmdArgs := []string{"-m", "pytest"}
		cmdArgs = append(cmdArgs, args...)
		return runCommandInteractive(python, cmdArgs, workDir)
	default:
		// Check if it's a Python file that might need special handling
		if strings.HasSuffix(script, ".py") {
			filePath := filepath.Join(workDir, script)
			
			// Check for Streamlit apps
			if isStreamlitApp(filePath) {
				fmt.Println("🎈 Detected Streamlit app, running with streamlit...")
				cmdArgs := []string{"-m", "streamlit", "run", script}
				cmdArgs = append(cmdArgs, args...)
				return runCommandInteractive(python, cmdArgs, workDir)
			}
			
			// Check for ASGI apps (FastAPI, Starlette, Litestar)
			if isASGIApp(filePath) {
				// Run with uvicorn for ASGI apps
				moduleName := strings.TrimSuffix(script, ".py")
				fmt.Println("🚀 Detected ASGI app, running with uvicorn...")
				cmdArgs := []string{"-m", "uvicorn", moduleName + ":app", "--reload"}
				cmdArgs = append(cmdArgs, args...)
				return runCommandInteractive(python, cmdArgs, workDir)
			}
		}
		// Try to run as a Python file
		cmdArgs := []string{script}
		cmdArgs = append(cmdArgs, args...)
		return runCommandInteractive(python, cmdArgs, workDir)
	}
}

// isStreamlitApp checks if a Python file contains Streamlit imports
func isStreamlitApp(filePath string) bool {
	content, err := os.ReadFile(filePath)
	if err != nil {
		return false
	}
	
	fileContent := string(content)
	streamlitPatterns := []string{
		"import streamlit",
		"from streamlit import",
	}
	
	for _, pattern := range streamlitPatterns {
		if strings.Contains(fileContent, pattern) {
			return true
		}
	}
	return false
}

// isASGIApp checks if a Python file contains ASGI framework imports (FastAPI, Starlette, Litestar)
func isASGIApp(filePath string) bool {
	content, err := os.ReadFile(filePath)
	if err != nil {
		return false
	}
	
	fileContent := string(content)
	// Check for common ASGI frameworks
	asgiPatterns := []string{
		"from fastapi import",
		"from fastapi.applications import",
		"import fastapi",
		"FastAPI()",
		"from starlette.applications import",
		"from starlette import",
		"import starlette",
		"Starlette()",
		"from litestar import",
		"import litestar",
		"Litestar(",
	}
	
	for _, pattern := range asgiPatterns {
		if strings.Contains(fileContent, pattern) {
			return true
		}
	}
	return false
}

func (p *Pip) Update(workDir string, packages []string) error {
	// Ensure venv exists
	if err := p.ensureVenv(workDir); err != nil {
		return err
	}

	pip := p.venvPip(workDir)

	if len(packages) == 0 {
		// Update all packages from requirements.txt
		reqPath := filepath.Join(workDir, "requirements.txt")
		if _, err := os.Stat(reqPath); err == nil {
			return runCommand(pip, []string{"install", "-U", "-r", "requirements.txt"}, workDir)
		}
		return nil
	}

	args := []string{"install", "-U"}
	args = append(args, packages...)
	return runCommand(pip, args, workDir)
}

func (p *Pip) IsAvailable() bool {
	_, err := exec.LookPath("pip")
	if err != nil {
		_, err = exec.LookPath("pip3")
	}
	return err == nil
}

// Poetry implements PackageManager for poetry
// Poetry handles venv automatically when in-project = true
type Poetry struct{}

func (p *Poetry) Name() string {
	return "poetry"
}

func (p *Poetry) Language() string {
	return "python"
}

func (p *Poetry) Add(workDir string, packages []string, dev bool) error {
	args := []string{"add"}
	if dev {
		args = append(args, "--group", "dev")
	}
	args = append(args, packages...)
	return runCommand("poetry", args, workDir)
}

func (p *Poetry) Remove(workDir string, packages []string) error {
	args := append([]string{"remove"}, packages...)
	return runCommand("poetry", args, workDir)
}

func (p *Poetry) Install(workDir string) error {
	return runCommand("poetry", []string{"install"}, workDir)
}

func (p *Poetry) Init(workDir string) error {
	// Initialize poetry project
	if err := runCommand("poetry", []string{"init", "-n"}, workDir); err != nil {
		return err
	}
	// Configure poetry to create venv in project directory
	return runCommand("poetry", []string{"config", "virtualenvs.in-project", "true", "--local"}, workDir)
}

func (p *Poetry) Run(workDir string, script string, args []string) error {
	cmdArgs := []string{"run", script}
	cmdArgs = append(cmdArgs, args...)
	return runCommandInteractive("poetry", cmdArgs, workDir)
}

func (p *Poetry) Update(workDir string, packages []string) error {
	args := []string{"update"}
	args = append(args, packages...)
	return runCommand("poetry", args, workDir)
}

func (p *Poetry) IsAvailable() bool {
	_, err := exec.LookPath("poetry")
	return err == nil
}
