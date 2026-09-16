package registry

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"

	"github.com/wuyouMaster/aicensus/internal/platform"
	"gopkg.in/yaml.v3"
)

// Entry describes one storage location and its cleanup classification.
type Entry struct {
	Path     string `yaml:"path"`
	Category string `yaml:"category"`
	Risk     string `yaml:"risk"`
	Note     string `yaml:"note,omitempty"`
}

// Tool is the user-facing metadata for one AI tool. Paths are kept here for
// the path and doctor commands; scanning uses ToolScanner.Discover so a tool
// can resolve paths dynamically when needed.
type Tool struct {
	ID            string             `yaml:"id"`
	Label         string             `yaml:"label"`
	Homepage      string             `yaml:"homepage,omitempty"`
	Paths         []Entry            `yaml:"paths,omitempty"`
	MacOSPaths    []Entry            `yaml:"macos_paths,omitempty"`
	PlatformPaths map[string][]Entry `yaml:"platform_paths,omitempty"`
}

// PlatformContext is exported through registry so scanner contributors do not
// need to know where the shared platform path implementation lives.
type PlatformContext = platform.Context

func CurrentPlatformContext() (PlatformContext, error) {
	return platform.Current()
}

// ToolScanner is the extension point for AI-tool-specific path discovery.
// Implementations should keep tool-specific knowledge here and leave
// filesystem traversal and aggregation to internal/scanner.
type ToolScanner interface {
	Definition() Tool
	Discover() ([]Entry, error)
}

type Registry struct {
	Tools    []Tool `yaml:"tools"`
	scanners []ToolScanner
}

// New creates a registry from scanner implementations. It is useful for
// embedding aisweep or testing a scanner without loading user configuration.
func New(scanners ...ToolScanner) *Registry {
	reg := &Registry{scanners: append([]ToolScanner(nil), scanners...)}
	for _, scanner := range scanners {
		reg.Tools = append(reg.Tools, scanner.Definition())
	}
	return reg
}

// Scanners returns the resolved scanners used by the generic scan pipeline.
func (r *Registry) Scanners() []ToolScanner {
	if len(r.scanners) == 0 {
		out := make([]ToolScanner, 0, len(r.Tools))
		for _, tool := range r.Tools {
			out = append(out, staticScanner{tool: tool})
		}
		return out
	}
	return append([]ToolScanner(nil), r.scanners...)
}

func (t Tool) AllPaths() []Entry {
	return t.AllPathsFor(runtime.GOOS)
}

func (t Tool) AllPathsFor(goos string) []Entry {
	paths := append([]Entry{}, t.Paths...)
	if goos == "darwin" && len(t.PlatformPaths["darwin"]) == 0 {
		paths = append(paths, t.MacOSPaths...)
	}
	paths = append(paths, t.PlatformPaths[goos]...)
	return paths
}

// Load builds the registry from Go implementations and then applies the
// optional user YAML override. Built-in tool definitions therefore live next
// to their discovery implementations, while local path customization remains
// possible without rebuilding the binary.
func Load() (*Registry, error) {
	override, err := loadOverride()
	if err != nil {
		return nil, err
	}
	overrides := map[string]Tool{}
	if override != nil {
		for _, t := range override.Tools {
			overrides[t.ID] = t
		}
	}

	reg := &Registry{}
	seen := map[string]bool{}
	for _, source := range BuiltinScanners() {
		base := source.Definition()
		over, hasOverride := overrides[base.ID]
		resolved := base
		if hasOverride {
			resolved = mergeTool(base, over)
		}

		var resolvedScanner ToolScanner = delegatedScanner{source: source, tool: resolved}
		if hasOverride && hasPathOverride(over) {
			resolvedScanner = staticScanner{tool: resolved}
		}
		entries, err := resolvedScanner.Discover()
		if err != nil {
			return nil, fmt.Errorf("discover %s: %w", base.ID, err)
		}
		resolved.Paths = entries
		resolved.MacOSPaths = nil
		resolved.PlatformPaths = nil
		reg.Tools = append(reg.Tools, resolved)
		reg.scanners = append(reg.scanners, resolvedScanner)
		seen[base.ID] = true
	}

	// A user override may also introduce a completely new tool. It uses the
	// same static scanner contract and can later be moved into its own Go
	// implementation through a normal pull request.
	if override != nil {
		for _, t := range override.Tools {
			if seen[t.ID] {
				continue
			}
			source := staticScanner{tool: t}
			entries, err := source.Discover()
			if err != nil {
				return nil, fmt.Errorf("discover %s: %w", t.ID, err)
			}
			t.Paths = entries
			t.MacOSPaths = nil
			t.PlatformPaths = nil
			reg.Tools = append(reg.Tools, t)
			reg.scanners = append(reg.scanners, source)
		}
	}
	return reg, nil
}

func loadOverride() (*Registry, error) {
	path, err := UserOverridePath()
	if err != nil {
		return nil, nil
	}
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, fmt.Errorf("read user registry %s: %w", path, err)
	}
	override := &Registry{}
	if err := yaml.Unmarshal(data, override); err != nil {
		return nil, fmt.Errorf("parse user registry %s: %w", path, err)
	}
	return override, nil
}

func mergeTool(base, override Tool) Tool {
	if len(override.Paths) > 0 {
		base.Paths = override.Paths
	}
	if len(override.MacOSPaths) > 0 {
		base.MacOSPaths = override.MacOSPaths
		if base.PlatformPaths != nil {
			delete(base.PlatformPaths, "darwin")
		}
	}
	if len(override.PlatformPaths) > 0 {
		if base.PlatformPaths == nil {
			base.PlatformPaths = map[string][]Entry{}
		}
		for goos, entries := range override.PlatformPaths {
			base.PlatformPaths[goos] = entries
			if goos == "darwin" {
				base.MacOSPaths = nil
			}
		}
	}
	if override.Label != "" {
		base.Label = override.Label
	}
	if override.Homepage != "" {
		base.Homepage = override.Homepage
	}
	return base
}

func hasPathOverride(t Tool) bool {
	return len(t.Paths) > 0 || len(t.MacOSPaths) > 0 || len(t.PlatformPaths) > 0
}

type delegatedScanner struct {
	source ToolScanner
	tool   Tool
}

func (s delegatedScanner) Definition() Tool { return s.tool }

func (s delegatedScanner) Discover() ([]Entry, error) {
	return s.source.Discover()
}

type staticScanner struct {
	tool Tool
}

func (s staticScanner) Definition() Tool { return s.tool }

func (s staticScanner) Discover() ([]Entry, error) {
	return expandList(s.tool.AllPaths()), nil
}

func expandList(es []Entry) []Entry {
	ctx, err := platform.Current()
	if err != nil {
		return nil
	}
	out := make([]Entry, 0, len(es))
	for _, e := range es {
		e.Path = platform.ExpandPath(e.Path, ctx)
		if e.Path == "" {
			continue
		}
		out = append(out, e)
	}
	return out
}

func expand(p string) string {
	ctx, err := platform.Current()
	if err != nil {
		return ""
	}
	return platform.ExpandPath(p, ctx)
}

func UserOverridePath() (string, error) {
	if p := os.Getenv("AISWEEP_REGISTRY"); p != "" {
		return p, nil
	}
	ctx, err := platform.Current()
	if err != nil {
		return "", err
	}
	base := ctx.ConfigDir
	if ctx.GOOS == "darwin" {
		// Preserve the pre-platform-support macOS override location.
		base = filepath.Join(ctx.HomeDir, ".config")
	}
	return filepath.Join(base, "aisweep", "registry.yaml"), nil
}
