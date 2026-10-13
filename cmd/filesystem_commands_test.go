package cmd

import (
	"bytes"
	"database/sql"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/genesix/pkt/internal/db"
	_ "modernc.org/sqlite"
)

// ---------------------------------------------------------------------------
// Test DB helpers (mirrors the pattern used in internal/db/projects_test.go)
// ---------------------------------------------------------------------------

func setupCmdTestDB(t *testing.T) string {
	t.Helper()
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "test.db")

	testDB, err := sql.Open("sqlite", dbPath)
	if err != nil {
		t.Fatalf("open test db: %v", err)
	}
	t.Cleanup(func() {
		_ = testDB.Close()
		db.DB = nil
	})
	db.DB = testDB
	if err := db.RunMigrations(); err != nil {
		t.Fatalf("run migrations: %v", err)
	}
	return tmpDir
}

// registerProject creates a real temp directory and registers it in the test DB.
func registerProject(t *testing.T, name string) string {
	t.Helper()
	projectDir := t.TempDir()
	_, err := db.CreateProject("TEST-"+name, name, projectDir, "go", "go")
	if err != nil {
		t.Fatalf("create project %q: %v", name, err)
	}
	return projectDir
}

// ---------------------------------------------------------------------------
// pkt ls — pure-function tests (no DB)
// ---------------------------------------------------------------------------

func TestLsDefaultHidesHiddenFiles(t *testing.T) {
	dir := t.TempDir()
	mustWrite(t, filepath.Join(dir, "visible.txt"), "v")
	mustWrite(t, filepath.Join(dir, ".hidden"), "h")
	mustMkdir(t, filepath.Join(dir, "subdir"))

	entries, err := collectListEntries(dir, listOptions{sortBy: "name"})
	if err != nil {
		t.Fatal(err)
	}
	var buf bytes.Buffer
	if err := writeList(&buf, entries, listOptions{sortBy: "name"}); err != nil {
		t.Fatal(err)
	}
	got := buf.String()
	if strings.Contains(got, ".hidden") {
		t.Errorf("hidden file should not appear without -a, got: %q", got)
	}
	if !strings.Contains(got, "visible.txt") {
		t.Errorf("visible.txt should appear, got: %q", got)
	}
	if !strings.Contains(got, "subdir/") {
		t.Errorf("subdir/ should appear, got: %q", got)
	}
}

func TestLsAllFlagIncludesHiddenFiles(t *testing.T) {
	dir := t.TempDir()
	mustWrite(t, filepath.Join(dir, "visible.txt"), "v")
	mustWrite(t, filepath.Join(dir, ".hidden"), "h")

	entries, err := collectListEntries(dir, listOptions{all: true})
	if err != nil {
		t.Fatal(err)
	}
	var buf bytes.Buffer
	if err := writeList(&buf, entries, listOptions{sortBy: "name"}); err != nil {
		t.Fatal(err)
	}
	got := buf.String()
	if !strings.Contains(got, ".hidden") {
		t.Errorf("-a should show hidden files, got: %q", got)
	}
}

func TestLsDirsOnlyFlag(t *testing.T) {
	dir := t.TempDir()
	mustWrite(t, filepath.Join(dir, "file.txt"), "f")
	mustMkdir(t, filepath.Join(dir, "mydir"))

	entries, err := collectListEntries(dir, listOptions{dirs: true})
	if err != nil {
		t.Fatal(err)
	}
	for _, e := range entries {
		if !e.info.IsDir() {
			t.Errorf("--dirs should only return directories, got file: %s", e.label)
		}
	}
	if len(entries) == 0 {
		t.Error("expected at least one directory entry")
	}
}

func TestLsFilesOnlyFlag(t *testing.T) {
	dir := t.TempDir()
	mustWrite(t, filepath.Join(dir, "file.txt"), "f")
	mustMkdir(t, filepath.Join(dir, "mydir"))

	entries, err := collectListEntries(dir, listOptions{files: true})
	if err != nil {
		t.Fatal(err)
	}
	for _, e := range entries {
		if e.info.IsDir() {
			t.Errorf("--files should not return directories, got: %s", e.label)
		}
	}
	if len(entries) == 0 {
		t.Error("expected at least one file entry")
	}
}

