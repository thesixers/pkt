package pm

import "testing"

func TestGetPM(t *testing.T) {
	tests := []struct {
		name        string
		pmName      string
		expectError bool
	}{
		{"pnpm exists", "pnpm", false},
		{"npm exists", "npm", false},
		{"bun exists", "bun", false},
		{"uv exists", "uv", false},
		{"pip exists", "pip", false},
		{"poetry exists", "poetry", false},
		{"go exists", "go", false},
		{"cargo exists", "cargo", false},
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

func TestGet(t *testing.T) {
	tests := []struct {
		name        string
		language    string
		pmName      string
		expectError bool
	}{
		{"javascript pnpm", "javascript", "pnpm", false},
		{"javascript npm", "javascript", "npm", false},
		{"javascript bun", "javascript", "bun", false},
		{"python uv", "python", "uv", false},
		{"python pip", "python", "pip", false},
		{"python poetry", "python", "poetry", false},
		{"go", "go", "go", false},
		{"rust cargo", "rust", "cargo", false},
		{"wrong language", "javascript", "cargo", true},
		{"unknown language", "unknown", "npm", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pm, err := Get(tt.language, tt.pmName)
			if tt.expectError {
				if err == nil {
					t.Errorf("Expected error for %s/%s, got nil", tt.language, tt.pmName)
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error for %s/%s: %v", tt.language, tt.pmName, err)
				}
				if pm == nil {
					t.Errorf("Expected non-nil PackageManager for %s/%s", tt.language, tt.pmName)
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
	if pm.Language() != "javascript" {
		t.Errorf("Expected language 'javascript', got '%s'", pm.Language())
	}
}

func TestNPMName(t *testing.T) {
	pm := &NPM{}
	if pm.Name() != "npm" {
		t.Errorf("Expected name 'npm', got '%s'", pm.Name())
	}
	if pm.Language() != "javascript" {
		t.Errorf("Expected language 'javascript', got '%s'", pm.Language())
	}
}

func TestBunName(t *testing.T) {
	pm := &Bun{}
	if pm.Name() != "bun" {
		t.Errorf("Expected name 'bun', got '%s'", pm.Name())
	}
	if pm.Language() != "javascript" {
		t.Errorf("Expected language 'javascript', got '%s'", pm.Language())
	}
}

func TestUVName(t *testing.T) {
	pm := &UV{}
	if pm.Name() != "uv" {
		t.Errorf("Expected name 'uv', got '%s'", pm.Name())
	}
	if pm.Language() != "python" {
		t.Errorf("Expected language 'python', got '%s'", pm.Language())
	}
}

func TestPipName(t *testing.T) {
	pm := &Pip{}
	if pm.Name() != "pip" {
		t.Errorf("Expected name 'pip', got '%s'", pm.Name())
	}
	if pm.Language() != "python" {
		t.Errorf("Expected language 'python', got '%s'", pm.Language())
	}
}

func TestPoetryName(t *testing.T) {
	pm := &Poetry{}
	if pm.Name() != "poetry" {
		t.Errorf("Expected name 'poetry', got '%s'", pm.Name())
	}
	if pm.Language() != "python" {
		t.Errorf("Expected language 'python', got '%s'", pm.Language())
	}
}

func TestGoModName(t *testing.T) {
	pm := &GoMod{}
	if pm.Name() != "go" {
		t.Errorf("Expected name 'go', got '%s'", pm.Name())
	}
	if pm.Language() != "go" {
		t.Errorf("Expected language 'go', got '%s'", pm.Language())
	}
}

func TestCargoName(t *testing.T) {
	pm := &Cargo{}
	if pm.Name() != "cargo" {
		t.Errorf("Expected name 'cargo', got '%s'", pm.Name())
	}
	if pm.Language() != "rust" {
		t.Errorf("Expected language 'rust', got '%s'", pm.Language())
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

func TestListForLanguage(t *testing.T) {
	// Test JavaScript package managers
	jsPMs := ListForLanguage("javascript")
	// At least npm should be available on most systems
	if len(jsPMs) == 0 {
		t.Log("Warning: no JavaScript package managers available")
	}

	// Test unknown language
	unknownPMs := ListForLanguage("unknown")
	if len(unknownPMs) > 0 {
		t.Error("Expected empty slice for unknown language")
	}
}
