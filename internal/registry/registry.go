package registry

import (
	_ "embed"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"gopkg.in/yaml.v3"
)

//go:embed registry.yaml
var builtinYAML []byte

type Entry struct {
	Path     string `yaml:"path"`
	Category string `yaml:"category"`
	Risk     string `yaml:"risk"`
	Note     string `yaml:"note,omitempty"`
}

type Tool struct {
	ID         string  `yaml:"id"`
	Label      string  `yaml:"label"`
	Homepage   string  `yaml:"homepage,omitempty"`
	Paths      []Entry `yaml:"paths,omitempty"`
	MacOSPaths []Entry `yaml:"macos_paths,omitempty"`
}

type Registry struct {
	Tools []Tool `yaml:"tools"`
}

func (t Tool) AllPaths() []Entry {
	if runtime.GOOS == "darwin" {
		return append(append([]Entry{}, t.Paths...), t.MacOSPaths...)
	}
	return t.Paths
}

func Load() (*Registry, error) {
	reg := &Registry{}
	if err := yaml.Unmarshal(builtinYAML, reg); err != nil {
		return nil, fmt.Errorf("parse builtin registry: %w", err)
	}
	if up, err := UserOverridePath(); err == nil {
		if data, err := os.ReadFile(up); err == nil {
			over := &Registry{}
			if err := yaml.Unmarshal(data, over); err != nil {
				return nil, fmt.Errorf("parse user registry %s: %w", up, err)
			}
			reg.merge(over)
		}
	}
	reg.expandPaths()
	return reg, nil
}

func (r *Registry) merge(o *Registry) {
	idx := map[string]int{}
	for i, t := range r.Tools {
		idx[t.ID] = i
	}
	for _, t := range o.Tools {
		i, ok := idx[t.ID]
		if !ok {
			r.Tools = append(r.Tools, t)
			continue
		}
		if len(t.Paths) > 0 {
			r.Tools[i].Paths = t.Paths
		}
		if len(t.MacOSPaths) > 0 {
			r.Tools[i].MacOSPaths = t.MacOSPaths
		}
		if t.Label != "" {
			r.Tools[i].Label = t.Label
		}
		if t.Homepage != "" {
			r.Tools[i].Homepage = t.Homepage
		}
	}
}

func (r *Registry) expandPaths() {
	for i := range r.Tools {
		t := &r.Tools[i]
		t.Paths = expandList(t.Paths)
		t.MacOSPaths = expandList(t.MacOSPaths)
	}
}

func expandList(es []Entry) []Entry {
	out := make([]Entry, 0, len(es))
	for _, e := range es {
		e.Path = expand(e.Path)
		if e.Path == "" {
			continue
		}
		out = append(out, e)
	}
	return out
}

func expand(p string) string {
	p = strings.TrimSpace(p)
	if p == "" {
		return ""
	}
	if strings.HasPrefix(p, "~") {
		home, _ := os.UserHomeDir()
		if strings.HasPrefix(p, "~/") {
			p = filepath.Join(home, p[2:])
		} else if p == "~" {
			p = home
		}
	}
	return filepath.Clean(p)
}

func UserOverridePath() (string, error) {
	if p := os.Getenv("AISWEEP_REGISTRY"); p != "" {
		return p, nil
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, ".config", "aisweep", "registry.yaml"), nil
}