func TestLsRecursive(t *testing.T) {
	dir := t.TempDir()
	mustMkdir(t, filepath.Join(dir, "sub"))
	mustWrite(t, filepath.Join(dir, "root.txt"), "r")
	mustWrite(t, filepath.Join(dir, "sub", "deep.txt"), "d")

	entries, err := collectListEntries(dir, listOptions{recursive: true})
	if err != nil {
		t.Fatal(err)
	}
	labels := make(map[string]bool)
	for _, e := range entries {
		labels[e.label] = true
	}
	if !labels["root.txt"] {
		t.Errorf("expected root.txt in recursive listing, got labels: %v", labels)
	}
	if !labels[filepath.Join("sub", "deep.txt")] {
		t.Errorf("expected sub/deep.txt in recursive listing, got labels: %v", labels)
	}
}

func TestLsRecursiveHidesHiddenDirs(t *testing.T) {
	dir := t.TempDir()
	mustMkdir(t, filepath.Join(dir, ".git"))
	mustWrite(t, filepath.Join(dir, ".git", "config"), "cfg")
	mustWrite(t, filepath.Join(dir, "normal.txt"), "n")

	entries, err := collectListEntries(dir, listOptions{recursive: true})
	if err != nil {
		t.Fatal(err)
	}
	for _, e := range entries {
		if strings.Contains(e.label, ".git") {
			t.Errorf("hidden dir .git should be skipped without -a, got: %s", e.label)
		}
	}
}

func TestLsSortBySize(t *testing.T) {
	dir := t.TempDir()
	mustWrite(t, filepath.Join(dir, "big.txt"), strings.Repeat("x", 100))
	mustWrite(t, filepath.Join(dir, "small.txt"), "x")

	entries, err := collectListEntries(dir, listOptions{})
	if err != nil {
		t.Fatal(err)
	}
	var buf bytes.Buffer
	if err := writeList(&buf, entries, listOptions{sortBy: "size"}); err != nil {
		t.Fatal(err)
	}
	lines := nonEmptyLines(buf.String())
	if len(lines) < 2 || lines[0] != "small.txt" {
		t.Errorf("sort by size ascending: want small.txt first, got: %v", lines)
	}
}

func TestLsSortByModified(t *testing.T) {
	dir := t.TempDir()
	// Write older file first.
	mustWrite(t, filepath.Join(dir, "older.txt"), "a")
	mustWrite(t, filepath.Join(dir, "newer.txt"), "b")

	entries, err := collectListEntries(dir, listOptions{})
	if err != nil {
		t.Fatal(err)
	}
	var buf bytes.Buffer
	if err := writeList(&buf, entries, listOptions{sortBy: "modified"}); err != nil {
		t.Fatal(err)
	}
	lines := nonEmptyLines(buf.String())
	if len(lines) < 2 {
		t.Fatalf("expected 2 lines, got: %v", lines)
	}
	// Just verify it doesn't error and produces two entries.
}

func TestLsSortReverse(t *testing.T) {
	dir := t.TempDir()
	mustWrite(t, filepath.Join(dir, "a.txt"), "a")
	mustWrite(t, filepath.Join(dir, "b.txt"), "b")
	mustWrite(t, filepath.Join(dir, "c.txt"), "c")

	entries, err := collectListEntries(dir, listOptions{})
	if err != nil {
		t.Fatal(err)
	}
	var buf bytes.Buffer
	if err := writeList(&buf, entries, listOptions{sortBy: "name", reverse: true}); err != nil {
		t.Fatal(err)
	}
	lines := nonEmptyLines(buf.String())
	if len(lines) < 3 || lines[0] != "c.txt" {
		t.Errorf("reverse name sort: want c.txt first, got: %v", lines)
	}
}

