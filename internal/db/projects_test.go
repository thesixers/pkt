package db

import (
	"database/sql"
	"os"
	"path/filepath"
	"testing"

	_ "modernc.org/sqlite"
)

// testDB holds the test database connection
var testDB *sql.DB

// setupTestDB creates a test database connection and runs migrations
func setupTestDB(t *testing.T) {
	t.Helper()

	// Create a temporary directory for test database
	tmpDir, err := os.MkdirTemp("", "pkt-test-db")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}

	// Store temp dir in test context for cleanup
	t.Cleanup(func() {
		if testDB != nil {
			_ = testDB.Close()
		}
		_ = os.RemoveAll(tmpDir)
		DB = nil
	})

	dbPath := filepath.Join(tmpDir, "test.db")

	// Open SQLite database
	testDB, err = sql.Open("sqlite", dbPath)
	if err != nil {
		t.Fatalf("Failed to open test database: %v", err)
	}

	// Enable foreign keys
	if _, err := testDB.Exec("PRAGMA foreign_keys = ON"); err != nil {
		t.Fatalf("Failed to enable foreign keys: %v", err)
	}

	// Set the global DB to our test database
	DB = testDB

	// Run migrations
	if err := RunMigrations(); err != nil {
		t.Fatalf("Failed to run migrations: %v", err)
	}
}

func TestCreateProject(t *testing.T) {
	setupTestDB(t)

	// Test creating a project
	project, err := CreateProject("TEST001", "test-project", "/tmp/test-project", "javascript", "pnpm")
	if err != nil {
		t.Fatalf("Failed to create project: %v", err)
	}

	if project.ID != "TEST001" {
		t.Errorf("Expected ID 'TEST001', got '%s'", project.ID)
	}
	if project.Name != "test-project" {
		t.Errorf("Expected Name 'test-project', got '%s'", project.Name)
	}
	if project.Path != "/tmp/test-project" {
		t.Errorf("Expected Path '/tmp/test-project', got '%s'", project.Path)
	}
	if project.Language != "javascript" {
		t.Errorf("Expected Language 'javascript', got '%s'", project.Language)
	}
	if project.PackageManager != "pnpm" {
		t.Errorf("Expected PackageManager 'pnpm', got '%s'", project.PackageManager)
	}
}

func TestCreateProjectDuplicatePath(t *testing.T) {
	setupTestDB(t)

	// Create first project
	_, err := CreateProject("TEST001", "project1", "/tmp/unique-path", "javascript", "pnpm")
	if err != nil {
		t.Fatalf("Failed to create first project: %v", err)
	}

	// Try to create second project with same path (should fail due to unique constraint)
	_, err = CreateProject("TEST002", "project2", "/tmp/unique-path", "python", "uv")
	if err == nil {
		t.Error("Expected error when creating project with duplicate path, got nil")
	}
}

func TestGetProjectByID(t *testing.T) {
	setupTestDB(t)

	// Create a project
	_, err := CreateProject("GETID001", "get-by-id-test", "/tmp/get-by-id", "go", "go")
	if err != nil {
		t.Fatalf("Failed to create project: %v", err)
	}

	// Retrieve by ID
	project, err := GetProjectByID("GETID001")
	if err != nil {
		t.Fatalf("Failed to get project by ID: %v", err)
	}

	if project.Name != "get-by-id-test" {
		t.Errorf("Expected Name 'get-by-id-test', got '%s'", project.Name)
	}
	if project.Language != "go" {
		t.Errorf("Expected Language 'go', got '%s'", project.Language)
	}
}

func TestGetProjectByIDNotFound(t *testing.T) {
	setupTestDB(t)

	// Try to get non-existent project
	_, err := GetProjectByID("NONEXISTENT")
	if err == nil {
		t.Error("Expected error when getting non-existent project, got nil")
	}
}

