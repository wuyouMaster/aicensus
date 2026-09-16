package registry

import (
	"os"
	"path/filepath"
	"testing"

	"gopkg.in/yaml.v3"
)

func TestBuiltinScannersHaveUniqueDefinitions(t *testing.T) {
	scanners := BuiltinScanners()
	if len(scanners) == 0 {
		t.Fatal("BuiltinScanners returned no scanners")
	}

	seen := make(map[string]bool, len(scanners))
	for _, scanner := range scanners {
		definition := scanner.Definition()
		if definition.ID == "" {
			t.Fatal("builtin scanner has an empty ID")
		}
		if definition.Label == "" {
			t.Fatalf("builtin scanner %q has an empty label", definition.ID)
		}
		if seen[definition.ID] {
			t.Fatalf("duplicate builtin scanner ID %q", definition.ID)
		}
		seen[definition.ID] = true

		entries, err := scanner.Discover()
		if err != nil {
			t.Fatalf("discover %q: %v", definition.ID, err)
		}
		for _, entry := range entries {
			if entry.Path == "" {
				t.Fatalf("scanner %q returned an empty path", definition.ID)
			}
		}
	}
}

func TestLoadAppliesUserPathOverride(t *testing.T) {
	overridePath := filepath.Join(t.TempDir(), "registry.yaml")
	customPath := filepath.Join(t.TempDir(), "custom-state")
	override := struct {
		Tools []Tool `yaml:"tools"`
	}{Tools: []Tool{{
		ID:    "codex-cli",
		Paths: []Entry{{Path: customPath, Category: "sessions", Risk: "archive"}},
	}}}
	data, err := yaml.Marshal(override)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(overridePath, data, 0o644); err != nil {
		t.Fatal(err)
	}
	t.Setenv("AISWEEP_REGISTRY", overridePath)

	reg, err := Load()
	if err != nil {
		t.Fatal(err)
	}
	var found Tool
	for _, tool := range reg.Tools {
		if tool.ID == "codex-cli" {
			found = tool
			break
		}
	}
	if found.ID == "" {
		t.Fatal("codex-cli was not loaded")
	}
	if len(found.AllPaths()) < 1 || found.AllPaths()[0].Path != customPath {
		t.Fatalf("got overridden paths %#v", found.AllPaths())
	}

	for _, scanner := range reg.Scanners() {
		if scanner.Definition().ID != "codex-cli" {
			continue
		}
		entries, err := scanner.Discover()
		if err != nil {
			t.Fatal(err)
		}
		if len(entries) == 0 || entries[0].Path != customPath {
			t.Fatalf("scanner returned %#v", entries)
		}
		return
	}
	t.Fatal("codex-cli scanner was not loaded")
}
