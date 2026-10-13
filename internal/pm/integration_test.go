//go:build integration

package pm

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"
	"time"
)

const integrationCommandTimeout = 45 * time.Second

func useIntegrationCommandTimeout(t *testing.T) {
	t.Helper()
	previous := commandTimeout
	commandTimeout = integrationCommandTimeout
	t.Cleanup(func() { commandTimeout = previous })
}

func newIntegrationProject(t *testing.T, prefix string) string {
	t.Helper()
	dir, err := os.MkdirTemp("", prefix)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = os.RemoveAll(dir) })
	return dir
}

func requirePackageManager(t *testing.T, name string) {
	t.Helper()
	if _, err := exec.LookPath(name); err != nil {
		t.Skipf("%s is not installed", name)
	}
}

func TestNPMAddAndRemoveIntegration(t *testing.T) {
	requirePackageManager(t, "npm")
	useIntegrationCommandTimeout(t)
	dir := newIntegrationProject(t, "pkt-npm-integration-")
	pm := &NPM{}

	if err := pm.Init(dir); err != nil {
		t.Fatal(err)
	}
	if err := pm.Add(dir, []string{"is-odd"}, false); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(filepath.Join(dir, "node_modules", "is-odd")); err != nil {
		t.Fatalf("is-odd was not installed: %v", err)
	}
	if err := pm.Remove(dir, []string{"is-odd"}); err != nil {
		t.Fatal(err)
	}
}

func TestNPMDevDependencyIntegration(t *testing.T) {
	requirePackageManager(t, "npm")
	useIntegrationCommandTimeout(t)
	dir := newIntegrationProject(t, "pkt-npm-dev-integration-")
	pm := &NPM{}

	if err := pm.Init(dir); err != nil {
		t.Fatal(err)
	}
	if err := pm.Add(dir, []string{"is-odd"}, true); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(filepath.Join(dir, "node_modules", "is-odd")); err != nil {
		t.Fatalf("is-odd dev dependency was not installed: %v", err)
	}
}

func TestPNPMAddAndRemoveIntegration(t *testing.T) {
	requirePackageManager(t, "pnpm")
	useIntegrationCommandTimeout(t)
	dir := newIntegrationProject(t, "pkt-pnpm-integration-")

	packageJSON := []byte(`{"name":"pkt-pnpm-integration","version":"1.0.0"}`)
	if err := os.WriteFile(filepath.Join(dir, "package.json"), packageJSON, 0644); err != nil {
		t.Fatalf("write package.json: %v", err)
	}
	pm := &PNPM{}
	if err := pm.Add(dir, []string{"is-odd"}, false); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(filepath.Join(dir, "node_modules", "is-odd")); err != nil {
		t.Fatalf("is-odd was not installed: %v", err)
	}
	if err := pm.Remove(dir, []string{"is-odd"}); err != nil {
		t.Fatal(err)
	}
}