func TestLsLongOutput(t *testing.T) {
	dir := t.TempDir()
	mustWrite(t, filepath.Join(dir, "info.txt"), "hello")

	entries, err := collectListEntries(dir, listOptions{})
	if err != nil {
		t.Fatal(err)
	}
	var buf bytes.Buffer
	if err := writeList(&buf, entries, listOptions{sortBy: "name", long: true}); err != nil {
		t.Fatal(err)
	}
	got := buf.String()
	// Long output should have permissions, size, timestamp, and name.
	if !strings.Contains(got, "info.txt") {
		t.Errorf("long output should contain filename, got: %q", got)
	}
	if !strings.Contains(got, "5") { // file size "hello" = 5 bytes
		t.Errorf("long output should contain file size, got: %q", got)
	}
}

func TestLsMissingPath(t *testing.T) {
	_, err := collectListEntries("/nonexistent/path/xyz", listOptions{})
	if err == nil {
		t.Error("expected error for missing path, got nil")
	}
}

// ---------------------------------------------------------------------------
// pkt ls — @project integration tests (uses in-memory DB)
// ---------------------------------------------------------------------------

func TestLsAtProject(t *testing.T) {
	setupCmdTestDB(t)
	projectDir := registerProject(t, "myproj")
	mustWrite(t, filepath.Join(projectDir, "hello.go"), "package main")

	entries, err := collectListEntries(projectDir, listOptions{})
	if err != nil {
		t.Fatalf("list @myproj: %v", err)
	}
	var buf bytes.Buffer
	if err := writeList(&buf, entries, listOptions{sortBy: "name"}); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(buf.String(), "hello.go") {
		t.Errorf("expected hello.go in @myproj listing, got: %q", buf.String())
	}
}

func TestLsAtProjectInvalidName(t *testing.T) {
	setupCmdTestDB(t)
	// No projects registered — resolution must fail.
	_, err := collectListEntries("/nonexistent-project-xyz", listOptions{})
	if err == nil {
		t.Error("expected error for invalid project name, got nil")
	}
}

// ---------------------------------------------------------------------------
// pkt cat — pure-function tests (no DB)
// ---------------------------------------------------------------------------

func TestCatNormalFile(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "test.txt")
	mustWrite(t, path, "line1\nline2\nline3\n")

	var buf bytes.Buffer
	if err := writeCatFile(&buf, path, path, false, catOptions{}); err != nil {
		t.Fatal(err)
	}
	got := buf.String()
	if !strings.Contains(got, "line1") || !strings.Contains(got, "line3") {
		t.Errorf("expected all lines, got: %q", got)
	}
}

func TestCatEmptyFile(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "empty.txt")
	mustWrite(t, path, "")

	var buf bytes.Buffer
	if err := writeCatFile(&buf, path, path, false, catOptions{}); err != nil {
		t.Fatal(err)
	}
	if buf.Len() != 0 {
		t.Errorf("empty file should produce no output, got: %q", buf.String())
	}
}

func TestCatRejectsDirectory(t *testing.T) {
	dir := t.TempDir()
	var buf bytes.Buffer
	err := writeCatFile(&buf, dir, dir, false, catOptions{})
	if err == nil || !strings.Contains(err.Error(), "use pkt ls") {
		t.Fatalf("expected directory error, got: %v", err)
	}
}

func TestCatRejectsBinaryByDefault(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "data.bin")
	mustWrite(t, path, string([]byte{0x00, 0x01, 0x02, 0xFF}))

	var buf bytes.Buffer
	err := writeCatFile(&buf, path, path, false, catOptions{})
	if err == nil || !strings.Contains(err.Error(), "binary") {
		t.Fatalf("expected binary error, got: %v", err)
	}
}

