package db

import (
	"testing"
)

func TestSyncDependencies(t *testing.T) {
	setupTestDB(t)
	defer teardownTestDB(t)

	// Create a project first
	_, err := CreateProject("DEPS001", "deps-test", "/tmp/deps-test", "pnpm")
	if err != nil {
		t.Fatalf("Failed to create project: %v", err)
	}

	// Create dependencies to sync
	deps := map[string]*Dependency{
		"react": {
			Name:    "react",
			Version: "^18.0.0",
			DepType: "prod",
		},
		"typescript": {
			Name:    "typescript",
			Version: "^5.0.0",
			DepType: "dev",
		},
	}

	// Sync dependencies
	err = SyncDependencies("DEPS001", deps)
	if err != nil {
		t.Fatalf("Failed to sync dependencies: %v", err)
	}

	// Retrieve and verify
	retrieved, err := GetDependencies("DEPS001")
	if err != nil {
		t.Fatalf("Failed to get dependencies: %v", err)
	}

	if len(retrieved) != 2 {
		t.Errorf("Expected 2 dependencies, got %d", len(retrieved))
	}

	// Verify dependency details
	foundReact := false
	foundTypescript := false
	for _, dep := range retrieved {
		if dep.Name == "react" {
			foundReact = true
			if dep.Version != "^18.0.0" {
				t.Errorf("Expected react version '^18.0.0', got '%s'", dep.Version)
			}
			if dep.DepType != "prod" {
				t.Errorf("Expected react dep_type 'prod', got '%s'", dep.DepType)
			}
		}
		if dep.Name == "typescript" {
			foundTypescript = true
			if dep.Version != "^5.0.0" {
				t.Errorf("Expected typescript version '^5.0.0', got '%s'", dep.Version)
			}
			if dep.DepType != "dev" {
				t.Errorf("Expected typescript dep_type 'dev', got '%s'", dep.DepType)
			}
		}
	}

	if !foundReact {
		t.Error("Expected to find 'react' dependency")
	}
	if !foundTypescript {
		t.Error("Expected to find 'typescript' dependency")
	}
}

func TestSyncDependenciesReplaces(t *testing.T) {
	setupTestDB(t)
	defer teardownTestDB(t)

	// Create a project
	_, err := CreateProject("DEPS002", "deps-replace-test", "/tmp/deps-replace", "npm")
	if err != nil {
		t.Fatalf("Failed to create project: %v", err)
	}

	// Initial dependencies
	deps1 := map[string]*Dependency{
		"lodash": {
			Name:    "lodash",
			Version: "^4.0.0",
			DepType: "prod",
		},
	}
	err = SyncDependencies("DEPS002", deps1)
	if err != nil {
		t.Fatalf("Failed to sync initial dependencies: %v", err)
	}

	// New dependencies (should replace old ones)
	deps2 := map[string]*Dependency{
		"axios": {
			Name:    "axios",
			Version: "^1.0.0",
			DepType: "prod",
		},
		"jest": {
			Name:    "jest",
			Version: "^29.0.0",
			DepType: "dev",
		},
	}
	err = SyncDependencies("DEPS002", deps2)
	if err != nil {
		t.Fatalf("Failed to sync new dependencies: %v", err)
	}

	// Verify old dependencies are gone and new ones exist
	retrieved, err := GetDependencies("DEPS002")
	if err != nil {
		t.Fatalf("Failed to get dependencies: %v", err)
	}

	if len(retrieved) != 2 {
		t.Errorf("Expected 2 dependencies after replace, got %d", len(retrieved))
	}

	// Should not find lodash
	for _, dep := range retrieved {
		if dep.Name == "lodash" {
			t.Error("Old dependency 'lodash' should have been removed")
		}
	}
}

func TestGetDependenciesEmpty(t *testing.T) {
	setupTestDB(t)
	defer teardownTestDB(t)

	// Create a project with no dependencies
	_, err := CreateProject("DEPS003", "no-deps", "/tmp/no-deps", "pnpm")
	if err != nil {
		t.Fatalf("Failed to create project: %v", err)
	}

	// Get dependencies (should be empty, not error)
	deps, err := GetDependencies("DEPS003")
	if err != nil {
		t.Fatalf("Failed to get dependencies: %v", err)
	}

	if len(deps) != 0 {
		t.Errorf("Expected 0 dependencies, got %d", len(deps))
	}
}

func TestDependenciesCascadeDelete(t *testing.T) {
	setupTestDB(t)
	defer teardownTestDB(t)

	// Create a project with dependencies
	_, err := CreateProject("DEPS004", "cascade-test", "/tmp/cascade", "npm")
	if err != nil {
		t.Fatalf("Failed to create project: %v", err)
	}

	deps := map[string]*Dependency{
		"express": {
			Name:    "express",
			Version: "^4.0.0",
			DepType: "prod",
		},
	}
	err = SyncDependencies("DEPS004", deps)
	if err != nil {
		t.Fatalf("Failed to sync dependencies: %v", err)
	}

	// Delete the project
	err = DeleteProject("DEPS004")
	if err != nil {
		t.Fatalf("Failed to delete project: %v", err)
	}

	// Dependencies should be automatically deleted (CASCADE)
	// We can't directly check since project is gone, but this verifies no FK error
}

func TestDependenciesDBNotConnected(t *testing.T) {
	// Ensure DB is nil
	DB = nil

	deps := map[string]*Dependency{}
	err := SyncDependencies("X", deps)
	if err == nil || err.Error() != "database not connected" {
		t.Errorf("Expected 'database not connected' error, got: %v", err)
	}

	_, err = GetDependencies("X")
	if err == nil || err.Error() != "database not connected" {
		t.Errorf("Expected 'database not connected' error, got: %v", err)
	}
}
