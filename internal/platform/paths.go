package platform

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"
)

// Context contains the user-level directories used by scanners and by
// aisweep itself. ConfigDir is the roaming/config location, DataDir is the
// persistent user-data location, and CacheDir is safe-to-rebuild storage.
type Context struct {
	GOOS      string
	HomeDir   string
	ConfigDir string
	DataDir   string
	CacheDir  string
	StateDir  string
}

// Current resolves the standard directories for the current host.
func Current() (Context, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return Context{}, err
	}
	return For(runtime.GOOS, home, os.Getenv), nil
}

// For builds a directory context without reading process-global state. It is
// used by tests to exercise Windows, Linux, and macOS path rules on one host.
func For(goos, home string, getenv func(string) string) Context {
	if getenv == nil {
		getenv = func(string) string { return "" }
	}

	switch goos {
	case "windows":
		roaming := firstNonEmpty(getenv("APPDATA"), filepath.Join(home, "AppData", "Roaming"))
		local := firstNonEmpty(getenv("LOCALAPPDATA"), filepath.Join(home, "AppData", "Local"))
		return Context{
			GOOS:      goos,
			HomeDir:   home,
			ConfigDir: roaming,
			DataDir:   local,
			CacheDir:  local,
			StateDir:  local,
		}
	case "darwin":
		return Context{
			GOOS:      goos,
			HomeDir:   home,
			ConfigDir: filepath.Join(home, "Library", "Application Support"),
			// Keep aisweep's existing snapshot location stable on macOS.
			DataDir:  filepath.Join(home, ".local", "share"),
			CacheDir: filepath.Join(home, "Library", "Caches"),
			StateDir: filepath.Join(home, ".local", "state"),
		}
	default:
		return Context{
			GOOS:      goos,
			HomeDir:   home,
			ConfigDir: xdgDir(getenv("XDG_CONFIG_HOME"), filepath.Join(home, ".config")),
			DataDir:   xdgDir(getenv("XDG_DATA_HOME"), filepath.Join(home, ".local", "share")),
			CacheDir:  xdgDir(getenv("XDG_CACHE_HOME"), filepath.Join(home, ".cache")),
			StateDir:  xdgDir(getenv("XDG_STATE_HOME"), filepath.Join(home, ".local", "state")),
		}
	}
}

// ExpandPath resolves aisweep's path templates. Besides ~, built-in scanners
// can use {home}, {config}, {data}, {cache}, and {state}. Templates keep
// platform-specific knowledge in the scanner while snapshots store absolute
// paths after discovery.
func ExpandPath(raw string, ctx Context) string {
	p := strings.TrimSpace(raw)
	if p == "" {
		return ""
	}
	for token, value := range map[string]string{
		"{home}":   ctx.HomeDir,
		"{config}": ctx.ConfigDir,
		"{data}":   ctx.DataDir,
		"{cache}":  ctx.CacheDir,
		"{state}":  ctx.StateDir,
	} {
		p = strings.ReplaceAll(p, token, value)
	}
	if p == "~" {
		p = ctx.HomeDir
	} else if strings.HasPrefix(p, "~/") || strings.HasPrefix(p, "~\\") {
		p = filepath.Join(ctx.HomeDir, p[2:])
	}
	return filepath.Clean(p)
}

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		if strings.TrimSpace(value) != "" {
			return value
		}
	}
	return ""
}

func xdgDir(value, fallback string) string {
	value = strings.TrimSpace(value)
	if value == "" {
		return fallback
	}
	if filepath.IsAbs(value) {
		return filepath.Clean(value)
	}
	return fallback
}
