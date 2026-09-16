package scanner

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/wuyouMaster/aicensus/internal/registry"
	"github.com/wuyouMaster/aicensus/internal/snapshot"
)

type testScanner struct {
	tool    registry.Tool
	entries []registry.Entry
}

func (s testScanner) Definition() registry.Tool { return s.tool }

func (s testScanner) Discover() ([]registry.Entry, error) {
	return append([]registry.Entry(nil), s.entries...), nil
}

func TestScanUsesToolScannerDiscovery(t *testing.T) {
	root := t.TempDir()
	if err := os.WriteFile(filepath.Join(root, "session.log"), []byte("hello"), 0o644); err != nil {
		t.Fatal(err)
	}
	cacheDir := filepath.Join(root, "cache")
	if err := os.Mkdir(cacheDir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(cacheDir, "index.db"), []byte("cache data"), 0o644); err != nil {
		t.Fatal(err)
	}
	nestedDir := filepath.Join(cacheDir, "nested")
	if err := os.Mkdir(nestedDir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(nestedDir, "nested.db"), []byte("nested data"), 0o644); err != nil {
		t.Fatal(err)
	}

	tool := registry.Tool{ID: "test-tool", Label: "Test Tool"}
	reg := registry.New(testScanner{
		tool: tool,
		entries: []registry.Entry{{
			Path:     root,
			Category: "logs",
			Risk:     "archive",
		}},
	})

	snap, err := Scan(reg, nil)
	if err != nil {
		t.Fatal(err)
	}
	if len(snap.Entries) != 1 {
		t.Fatalf("got %d snapshot entries, want 1", len(snap.Entries))
	}
	entry := snap.Entries[0]
	if entry.ToolID != "test-tool" || entry.ToolLabel != "Test Tool" {
		t.Fatalf("got tool %q/%q, want test-tool/Test Tool", entry.ToolID, entry.ToolLabel)
	}
	if entry.Category != "logs" || entry.Risk != "archive" {
		t.Fatalf("got classification %q/%q", entry.Category, entry.Risk)
	}
	if entry.SizeBytes != int64(len("hello")+len("cache data")+len("nested data")) {
		t.Fatalf("got %d bytes, want %d", entry.SizeBytes, len("hello")+len("cache data")+len("nested data"))
	}
	if entry.FileCount != 3 {
		t.Fatalf("got %d files, want 3", entry.FileCount)
	}
	if len(entry.TopSubs) != 1 || entry.TopSubs[0].Files != 2 {
		t.Fatalf("got top subdir files %#v, want one subdir with 2 files", entry.TopSubs)
	}
}

func TestScanDoesNotDoubleCountNestedEntries(t *testing.T) {
	root := t.TempDir()
	if err := os.WriteFile(filepath.Join(root, "root.log"), []byte("root"), 0o644); err != nil {
		t.Fatal(err)
	}

	nestedDir := filepath.Join(root, "sessions")
	if err := os.Mkdir(nestedDir, 0o755); err != nil {
		t.Fatal(err)
	}
	nestedSession := filepath.Join(nestedDir, "session.json")
	if err := os.WriteFile(nestedSession, []byte("session"), 0o644); err != nil {
		t.Fatal(err)
	}

	nestedFile := filepath.Join(root, "thread_history.sqlite")
	if err := os.WriteFile(nestedFile, []byte("history"), 0o644); err != nil {
		t.Fatal(err)
	}

	tool := registry.Tool{ID: "overlap-tool", Label: "Overlap Tool"}
	reg := registry.New(testScanner{
		tool: tool,
		entries: []registry.Entry{
			{Path: root, Category: "cache", Risk: "safe"},
			{Path: nestedDir, Category: "sessions", Risk: "archive"},
			{Path: nestedFile, Category: "logs", Risk: "manual"},
		},
	})

	snap, err := Scan(reg, nil)
	if err != nil {
		t.Fatal(err)
	}
	if len(snap.Entries) != 3 {
		t.Fatalf("got %d snapshot entries, want 3", len(snap.Entries))
	}

	byPath := make(map[string]snapshot.Entry, len(snap.Entries))
	for _, entry := range snap.Entries {
		byPath[entry.Path] = entry
	}

	if got := byPath[root]; got.SizeBytes != int64(len("root")) || got.FileCount != 1 {
		t.Fatalf("parent entry got %d bytes/%d files, want %d bytes/1 file", got.SizeBytes, got.FileCount, len("root"))
	}
	if got := byPath[nestedDir]; got.SizeBytes != int64(len("session")) || got.FileCount != 1 {
		t.Fatalf("nested directory got %d bytes/%d files, want %d bytes/1 file", got.SizeBytes, got.FileCount, len("session"))
	}
	if got := byPath[nestedFile]; got.SizeBytes != int64(len("history")) || got.FileCount != 1 {
		t.Fatalf("nested file got %d bytes/%d files, want %d bytes/1 file", got.SizeBytes, got.FileCount, len("history"))
	}
}

func TestScanHidesParentCoveredByNestedEntries(t *testing.T) {
	root := t.TempDir()
	nestedDir := filepath.Join(root, "sessions")
	if err := os.Mkdir(nestedDir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(nestedDir, "session.json"), []byte("session"), 0o644); err != nil {
		t.Fatal(err)
	}

	tool := registry.Tool{ID: "covered-tool", Label: "Covered Tool"}
	reg := registry.New(testScanner{
		tool: tool,
		entries: []registry.Entry{
			{Path: root, Category: "cache", Risk: "safe"},
			{Path: nestedDir, Category: "sessions", Risk: "archive"},
		},
	})

	snap, err := Scan(reg, nil)
	if err != nil {
		t.Fatal(err)
	}
	if len(snap.Entries) != 1 {
		t.Fatalf("got %d snapshot entries, want only the nested entry", len(snap.Entries))
	}
	if snap.Entries[0].Path != nestedDir {
		t.Fatalf("got %q, want nested path %q", snap.Entries[0].Path, nestedDir)
	}
}

func TestPathKeyNormalizesWindowsCase(t *testing.T) {
	upper := pathKeyFor(filepath.Join(string(filepath.Separator)+"Users", "Test", "AppData"), "windows")
	lower := pathKeyFor(filepath.Join(string(filepath.Separator)+"users", "test", "appdata"), "windows")
	if upper != lower {
		t.Fatalf("Windows path keys differ: %q vs %q", upper, lower)
	}
}

func TestScanAggregatesRelativeDirectoryPaths(t *testing.T) {
	base := t.TempDir()
	root := filepath.Join(base, "storage")
	if err := os.Mkdir(root, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(root, "state.db"), []byte("state"), 0o644); err != nil {
		t.Fatal(err)
	}

	oldWorkingDir, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	if err := os.Chdir(base); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = os.Chdir(oldWorkingDir) })

	reg := registry.New(testScanner{
		tool:    registry.Tool{ID: "relative-tool", Label: "Relative Tool"},
		entries: []registry.Entry{{Path: "storage", Category: "cache", Risk: "safe"}},
	})
	snap, err := Scan(reg, nil)
	if err != nil {
		t.Fatal(err)
	}
	if len(snap.Entries) != 1 {
		t.Fatalf("got %d snapshot entries, want 1", len(snap.Entries))
	}
	if got := snap.Entries[0]; got.SizeBytes != 5 || got.FileCount != 1 || got.Path != "storage" {
		t.Fatalf("got relative entry %#v, want storage with 5 bytes and 1 file", got)
	}
}
