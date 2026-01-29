package utils

import (
	"os"
	"path/filepath"
	"testing"
)

func TestParsePackageJSON(t *testing.T) {
	// Create a temporary directory
	tmpDir, err := os.MkdirTemp("", "pkt-test-parse")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer func() { _ = os.RemoveAll(tmpDir) }()

	// Create a package.json file
	pkgJSON := `{
		"name": "test-package",
		"version": "1.0.0",
		"dependencies": {
			"react": "^18.0.0",
			"express": "^4.18.0"
		},
		"devDependencies": {
			"typescript": "^5.0.0",
			"jest": "^29.0.0"
		}
	}`

	pkgPath := filepath.Join(tmpDir, "package.json")
	err = os.WriteFile(pkgPath, []byte(pkgJSON), 0644)
	if err != nil {
		t.Fatalf("Failed to write package.json: %v", err)
	}

	// Parse it
	deps, err := ParsePackageJSON(tmpDir)
	if err != nil {
		t.Fatalf("Failed to parse package.json: %v", err)
	}

	// Verify dependencies
	if len(deps) != 4 {
		t.Errorf("Expected 4 dependencies, got %d", len(deps))
	}

	// Check prod dependencies
	if deps["react"] == nil {
		t.Error("Expected 'react' dependency")
	} else {
		if deps["react"].Version != "^18.0.0" {
			t.Errorf("Expected react version '^18.0.0', got '%s'", deps["react"].Version)
		}
		if deps["react"].DepType != "prod" {
			t.Errorf("Expected react dep_type 'prod', got '%s'", deps["react"].DepType)
		}
	}

	// Check dev dependencies
	if deps["typescript"] == nil {
		t.Error("Expected 'typescript' dependency")
	} else {
		if deps["typescript"].DepType != "dev" {
			t.Errorf("Expected typescript dep_type 'dev', got '%s'", deps["typescript"].DepType)
		}
	}
}

func TestParsePackageJSONNotFound(t *testing.T) {
	// Parse a directory without package.json
	tmpDir, err := os.MkdirTemp("", "pkt-test-nopackage")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer func() { _ = os.RemoveAll(tmpDir) }()

	// Should return empty map, not error
	deps, err := ParsePackageJSON(tmpDir)
	if err != nil {
		t.Fatalf("Expected no error for missing package.json, got: %v", err)
	}
	if len(deps) != 0 {
		t.Errorf("Expected 0 dependencies, got %d", len(deps))
	}
}

func TestParsePackageJSONInvalid(t *testing.T) {
	// Create a temp directory with invalid JSON
	tmpDir, err := os.MkdirTemp("", "pkt-test-invalid")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer func() { _ = os.RemoveAll(tmpDir) }()

	// Write invalid JSON
	pkgPath := filepath.Join(tmpDir, "package.json")
	err = os.WriteFile(pkgPath, []byte("{ invalid json }"), 0644)
	if err != nil {
		t.Fatalf("Failed to write invalid package.json: %v", err)
	}

	// Should return error
	_, err = ParsePackageJSON(tmpDir)
	if err == nil {
		t.Error("Expected error for invalid JSON, got nil")
	}
}

func TestCreatePackageJSON(t *testing.T) {
	// Create a temp directory
	tmpDir, err := os.MkdirTemp("", "pkt-test-create")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer func() { _ = os.RemoveAll(tmpDir) }()

	// Create package.json
	err = CreatePackageJSON(tmpDir, "my-new-project")
	if err != nil {
		t.Fatalf("Failed to create package.json: %v", err)
	}

	// Verify it exists
	pkgPath := filepath.Join(tmpDir, "package.json")
	if _, err := os.Stat(pkgPath); os.IsNotExist(err) {
		t.Error("package.json was not created")
	}

	// Parse and verify contents
	deps, err := ParsePackageJSON(tmpDir)
	if err != nil {
		t.Fatalf("Failed to parse created package.json: %v", err)
	}

	// Should have no dependencies
	if len(deps) != 0 {
		t.Errorf("Expected 0 dependencies in new package.json, got %d", len(deps))
	}
}

func TestCreatePackageJSONAlreadyExists(t *testing.T) {
	// Create a temp directory
	tmpDir, err := os.MkdirTemp("", "pkt-test-exists")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer func() { _ = os.RemoveAll(tmpDir) }()

	// Create an existing package.json
	existingContent := `{"name": "existing", "version": "2.0.0"}`
	pkgPath := filepath.Join(tmpDir, "package.json")
	err = os.WriteFile(pkgPath, []byte(existingContent), 0644)
	if err != nil {
		t.Fatalf("Failed to write existing package.json: %v", err)
	}

	// Try to create (should not overwrite)
	err = CreatePackageJSON(tmpDir, "new-name")
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	// Verify original content is preserved
	content, err := os.ReadFile(pkgPath)
	if err != nil {
		t.Fatalf("Failed to read package.json: %v", err)
	}
	if string(content) != existingContent {
		t.Error("Existing package.json should not have been overwritten")
	}
}

func TestRewriteScripts(t *testing.T) {
	// Create a temp directory
	tmpDir, err := os.MkdirTemp("", "pkt-test-rewrite")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer func() { _ = os.RemoveAll(tmpDir) }()

	// Create package.json with npm scripts
	pkgJSON := `{
  "name": "test-package",
  "version": "1.0.0",
  "scripts": {
    "build": "npm run compile",
    "test": "npm test",
    "start": "npm run serve"
  }
}`
	pkgPath := filepath.Join(tmpDir, "package.json")
	err = os.WriteFile(pkgPath, []byte(pkgJSON), 0644)
	if err != nil {
		t.Fatalf("Failed to write package.json: %v", err)
	}

	// Rewrite scripts to use pnpm
	err = RewriteScripts(tmpDir, "pnpm")
	if err != nil {
		t.Fatalf("Failed to rewrite scripts: %v", err)
	}

	// Read and verify
	content, err := os.ReadFile(pkgPath)
	if err != nil {
		t.Fatalf("Failed to read package.json: %v", err)
	}

	// Check that npm was replaced with pnpm
	contentStr := string(content)
	if !contains(contentStr, "pnpm run compile") {
		t.Error("Expected 'npm run compile' to be rewritten to 'pnpm run compile'")
	}
	if !contains(contentStr, "pnpm test") {
		t.Error("Expected 'npm test' to be rewritten to 'pnpm test'")
	}
}

func TestRewriteScriptsNoScripts(t *testing.T) {
	// Create a temp directory
	tmpDir, err := os.MkdirTemp("", "pkt-test-noscripts")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer func() { _ = os.RemoveAll(tmpDir) }()

	// Create package.json without scripts
	pkgJSON := `{"name": "test", "version": "1.0.0"}`
	pkgPath := filepath.Join(tmpDir, "package.json")
	err = os.WriteFile(pkgPath, []byte(pkgJSON), 0644)
	if err != nil {
		t.Fatalf("Failed to write package.json: %v", err)
	}

	// Should not error
	err = RewriteScripts(tmpDir, "pnpm")
	if err != nil {
		t.Errorf("Unexpected error for package with no scripts: %v", err)
	}
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > 0 && containsHelper(s, substr))
}

func containsHelper(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
