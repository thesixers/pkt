package db

import (
	"database/sql"
	"os"
	"testing"

	_ "github.com/lib/pq"
)

// testDB holds the test database connection
var testDB *sql.DB

// setupTestDB creates a test database connection and runs migrations
func setupTestDB(t *testing.T) {
	t.Helper()

	// Use environment variables for test database, with defaults
	host := getTestEnv("PKT_TEST_DB_HOST", "127.0.0.1")
	port := getTestEnv("PKT_TEST_DB_PORT", "5432")
	user := getTestEnv("PKT_TEST_DB_USER", "pkt_user")
	password := getTestEnv("PKT_TEST_DB_PASSWORD", "yourpassword")
	dbname := getTestEnv("PKT_TEST_DB_NAME", "pkt_test_db")

	// Build connection string for postgres database to create test db
	connStr := "host=" + host + " port=" + port + " user=" + user + " dbname=postgres sslmode=disable"
	if password != "" {
		connStr += " password=" + password
	}

	// Connect to postgres to create test database
	adminDB, err := sql.Open("postgres", connStr)
	if err != nil {
		t.Skipf("Skipping database tests: cannot connect to postgres: %v", err)
	}
	defer func() { _ = adminDB.Close() }()

	// Check if we can actually connect
	if err := adminDB.Ping(); err != nil {
		t.Skipf("Skipping database tests: cannot ping postgres: %v", err)
	}

	// Drop and recreate test database for clean state
	_, _ = adminDB.Exec("DROP DATABASE IF EXISTS " + dbname)
	_, err = adminDB.Exec("CREATE DATABASE " + dbname)
	if err != nil {
		t.Skipf("Skipping database tests: cannot create test database: %v", err)
	}

	// Connect to test database
	testConnStr := "host=" + host + " port=" + port + " user=" + user + " dbname=" + dbname + " sslmode=disable"
	if password != "" {
		testConnStr += " password=" + password
	}

	testDB, err = sql.Open("postgres", testConnStr)
	if err != nil {
		t.Fatalf("Failed to connect to test database: %v", err)
	}

	// Set the global DB to our test database
	DB = testDB

	// Run migrations
	if err := RunMigrations(); err != nil {
		t.Fatalf("Failed to run migrations: %v", err)
	}
}

// teardownTestDB cleans up the test database
func teardownTestDB(t *testing.T) {
	t.Helper()

	if testDB != nil {
		_ = testDB.Close()
	}
	DB = nil
}

func getTestEnv(key, defaultValue string) string {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	return value
}

func TestCreateProject(t *testing.T) {
	setupTestDB(t)
	defer teardownTestDB(t)

	// Test creating a project
	project, err := CreateProject("TEST001", "test-project", "/tmp/test-project", "pnpm")
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
	if project.PackageManager != "pnpm" {
		t.Errorf("Expected PackageManager 'pnpm', got '%s'", project.PackageManager)
	}
}

func TestCreateProjectDuplicatePath(t *testing.T) {
	setupTestDB(t)
	defer teardownTestDB(t)

	// Create first project
	_, err := CreateProject("TEST001", "project1", "/tmp/unique-path", "pnpm")
	if err != nil {
		t.Fatalf("Failed to create first project: %v", err)
	}

	// Try to create second project with same path (should fail due to unique constraint)
	_, err = CreateProject("TEST002", "project2", "/tmp/unique-path", "npm")
	if err == nil {
		t.Error("Expected error when creating project with duplicate path, got nil")
	}
}

func TestGetProjectByID(t *testing.T) {
	setupTestDB(t)
	defer teardownTestDB(t)

	// Create a project
	_, err := CreateProject("GETID001", "get-by-id-test", "/tmp/get-by-id", "npm")
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
}

func TestGetProjectByIDNotFound(t *testing.T) {
	setupTestDB(t)
	defer teardownTestDB(t)

	// Try to get non-existent project
	_, err := GetProjectByID("NONEXISTENT")
	if err == nil {
		t.Error("Expected error when getting non-existent project, got nil")
	}
}

func TestGetProjectByPath(t *testing.T) {
	setupTestDB(t)
	defer teardownTestDB(t)

	// Create a project
	_, err := CreateProject("GETPATH001", "get-by-path-test", "/tmp/get-by-path", "bun")
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
}

func TestGetProjectsByName(t *testing.T) {
	setupTestDB(t)
	defer teardownTestDB(t)

	// Create multiple projects with same name
	_, err := CreateProject("NAME001", "shared-name", "/tmp/path1", "pnpm")
	if err != nil {
		t.Fatalf("Failed to create first project: %v", err)
	}
	_, err = CreateProject("NAME002", "shared-name", "/tmp/path2", "npm")
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

func TestListAllProjects(t *testing.T) {
	setupTestDB(t)
	defer teardownTestDB(t)

	// Initially empty
	projects, err := ListAllProjects()
	if err != nil {
		t.Fatalf("Failed to list projects: %v", err)
	}
	if len(projects) != 0 {
		t.Errorf("Expected 0 projects initially, got %d", len(projects))
	}

	// Create some projects
	_, _ = CreateProject("LIST001", "project1", "/tmp/list1", "pnpm")
	_, _ = CreateProject("LIST002", "project2", "/tmp/list2", "npm")
	_, _ = CreateProject("LIST003", "project3", "/tmp/list3", "bun")

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
	defer teardownTestDB(t)

	// Create a project
	_, err := CreateProject("DEL001", "to-delete", "/tmp/to-delete", "pnpm")
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
	defer teardownTestDB(t)

	// Try to delete non-existent project
	err := DeleteProject("NONEXISTENT")
	if err == nil {
		t.Error("Expected error when deleting non-existent project, got nil")
	}
}

func TestUpdateProjectPM(t *testing.T) {
	setupTestDB(t)
	defer teardownTestDB(t)

	// Create a project
	_, err := CreateProject("UPDATE001", "update-pm-test", "/tmp/update-pm", "pnpm")
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
	_, err := CreateProject("X", "x", "/x", "npm")
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
}
