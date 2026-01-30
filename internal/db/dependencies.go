package db

import (
	"fmt"
	"time"
)

// Dependency represents a project dependency
type Dependency struct {
	ID        int
	ProjectID string
	Name      string
	Version   string
	DepType   string // "prod" or "dev"
	CreatedAt time.Time
}

// SyncDependencies replaces all dependencies for a project with a new set
func SyncDependencies(projectID string, deps map[string]*Dependency) error {
	if DB == nil {
		return fmt.Errorf("database not connected")
	}

	// Start transaction
	tx, err := DB.Begin()
	if err != nil {
		return fmt.Errorf("failed to start transaction: %w", err)
	}
	defer func() { _ = tx.Rollback() }()

	// Delete existing dependencies
	_, err = tx.Exec("DELETE FROM dependencies WHERE project_id = ?", projectID)
	if err != nil {
		return fmt.Errorf("failed to delete existing dependencies: %w", err)
	}

	// Insert new dependencies
	query := `
		INSERT INTO dependencies (project_id, name, version, dep_type, created_at)
		VALUES (?, ?, ?, ?, ?)
	`

	for _, dep := range deps {
		_, err := tx.Exec(query, projectID, dep.Name, dep.Version, dep.DepType, time.Now())
		if err != nil {
			return fmt.Errorf("failed to insert dependency: %w", err)
		}
	}

	// Commit transaction
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

// GetDependencies retrieves all dependencies for a project
func GetDependencies(projectID string) ([]*Dependency, error) {
	if DB == nil {
		return nil, fmt.Errorf("database not connected")
	}

	query := `
		SELECT id, project_id, name, version, dep_type, created_at
		FROM dependencies
		WHERE project_id = ?
		ORDER BY dep_type, name
	`

	rows, err := DB.Query(query, projectID)
	if err != nil {
		return nil, fmt.Errorf("failed to query dependencies: %w", err)
	}
	defer func() { _ = rows.Close() }()

	var dependencies []*Dependency
	for rows.Next() {
		dep := &Dependency{}
		err := rows.Scan(
			&dep.ID,
			&dep.ProjectID,
			&dep.Name,
			&dep.Version,
			&dep.DepType,
			&dep.CreatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan dependency: %w", err)
		}
		dependencies = append(dependencies, dep)
	}

	return dependencies, nil
}
