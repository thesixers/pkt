package utils

import (
	"testing"
)

func TestIsULID(t *testing.T) {
	tests := []struct {
		input    string
		expected bool
	}{
		// Valid ULIDs (26 characters, Crockford base32)
		{"01ARZ3NDEKTSV4RRFFQ69G5FAV", true},
		{"01HQJY9J5C8N7XWQR3MBVPKG1N", true},

		// Invalid - wrong length
		{"", false},
		{"01ARZ3NDEKTSV4RRFFQ69G5FA", false},  // 25 chars
		{"01ARZ3NDEKTSV4RRFFQ69G5FAVV", false}, // 27 chars

		// Invalid - not ULID characters
		{"not-a-ulid-at-all!!!!!!!!", false},
		{"OOOOOOOOOOOOOOOOOOOOOOOOOO", false}, // all O's which when parsed may overflow
	}

	for _, test := range tests {
		result := isULID(test.input)
		if result != test.expected {
			t.Errorf("isULID(%q) = %v, expected %v", test.input, result, test.expected)
		}
	}
}

func TestGenerateID(t *testing.T) {
	// Generate multiple IDs and verify they are valid ULIDs
	ids := make(map[string]bool)

	for i := 0; i < 100; i++ {
		id := GenerateID()

		// Check length
		if len(id) != 26 {
			t.Errorf("Generated ID has wrong length: %d (expected 26)", len(id))
		}

		// Check uniqueness
		if ids[id] {
			t.Errorf("Generated duplicate ID: %s", id)
		}
		ids[id] = true

		// Check it's a valid ULID
		if !isULID(id) {
			t.Errorf("Generated invalid ULID: %s", id)
		}
	}
}

func TestGenerateIDUppercase(t *testing.T) {
	// Verify IDs are uppercase
	id := GenerateID()

	for _, c := range id {
		if c >= 'a' && c <= 'z' {
			t.Errorf("Generated ID contains lowercase character: %s", id)
			break
		}
	}
}
