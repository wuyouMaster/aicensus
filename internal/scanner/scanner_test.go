package scanner

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/uwa/aisweep/internal/registry"
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
	if entry.SizeBytes != int64(len("hello")+len("cache data")) {
		t.Fatalf("got %d bytes, want %d", entry.SizeBytes, len("hello")+len("cache data"))
	}
	if entry.FileCount != 2 {
		t.Fatalf("got %d files, want 2", entry.FileCount)
	}
}