func TestGetProjectByPath(t *testing.T) {
	setupTestDB(t)

	// Create a project
	_, err := CreateProject("GETPATH001", "get-by-path-test", "/tmp/get-by-path", "rust", "cargo")
	if err != nil {
		t.Fatalf("Failed to create project: %v", err)
	}

	// Retrieve by path
	project, err := GetProjectByPath("/tmp/get-by-path")
	if err != nil {
		t.Fatalf("Failed to get project by path: %v", err)
	}

	if project.ID != "GETPATH001" {
		t.Errorf("Expected ID 'GETPATH001', got '%s'", project.ID)
	}
	if project.Language != "rust" {
		t.Errorf("Expected Language 'rust', got '%s'", project.Language)
	}
}

func TestGetProjectsByName(t *testing.T) {
	setupTestDB(t)

	// Create multiple projects with same name
	_, err := CreateProject("NAME001", "shared-name", "/tmp/path1", "javascript", "pnpm")
	if err != nil {
		t.Fatalf("Failed to create first project: %v", err)
	}
	_, err = CreateProject("NAME002", "shared-name", "/tmp/path2", "python", "uv")
	if err != nil {
		t.Fatalf("Failed to create second project: %v", err)
	}

	// Get all projects with that name
	projects, err := GetProjectsByName("shared-name")
	if err != nil {
		t.Fatalf("Failed to get projects by name: %v", err)
	}

	if len(projects) != 2 {
		t.Errorf("Expected 2 projects, got %d", len(projects))
	}
}

func TestGetProjectsByLanguage(t *testing.T) {
	setupTestDB(t)

	// Create projects with different languages
	_, _ = CreateProject("LANG001", "js-project1", "/tmp/js1", "javascript", "pnpm")
	_, _ = CreateProject("LANG002", "js-project2", "/tmp/js2", "javascript", "npm")
	_, _ = CreateProject("LANG003", "py-project", "/tmp/py1", "python", "uv")
	_, _ = CreateProject("LANG004", "go-project", "/tmp/go1", "go", "go")

	// Get JavaScript projects
	jsProjects, err := GetProjectsByLanguage("javascript")
	if err != nil {
		t.Fatalf("Failed to get projects by language: %v", err)
	}
	if len(jsProjects) != 2 {
		t.Errorf("Expected 2 JavaScript projects, got %d", len(jsProjects))
	}

	// Get Python projects
	pyProjects, err := GetProjectsByLanguage("python")
	if err != nil {
		t.Fatalf("Failed to get projects by language: %v", err)
	}
	if len(pyProjects) != 1 {
		t.Errorf("Expected 1 Python project, got %d", len(pyProjects))
	}
}

func TestListAllProjects(t *testing.T) {
	setupTestDB(t)

	// Initially empty
	projects, err := ListAllProjects()
	if err != nil {
		t.Fatalf("Failed to list projects: %v", err)
	}
	if len(projects) != 0 {
		t.Errorf("Expected 0 projects initially, got %d", len(projects))
	}

	// Create projects with different languages
	_, _ = CreateProject("LIST001", "project1", "/tmp/list1", "javascript", "pnpm")
	_, _ = CreateProject("LIST002", "project2", "/tmp/list2", "python", "uv")
	_, _ = CreateProject("LIST003", "project3", "/tmp/list3", "go", "go")

	// List all
	projects, err = ListAllProjects()
	if err != nil {
		t.Fatalf("Failed to list projects: %v", err)
	}
	if len(projects) != 3 {
		t.Errorf("Expected 3 projects, got %d", len(projects))
	}
}

func TestDeleteProject(t *testing.T) {
	setupTestDB(t)

	// Create a project
	_, err := CreateProject("DEL001", "to-delete", "/tmp/to-delete", "javascript", "pnpm")
	if err != nil {
		t.Fatalf("Failed to create project: %v", err)
	}

	// Verify it exists
	_, err = GetProjectByID("DEL001")
	if err != nil {
		t.Fatalf("Project should exist before deletion: %v", err)
	}

	// Delete it
	err = DeleteProject("DEL001")
	if err != nil {
		t.Fatalf("Failed to delete project: %v", err)
	}

	// Verify it's gone
	_, err = GetProjectByID("DEL001")
	if err == nil {
		t.Error("Project should not exist after deletion")
	}
}

func TestDeleteProjectNotFound(t *testing.T) {
	setupTestDB(t)

	// Try to delete non-existent project
	err := DeleteProject("NONEXISTENT")
	if err == nil {
		t.Error("Expected error when deleting non-existent project, got nil")
	}
}

