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

func TestAllPathsForSelectsPlatformSpecificPaths(t *testing.T) {
	tool := Tool{
		Paths:      []Entry{{Path: "{home}/common"}},
		MacOSPaths: []Entry{{Path: "{home}/legacy-mac"}},
		PlatformPaths: map[string][]Entry{
			"darwin":  {{Path: "{config}/mac"}},
			"windows": {{Path: "{config}/windows"}},
			"linux":   {{Path: "{config}/linux"}},
		},
	}

	for goos, want := range map[string]string{
		"darwin":  "{config}/mac",
		"windows": "{config}/windows",
		"linux":   "{config}/linux",
	} {
		paths := tool.AllPathsFor(goos)
		if len(paths) != 2 || paths[1].Path != want {
			t.Fatalf("%s paths = %#v, want common plus %q", goos, paths, want)
		}
	}

	legacy := Tool{Paths: tool.Paths, MacOSPaths: tool.MacOSPaths}
	paths := legacy.AllPathsFor("darwin")
	if len(paths) != 2 || paths[1].Path != "{home}/legacy-mac" {
		t.Fatalf("legacy macOS paths = %#v", paths)
	}
}

func TestDynamicScannersHonorHomeOverrides(t *testing.T) {
	tests := []struct {
		name    string
		scanner ToolScanner
		env     string
		root    string
	}{
		{name: "qoder", scanner: qoderScanner{}, env: "QODER_CONFIG_DIR", root: filepath.Join(t.TempDir(), "qoder")},
		{name: "kiro", scanner: kiroScanner{}, env: "KIRO_HOME", root: filepath.Join(t.TempDir(), "kiro")},
		{name: "cline", scanner: clineScanner{}, env: "CLINE_DATA_DIR", root: filepath.Join(t.TempDir(), "cline")},
		{name: "copilot", scanner: copilotCLIScanner{}, env: "COPILOT_HOME", root: filepath.Join(t.TempDir(), "copilot")},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Setenv(tt.env, tt.root)
			entries, err := tt.scanner.Discover()
			if err != nil {
				t.Fatal(err)
			}
			if len(entries) == 0 || entries[0].Path != tt.root {
				t.Fatalf("got first path %#v, want %q", entries, tt.root)
			}
		})
	}
}

func TestGeminiCLIDiscoverRequiresGeminiSpecificPath(t *testing.T) {
	home := t.TempDir()
	t.Setenv("GEMINI_CLI_HOME", home)
	scanner := geminiCLIScanner{}

	entries, err := scanner.Discover()
	if err != nil {
		t.Fatal(err)
	}
	if len(entries) != 0 {
		t.Fatalf("got paths before a Gemini marker exists: %#v", entries)
	}

	root := filepath.Join(home, ".gemini")
	if err := os.MkdirAll(root, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(root, "settings.json"), []byte("{}"), 0o644); err != nil {
		t.Fatal(err)
	}

	entries, err = scanner.Discover()
	if err != nil {
		t.Fatal(err)
	}
	if len(entries) == 0 || entries[0].Path != root {
		t.Fatalf("got paths %#v, want expanded Gemini root %q", entries, root)
	}
}
