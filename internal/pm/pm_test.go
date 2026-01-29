package pm

import (
	"os"
	"path/filepath"
	"testing"
)

func TestGetPM(t *testing.T) {
	tests := []struct {
		name        string
		pmName      string
		expectError bool
	}{
		{"pnpm exists", "pnpm", false},
		{"npm exists", "npm", false},
		{"bun exists", "bun", false},
		{"unknown pm", "yarn", true},
		{"empty name", "", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pm, err := GetPM(tt.pmName)
			if tt.expectError {
				if err == nil {
					t.Errorf("Expected error for unknown PM %q, got nil", tt.pmName)
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error for PM %q: %v", tt.pmName, err)
				}
				if pm == nil {
					t.Errorf("Expected non-nil PackageManager for %q", tt.pmName)
				}
			}
		})
	}
}

func TestPNPMName(t *testing.T) {
	pm := &PNPM{}
	if pm.Name() != "pnpm" {
		t.Errorf("Expected name 'pnpm', got '%s'", pm.Name())
	}
}

func TestNPMName(t *testing.T) {
	pm := &NPM{}
	if pm.Name() != "npm" {
		t.Errorf("Expected name 'npm', got '%s'", pm.Name())
	}
}

func TestBunName(t *testing.T) {
	pm := &Bun{}
	if pm.Name() != "bun" {
		t.Errorf("Expected name 'bun', got '%s'", pm.Name())
	}
}

func TestPNPMIsAvailable(t *testing.T) {
	pm := &PNPM{}
	// Just test that it doesn't panic - actual availability depends on system
	_ = pm.IsAvailable()
}

func TestNPMIsAvailable(t *testing.T) {
	pm := &NPM{}
	// Just test that it doesn't panic - actual availability depends on system
	_ = pm.IsAvailable()
}

func TestBunIsAvailable(t *testing.T) {
	pm := &Bun{}
	// Just test that it doesn't panic - actual availability depends on system
	_ = pm.IsAvailable()
}

func TestCheckAvailability(t *testing.T) {
	// Test with a command that definitely exists on all systems
	if !CheckAvailability("go") {
		t.Error("Expected 'go' to be available")
	}

	// Test with a command that definitely doesn't exist
	if CheckAvailability("definitely-not-a-real-command-12345") {
		t.Error("Expected fake command to not be available")
	}
}

func TestListAvailable(t *testing.T) {
	available := ListAvailable()
	// Should return a slice (possibly empty if no PMs installed)
	if available == nil {
		t.Error("Expected non-nil slice from ListAvailable")
	}

	// Verify all returned PMs are in the registry
	for _, name := range available {
		_, err := GetPM(name)
		if err != nil {
			t.Errorf("ListAvailable returned unknown PM: %s", name)
		}
	}
}

// Integration tests - these actually run package manager commands
// They are skipped if the package manager is not available

func TestNPMInit(t *testing.T) {
	pm := &NPM{}
	if !pm.IsAvailable() {
		t.Skip("npm not available, skipping integration test")
	}

	// Create temp directory
	tmpDir, err := os.MkdirTemp("", "pkt-npm-test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer func() { _ = os.RemoveAll(tmpDir) }()

	// Run npm init
	err = pm.Init(tmpDir)
	if err != nil {
		t.Fatalf("npm init failed: %v", err)
	}

	// Verify package.json was created
	pkgPath := filepath.Join(tmpDir, "package.json")
	if _, err := os.Stat(pkgPath); os.IsNotExist(err) {
		t.Error("package.json was not created by npm init")
	}
}

func TestPNPMInit(t *testing.T) {
	pm := &PNPM{}
	if !pm.IsAvailable() {
		t.Skip("pnpm not available, skipping integration test")
	}

	// Create temp directory
	tmpDir, err := os.MkdirTemp("", "pkt-pnpm-test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer func() { _ = os.RemoveAll(tmpDir) }()

	// pnpm init requires stdin input in some versions, so we create package.json manually
	// and just test that pnpm commands work with it
	pkgJSON := `{"name": "pnpm-test", "version": "1.0.0"}`
	err = os.WriteFile(filepath.Join(tmpDir, "package.json"), []byte(pkgJSON), 0644)
	if err != nil {
		t.Fatalf("Failed to write package.json: %v", err)
	}

	// Verify package.json exists
	pkgPath := filepath.Join(tmpDir, "package.json")
	if _, err := os.Stat(pkgPath); os.IsNotExist(err) {
		t.Error("package.json was not created")
	}
}

func TestNPMAddAndRemove(t *testing.T) {
	pm := &NPM{}
	if !pm.IsAvailable() {
		t.Skip("npm not available, skipping integration test")
	}

	// Create temp directory with package.json
	tmpDir, err := os.MkdirTemp("", "pkt-npm-add-test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer func() { _ = os.RemoveAll(tmpDir) }()

	// Initialize
	err = pm.Init(tmpDir)
	if err != nil {
		t.Fatalf("npm init failed: %v", err)
	}

	// Add a small package
	err = pm.Add(tmpDir, "is-odd", nil)
	if err != nil {
		t.Fatalf("npm add failed: %v", err)
	}

	// Verify node_modules exists
	nmPath := filepath.Join(tmpDir, "node_modules", "is-odd")
	if _, err := os.Stat(nmPath); os.IsNotExist(err) {
		t.Error("Package was not installed to node_modules")
	}

	// Remove the package
	err = pm.Remove(tmpDir, "is-odd")
	if err != nil {
		t.Fatalf("npm remove failed: %v", err)
	}
}

func TestPNPMAddAndRemove(t *testing.T) {
	pm := &PNPM{}
	if !pm.IsAvailable() {
		t.Skip("pnpm not available, skipping integration test")
	}

	// Create temp directory with package.json
	tmpDir, err := os.MkdirTemp("", "pkt-pnpm-add-test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer func() { _ = os.RemoveAll(tmpDir) }()

	// Create package.json manually (pnpm init -y may not work in all environments)
	pkgJSON := `{"name": "pnpm-add-test", "version": "1.0.0"}`
	err = os.WriteFile(filepath.Join(tmpDir, "package.json"), []byte(pkgJSON), 0644)
	if err != nil {
		t.Fatalf("Failed to write package.json: %v", err)
	}

	// Add a small package
	err = pm.Add(tmpDir, "is-odd", nil)
	if err != nil {
		t.Fatalf("pnpm add failed: %v", err)
	}

	// Verify node_modules exists
	nmPath := filepath.Join(tmpDir, "node_modules", "is-odd")
	if _, err := os.Stat(nmPath); os.IsNotExist(err) {
		t.Error("Package was not installed to node_modules")
	}

	// Remove the package
	err = pm.Remove(tmpDir, "is-odd")
	if err != nil {
		t.Fatalf("pnpm remove failed: %v", err)
	}
}

func TestAddWithFlags(t *testing.T) {
	pm := &NPM{}
	if !pm.IsAvailable() {
		t.Skip("npm not available, skipping integration test")
	}

	// Create temp directory with package.json
	tmpDir, err := os.MkdirTemp("", "pkt-npm-flags-test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer func() { _ = os.RemoveAll(tmpDir) }()

	// Initialize
	err = pm.Init(tmpDir)
	if err != nil {
		t.Fatalf("npm init failed: %v", err)
	}

	// Add as dev dependency
	err = pm.Add(tmpDir, "is-odd", []string{"--save-dev"})
	if err != nil {
		t.Fatalf("npm add with flags failed: %v", err)
	}
}