func TestUpdateProjectPM(t *testing.T) {
	setupTestDB(t)

	// Create a project
	_, err := CreateProject("UPDATE001", "update-pm-test", "/tmp/update-pm", "javascript", "pnpm")
	if err != nil {
		t.Fatalf("Failed to create project: %v", err)
	}

	// Update package manager
	err = UpdateProjectPM("UPDATE001", "npm")
	if err != nil {
		t.Fatalf("Failed to update package manager: %v", err)
	}

	// Verify the update
	project, err := GetProjectByID("UPDATE001")
	if err != nil {
		t.Fatalf("Failed to get project: %v", err)
	}

	if project.PackageManager != "npm" {
		t.Errorf("Expected PackageManager 'npm', got '%s'", project.PackageManager)
	}
}

func TestDatabaseNotConnected(t *testing.T) {
	// Ensure DB is nil
	DB = nil

	// All operations should return "database not connected" error
	_, err := CreateProject("X", "x", "/x", "javascript", "npm")
	if err == nil || err.Error() != "database not connected" {
		t.Errorf("Expected 'database not connected' error, got: %v", err)
	}

	_, err = GetProjectByID("X")
	if err == nil || err.Error() != "database not connected" {
		t.Errorf("Expected 'database not connected' error, got: %v", err)
	}

	_, err = GetProjectByPath("/x")
	if err == nil || err.Error() != "database not connected" {
		t.Errorf("Expected 'database not connected' error, got: %v", err)
	}

	_, err = GetProjectsByName("x")
	if err == nil || err.Error() != "database not connected" {
		t.Errorf("Expected 'database not connected' error, got: %v", err)
	}

	_, err = GetProjectsByLanguage("javascript")
	if err == nil || err.Error() != "database not connected" {
		t.Errorf("Expected 'database not connected' error, got: %v", err)
	}

	_, err = ListAllProjects()
	if err == nil || err.Error() != "database not connected" {
		t.Errorf("Expected 'database not connected' error, got: %v", err)
	}

	err = DeleteProject("X")
	if err == nil || err.Error() != "database not connected" {
		t.Errorf("Expected 'database not connected' error, got: %v", err)
	}

	err = UpdateProjectPM("X", "npm")
	if err == nil || err.Error() != "database not connected" {
		t.Errorf("Expected 'database not connected' error, got: %v", err)
	}

	err = RenameProject("X", "y")
	if err == nil || err.Error() != "database not connected" {
		t.Errorf("Expected 'database not connected' error, got: %v", err)
	}
}

func TestRenameProject(t *testing.T) {
	setupTestDB(t)

	// Create a project
	_, err := CreateProject("RENAME001", "old-name", "/tmp/rename-test", "javascript", "pnpm")
	if err != nil {
		t.Fatalf("Failed to create project: %v", err)
	}

	// Rename the project
	err = RenameProject("RENAME001", "new-name")
	if err != nil {
		t.Fatalf("Failed to rename project: %v", err)
	}

	// Verify the rename
	project, err := GetProjectByID("RENAME001")
	if err != nil {
		t.Fatalf("Failed to get project: %v", err)
	}

	if project.Name != "new-name" {
		t.Errorf("Expected Name 'new-name', got '%s'", project.Name)
	}

	// Verify old name no longer works
	projects, err := GetProjectsByName("old-name")
	if err != nil {
		t.Fatalf("Failed to query by old name: %v", err)
	}
	if len(projects) != 0 {
		t.Errorf("Expected 0 projects with old name, got %d", len(projects))
	}

	// Verify new name works
	projects, err = GetProjectsByName("new-name")
	if err != nil {
		t.Fatalf("Failed to query by new name: %v", err)
	}
	if len(projects) != 1 {
		t.Errorf("Expected 1 project with new name, got %d", len(projects))
	}
}

func TestRenameProjectNotFound(t *testing.T) {
	setupTestDB(t)

	// Try to rename non-existent project
	err := RenameProject("NONEXISTENT", "new-name")
	if err == nil {
		t.Error("Expected error when renaming non-existent project, got nil")
	}
}
