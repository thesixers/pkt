package utils

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/AlecAivazis/survey/v2"
	"github.com/genesix/pkt/internal/db"
	"github.com/oklog/ulid/v2"
)

// ResolveProject resolves a project by name, ID, or current directory
func ResolveProject(input string) (*db.Project, error) {
	// Case 1: Current directory (".")
	if input == "." {
		cwd, err := os.Getwd()
		if err != nil {
			return nil, fmt.Errorf("failed to get current directory: %w", err)
		}
		return db.GetProjectByPath(cwd)
	}

	// Case 2: Try by ID (ULID format)
	if isULID(input) {
		project, err := db.GetProjectByID(input)
		if err == nil {
			return project, nil
		}
		// If not found by ID, try by name
	}

	// Case 3: Try by name
	projects, err := db.GetProjectsByName(input)
	if err != nil {
		return nil, fmt.Errorf("failed to query projects: %w", err)
	}

	if len(projects) == 0 {
		return nil, fmt.Errorf("project not found: %s", input)
	}

	if len(projects) == 1 {
		return projects[0], nil
	}

	// Case 4: Multiple projects with same name - prompt user
	return selectProject(projects)
}

// selectProject prompts the user to select from multiple projects
func selectProject(projects []*db.Project) (*db.Project, error) {
	options := make([]string, len(projects))
	for i, p := range projects {
		options[i] = fmt.Sprintf("%s (%s) - %s", p.Name, p.ID, p.Path)
	}

	var selected string
	prompt := &survey.Select{
		Message: "Multiple projects found. Select one:",
		Options: options,
	}

	if err := survey.AskOne(prompt, &selected); err != nil {
		return nil, fmt.Errorf("project selection cancelled: %w", err)
	}

	// Find selected project
	for i, opt := range options {
		if opt == selected {
			return projects[i], nil
		}
	}

	return nil, fmt.Errorf("invalid selection")
}

// isULID checks if a string looks like a ULID
func isULID(s string) bool {
	if len(s) != 26 {
		return false
	}
	// Try to parse as ULID
	_, err := ulid.Parse(s)
	return err == nil
}

// GenerateID generates a new ULID for a project
func GenerateID() string {
	return strings.ToUpper(ulid.Make().String())
}

// ResolvePktPath takes a string like "@my-project/src/index.js" or "./local.txt"
// and resolves it to a full absolute filesystem path.
func ResolvePktPath(input string) (string, error) {
	if !strings.HasPrefix(input, "@") {
		// Standard file path
		return filepath.Abs(input)
	}

	// It's a project path. Extract project name and relative path.
	// Format: @project-name/rest/of/path
	pathWithoutAt := input[1:]
	
	// Find the first separator
	sepIdx := strings.IndexAny(pathWithoutAt, "/\\")
	
	var projectName, relativePath string
	if sepIdx == -1 {
		// Just @project-name
		projectName = pathWithoutAt
		relativePath = ""
	} else {
		projectName = pathWithoutAt[:sepIdx]
		relativePath = pathWithoutAt[sepIdx+1:]
	}

	proj, err := ResolveProject(projectName)
	if err != nil {
		return "", fmt.Errorf("could not resolve project '%s' from path '%s': %w", projectName, input, err)
	}

	return filepath.Join(proj.Path, relativePath), nil
}
