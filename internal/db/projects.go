package db

import (
	"database/sql"
	"fmt"
	"time"
)

// Project represents a tracked project
type Project struct {
	ID             string
	Name           string
	Path           string
	PackageManager string
	CreatedAt      time.Time
}

// CreateProject inserts a new project into the database
func CreateProject(id, name, path, pm string) (*Project, error) {
	if DB == nil {
		return nil, fmt.Errorf("database not connected")
	}

	project := &Project{
		ID:             id,
		Name:           name,
		Path:           path,
		PackageManager: pm,
		CreatedAt:      time.Now(),
	}

	query := `
		INSERT INTO projects (id, name, path, package_manager, created_at)
		VALUES ($1, $2, $3, $4, $5)
	`

	_, err := DB.Exec(query, project.ID, project.Name, project.Path, project.PackageManager, project.CreatedAt)
	if err != nil {
		return nil, fmt.Errorf("failed to create project: %w", err)
	}

	return project, nil
}

// GetProjectByID retrieves a project by its ID
func GetProjectByID(id string) (*Project, error) {
	if DB == nil {
		return nil, fmt.Errorf("database not connected")
	}

	project := &Project{}
	query := `SELECT id, name, path, package_manager, created_at FROM projects WHERE id = $1`

	err := DB.QueryRow(query, id).Scan(
		&project.ID,
		&project.Name,
		&project.Path,
		&project.PackageManager,
		&project.CreatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("project not found")
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get project: %w", err)
	}

	return project, nil
}

// GetProjectByPath retrieves a project by its filesystem path
func GetProjectByPath(path string) (*Project, error) {
	if DB == nil {
		return nil, fmt.Errorf("database not connected")
	}

	project := &Project{}
	query := `SELECT id, name, path, package_manager, created_at FROM projects WHERE path = $1`

	err := DB.QueryRow(query, path).Scan(
		&project.ID,
		&project.Name,
		&project.Path,
		&project.PackageManager,
		&project.CreatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("project not found")
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get project: %w", err)
	}

	return project, nil
}

// GetProjectsByName retrieves all projects with a given name (duplicates allowed)
func GetProjectsByName(name string) ([]*Project, error) {
	if DB == nil {
		return nil, fmt.Errorf("database not connected")
	}

	query := `SELECT id, name, path, package_manager, created_at FROM projects WHERE name = $1 ORDER BY created_at DESC`

	rows, err := DB.Query(query, name)
	if err != nil {
		return nil, fmt.Errorf("failed to query projects: %w", err)
	}
	defer func() { _ = rows.Close() }()

	var projects []*Project
	for rows.Next() {
		project := &Project{}
		err := rows.Scan(
			&project.ID,
			&project.Name,
			&project.Path,
			&project.PackageManager,
			&project.CreatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan project: %w", err)
		}
		projects = append(projects, project)
	}

	return projects, nil
}

// ListAllProjects retrieves all projects from the database
func ListAllProjects() ([]*Project, error) {
	if DB == nil {
		return nil, fmt.Errorf("database not connected")
	}

	query := `SELECT id, name, path, package_manager, created_at FROM projects ORDER BY created_at DESC`

	rows, err := DB.Query(query)
	if err != nil {
		return nil, fmt.Errorf("failed to query projects: %w", err)
	}
	defer func() { _ = rows.Close() }()

	var projects []*Project
	for rows.Next() {
		project := &Project{}
		err := rows.Scan(
			&project.ID,
			&project.Name,
			&project.Path,
			&project.PackageManager,
			&project.CreatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan project: %w", err)
		}
		projects = append(projects, project)
	}

	return projects, nil
}

// DeleteProject removes a project from the database
func DeleteProject(id string) error {
	if DB == nil {
		return fmt.Errorf("database not connected")
	}

	query := `DELETE FROM projects WHERE id = $1`
	result, err := DB.Exec(query, id)
	if err != nil {
		return fmt.Errorf("failed to delete project: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to check rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("project not found")
	}

	return nil
}

// UpdateProjectPM updates the package manager for a project
func UpdateProjectPM(id, pm string) error {
	if DB == nil {
		return fmt.Errorf("database not connected")
	}

	query := `UPDATE projects SET package_manager = $1 WHERE id = $2`
	result, err := DB.Exec(query, pm, id)
	if err != nil {
		return fmt.Errorf("failed to update project: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to check rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("project not found")
	}

	return nil
}

// RenameProject updates the name of a project
func RenameProject(id, newName string) error {
	if DB == nil {
		return fmt.Errorf("database not connected")
	}

	query := `UPDATE projects SET name = $1 WHERE id = $2`
	result, err := DB.Exec(query, newName, id)
	if err != nil {
		return fmt.Errorf("failed to rename project: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to check rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("project not found")
	}

	return nil
}
