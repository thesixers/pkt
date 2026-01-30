package utils

import (
	"bufio"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/genesix/pkt/internal/db"
)

// ParseRequirementsTxt parses Python requirements.txt
func ParseRequirementsTxt(projectPath string) (map[string]*db.Dependency, error) {
	reqPath := filepath.Join(projectPath, "requirements.txt")
	deps := make(map[string]*db.Dependency)

	file, err := os.Open(reqPath)
	if err != nil {
		if os.IsNotExist(err) {
			return deps, nil
		}
		return nil, err
	}
	defer func() { _ = file.Close() }()

	// Regex to match package==version or package>=version etc
	re := regexp.MustCompile(`^([a-zA-Z0-9_-]+)([=<>!~]+)?(.+)?$`)

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())

		// Skip comments and empty lines
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		// Skip options like -e, --index-url, etc
		if strings.HasPrefix(line, "-") {
			continue
		}

		matches := re.FindStringSubmatch(line)
		if len(matches) >= 2 {
			name := matches[1]
			version := ""
			if len(matches) >= 4 && matches[3] != "" {
				version = matches[2] + matches[3]
			}

			deps[name] = &db.Dependency{
				Name:    name,
				Version: version,
				DepType: "prod",
			}
		}
	}

	return deps, scanner.Err()
}

// ParseGoMod parses Go go.mod file
func ParseGoMod(projectPath string) (map[string]*db.Dependency, error) {
	modPath := filepath.Join(projectPath, "go.mod")
	deps := make(map[string]*db.Dependency)

	data, err := os.ReadFile(modPath)
	if err != nil {
		if os.IsNotExist(err) {
			return deps, nil
		}
		return nil, err
	}

	lines := strings.Split(string(data), "\n")
	inRequire := false

	for _, line := range lines {
		trimmed := strings.TrimSpace(line)

		// Start of require block
		if trimmed == "require (" {
			inRequire = true
			continue
		}

		// End of require block
		if inRequire && trimmed == ")" {
			inRequire = false
			continue
		}

		// Single line require: require github.com/foo/bar v1.0.0
		if strings.HasPrefix(trimmed, "require ") && !strings.Contains(trimmed, "(") {
			parts := strings.Fields(trimmed)
			if len(parts) >= 3 {
				name := parts[1]
				version := parts[2]

				deps[name] = &db.Dependency{
					Name:    name,
					Version: version,
					DepType: "prod",
				}
			}
			continue
		}

		// Inside require block: github.com/foo/bar v1.0.0 // indirect
		if inRequire && trimmed != "" && !strings.HasPrefix(trimmed, "//") {
			parts := strings.Fields(trimmed)
			if len(parts) >= 2 {
				name := parts[0]
				version := parts[1]

				// Go indirect deps are still production deps
				depType := "prod"

				deps[name] = &db.Dependency{
					Name:    name,
					Version: version,
					DepType: depType,
				}
			}
		}
	}

	return deps, nil
}

// ParseCargoToml parses Rust Cargo.toml file
func ParseCargoToml(projectPath string) (map[string]*db.Dependency, error) {
	cargoPath := filepath.Join(projectPath, "Cargo.toml")
	deps := make(map[string]*db.Dependency)

	data, err := os.ReadFile(cargoPath)
	if err != nil {
		if os.IsNotExist(err) {
			return deps, nil
		}
		return nil, err
	}

	lines := strings.Split(string(data), "\n")
	inDeps := false
	inDevDeps := false

	// Simple regex for dependency lines like: name = "version" or name = { version = "x" }
	simpleRe := regexp.MustCompile(`^([a-zA-Z0-9_-]+)\s*=\s*"([^"]+)"`)
	complexRe := regexp.MustCompile(`^([a-zA-Z0-9_-]+)\s*=\s*\{.*version\s*=\s*"([^"]+)"`)

	for _, line := range lines {
		trimmed := strings.TrimSpace(line)

		// Detect section headers
		if trimmed == "[dependencies]" {
			inDeps = true
			inDevDeps = false
			continue
		}
		if trimmed == "[dev-dependencies]" {
			inDeps = false
			inDevDeps = true
			continue
		}
		if strings.HasPrefix(trimmed, "[") {
			inDeps = false
			inDevDeps = false
			continue
		}

		if !inDeps && !inDevDeps {
			continue
		}

		depType := "prod"
		if inDevDeps {
			depType = "dev"
		}

		// Try simple format first
		if matches := simpleRe.FindStringSubmatch(trimmed); len(matches) >= 3 {
			deps[matches[1]] = &db.Dependency{
				Name:    matches[1],
				Version: matches[2],
				DepType: depType,
			}
			continue
		}

		// Try complex format
		if matches := complexRe.FindStringSubmatch(trimmed); len(matches) >= 3 {
			deps[matches[1]] = &db.Dependency{
				Name:    matches[1],
				Version: matches[2],
				DepType: depType,
			}
		}
	}

	return deps, nil
}

// ParseDependencies parses dependencies based on language
func ParseDependencies(projectPath, language string) (map[string]*db.Dependency, error) {
	switch language {
	case "javascript":
		return ParsePackageJSON(projectPath)
	case "python":
		return ParseRequirementsTxt(projectPath)
	case "go":
		return ParseGoMod(projectPath)
	case "rust":
		return ParseCargoToml(projectPath)
	default:
		return make(map[string]*db.Dependency), nil
	}
}