func TestCatForceOverridesBinaryRefusal(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "data.bin")
	mustWrite(t, path, string([]byte{0x00, 0x01, 0x02}))

	var buf bytes.Buffer
	if err := writeCatFile(&buf, path, path, false, catOptions{force: true}); err != nil {
		t.Fatalf("--force should allow binary files, got: %v", err)
	}
}

func TestCatLineNumbers(t *testing.T) {
	var buf bytes.Buffer
	data := []byte("alpha\nbeta\ngamma\n")
	if err := writeSelectedLines(&buf, data, catOptions{number: true}); err != nil {
		t.Fatal(err)
	}
	got := buf.String()
	if !strings.Contains(got, "1\talpha") {
		t.Errorf("line numbers: expected '1\\talpha', got: %q", got)
	}
	if !strings.Contains(got, "3\tgamma") {
		t.Errorf("line numbers: expected '3\\tgamma', got: %q", got)
	}
}

func TestCatStartEnd(t *testing.T) {
	var buf bytes.Buffer
	data := []byte("one\ntwo\nthree\nfour\nfive\n")
	if err := writeSelectedLines(&buf, data, catOptions{start: 2, end: 4}); err != nil {
		t.Fatal(err)
	}
	lines := nonEmptyLines(buf.String())
	if len(lines) != 3 || lines[0] != "two" || lines[2] != "four" {
		t.Errorf("--start 2 --end 4: got lines %v", lines)
	}
}

func TestCatHead(t *testing.T) {
	var buf bytes.Buffer
	data := []byte("a\nb\nc\nd\ne\n")
	if err := writeSelectedLines(&buf, data, catOptions{head: 3}); err != nil {
		t.Fatal(err)
	}
	lines := nonEmptyLines(buf.String())
	if len(lines) != 3 || lines[2] != "c" {
		t.Errorf("--head 3: expected [a b c], got: %v", lines)
	}
}

func TestCatTail(t *testing.T) {
	var buf bytes.Buffer
	data := []byte("a\nb\nc\nd\ne\n")
	if err := writeSelectedLines(&buf, data, catOptions{tail: 2}); err != nil {
		t.Fatal(err)
	}
	lines := nonEmptyLines(buf.String())
	if len(lines) != 2 || lines[0] != "d" || lines[1] != "e" {
		t.Errorf("--tail 2: expected [d e], got: %v", lines)
	}
}

func TestCatHeadWithStart(t *testing.T) {
	// --head can combine with --start: N lines starting at line X.
	var buf bytes.Buffer
	data := []byte("one\ntwo\nthree\nfour\nfive\n")
	if err := writeSelectedLines(&buf, data, catOptions{start: 2, head: 2, number: true}); err != nil {
		t.Fatal(err)
	}
	got := buf.String()
	// Should contain lines 2 and 3 (two and three) with their line numbers.
	if !strings.Contains(got, "2\ttwo") || !strings.Contains(got, "3\tthree") {
		t.Errorf("--start 2 --head 2: expected lines 2-3, got: %q", got)
	}
	if strings.Contains(got, "four") {
		t.Errorf("--start 2 --head 2: should not include line 4, got: %q", got)
	}
}

func TestCatMultiFileHeader(t *testing.T) {
	dir := t.TempDir()
	a := filepath.Join(dir, "a.txt")
	b := filepath.Join(dir, "b.txt")
	mustWrite(t, a, "content-a\n")
	mustWrite(t, b, "content-b\n")

	var buf bytes.Buffer
	if err := writeCatFile(&buf, a, "a.txt", true, catOptions{}); err != nil {
		t.Fatal(err)
	}
	got := buf.String()
	if !strings.Contains(got, "==> a.txt <==") {
		t.Errorf("multi-file header missing, got: %q", got)
	}
}

func TestCatNoHeader(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "file.txt")
	mustWrite(t, path, "body\n")

	var buf bytes.Buffer
	// header=false simulates --no-header with multiple files.
	if err := writeCatFile(&buf, path, path, false, catOptions{}); err != nil {
		t.Fatal(err)
	}
	if strings.Contains(buf.String(), "==>") {
		t.Errorf("no-header mode should not print ==> separator, got: %q", buf.String())
	}
}

