package platform

import (
	"path/filepath"
	"testing"
)

func TestForUsesPlatformDirectoryConventions(t *testing.T) {
	home := filepath.Join("test", "home")
	paths := map[string]string{
		"APPDATA":      filepath.Join("test", "roaming"),
		"LOCALAPPDATA": filepath.Join("test", "local"),
	}
	getenv := func(key string) string { return paths[key] }

	tests := []struct {
		name      string
		goos      string
		configDir string
		dataDir   string
		cacheDir  string
		stateDir  string
	}{
		{
			name:      "windows",
			goos:      "windows",
			configDir: paths["APPDATA"],
			dataDir:   paths["LOCALAPPDATA"],
			cacheDir:  paths["LOCALAPPDATA"],
			stateDir:  paths["LOCALAPPDATA"],
		},
		{
			name:      "darwin",
			goos:      "darwin",
			configDir: filepath.Join(home, "Library", "Application Support"),
			dataDir:   filepath.Join(home, ".local", "share"),
			cacheDir:  filepath.Join(home, "Library", "Caches"),
			stateDir:  filepath.Join(home, ".local", "state"),
		},
		{
			name:      "linux",
			goos:      "linux",
			configDir: filepath.Join(home, ".config"),
			dataDir:   filepath.Join(home, ".local", "share"),
			cacheDir:  filepath.Join(home, ".cache"),
			stateDir:  filepath.Join(home, ".local", "state"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := For(tt.goos, home, getenv)
			if ctx.ConfigDir != tt.configDir || ctx.DataDir != tt.dataDir || ctx.CacheDir != tt.cacheDir || ctx.StateDir != tt.stateDir {
				t.Fatalf("got context %#v", ctx)
			}
		})
	}
}

func TestForUsesXDGOverrides(t *testing.T) {
	root := t.TempDir()
	home := filepath.Join(root, "home")
	overrides := map[string]string{
		"XDG_CONFIG_HOME": filepath.Join(root, "xdg-config"),
		"XDG_DATA_HOME":   filepath.Join(root, "xdg-data"),
		"XDG_CACHE_HOME":  filepath.Join(root, "xdg-cache"),
		"XDG_STATE_HOME":  filepath.Join(root, "xdg-state"),
	}
	ctx := For("linux", home, func(key string) string { return overrides[key] })
	if ctx.ConfigDir != overrides["XDG_CONFIG_HOME"] || ctx.DataDir != overrides["XDG_DATA_HOME"] || ctx.CacheDir != overrides["XDG_CACHE_HOME"] || ctx.StateDir != overrides["XDG_STATE_HOME"] {
		t.Fatalf("got context %#v", ctx)
	}
}

func TestExpandPathResolvesTemplates(t *testing.T) {
	ctx := Context{
		HomeDir:   filepath.Join("test", "home"),
		ConfigDir: filepath.Join("test", "config"),
		DataDir:   filepath.Join("test", "data"),
		CacheDir:  filepath.Join("test", "cache"),
		StateDir:  filepath.Join("test", "state"),
	}
	tests := map[string]string{
		"~/.codex":        filepath.Join("test", "home", ".codex"),
		"{config}/Cursor": filepath.Join("test", "config", "Cursor"),
		"{data}/aisweep":  filepath.Join("test", "data", "aisweep"),
		"{cache}/copilot": filepath.Join("test", "cache", "copilot"),
		"{state}/aisweep": filepath.Join("test", "state", "aisweep"),
	}
	for raw, want := range tests {
		if got := ExpandPath(raw, ctx); got != want {
			t.Errorf("ExpandPath(%q) = %q, want %q", raw, got, want)
		}
	}
}