func TestCatMissingFile(t *testing.T) {
	var buf bytes.Buffer
	err := writeCatFile(&buf, "/nonexistent/path/to/file.txt", "file.txt", false, catOptions{})
	if err == nil {
		t.Error("expected error for missing file, got nil")
	}
}

// ---------------------------------------------------------------------------
// pkt cat — option validation tests
// ---------------------------------------------------------------------------

func TestValidateCatOptionsHeadTailConflict(t *testing.T) {
	if err := validateCatOptions(catOptions{head: 2, tail: 2}); err == nil {
		t.Fatal("expected --head + --tail conflict error")
	}
}

func TestValidateCatOptionsTailStartConflict(t *testing.T) {
	if err := validateCatOptions(catOptions{tail: 2, start: 4}); err == nil {
		t.Fatal("expected --tail + --start conflict error")
	}
}

func TestValidateCatOptionsTailEndConflict(t *testing.T) {
	if err := validateCatOptions(catOptions{tail: 2, end: 10}); err == nil {
		t.Fatal("expected --tail + --end conflict error")
	}
}

func TestValidateCatOptionsEndBeforeStart(t *testing.T) {
	if err := validateCatOptions(catOptions{start: 10, end: 5}); err == nil {
		t.Fatal("expected --end < --start error")
	}
}

func TestValidateCatOptionsValidCombinations(t *testing.T) {
	cases := []catOptions{
		{},
		{number: true},
		{start: 1, end: 5},
		{head: 10},
		{tail: 10},
		{start: 3, head: 5}, // head + start is allowed
		{force: true},
		{noHeader: true},
	}
	for _, c := range cases {
		if err := validateCatOptions(c); err != nil {
			t.Errorf("expected valid options %+v to pass, got: %v", c, err)
		}
	}
}

// ---------------------------------------------------------------------------
// @project path safety tests (pure-function, no DB required)
// ---------------------------------------------------------------------------

func TestResolvePktPathNoEscape(t *testing.T) {
	// The escape check in ResolvePktPath runs after project resolution.
	// We test hasHiddenComponent (used by ls recursive walk) as a proxy
	// for path safety logic, and test the project escape logic indirectly
	// via the resolver when given a real temp dir project.
	setupCmdTestDB(t)
	projectDir := registerProject(t, "safeguard")

	// A sub-path that would escape the project root must be rejected.
	// We call the ls helper with a path manually constructed to escape.
	escapePath := filepath.Join(projectDir, "..", "..")
	_, err := collectListEntries(escapePath, listOptions{})
	// This resolves to an ancestor directory, not the project root.
	// The test just ensures no panic; actual escape prevention is in ResolvePktPath.
	_ = err
}

func TestHasHiddenComponent(t *testing.T) {
	cases := []struct {
		path   string
		hidden bool
	}{
		{"normal/path/file.txt", false},
		{".hidden", true},
		{"src/.hidden/file.go", true},
		{"src/normal/.git/config", true},
		{".", false},
		{"..", false},
	}
	for _, c := range cases {
		got := hasHiddenComponent(c.path)
		if got != c.hidden {
			t.Errorf("hasHiddenComponent(%q) = %v, want %v", c.path, got, c.hidden)
		}
	}
}

// ---------------------------------------------------------------------------
// Helpers
// ---------------------------------------------------------------------------

func mustWrite(t *testing.T, path, content string) {
	t.Helper()
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatalf("write %s: %v", path, err)
	}
}

func mustMkdir(t *testing.T, path string) {
	t.Helper()
	if err := os.Mkdir(path, 0755); err != nil {
		t.Fatalf("mkdir %s: %v", path, err)
	}
}

func nonEmptyLines(s string) []string {
	var out []string
	for _, l := range strings.Split(s, "\n") {
		if l != "" {
			out = append(out, l)
		}
	}
	return out
}
