package registry

import (
	"os"
	"strings"
)

type qoderScanner struct{}

func (qoderScanner) Definition() Tool {
	return Tool{
		ID:            "qoder",
		Label:         "Qoder",
		Homepage:      "https://qoder.com",
		Paths:         qoderUserEntries("~/.qoder"),
		PlatformPaths: qoderPlatformEntries(),
	}
}

func (s qoderScanner) Discover() ([]Entry, error) {
	root := "~/.qoder"
	if configured := strings.TrimSpace(os.Getenv("QODER_CONFIG_DIR")); configured != "" {
		root = configured
	}
	t := s.Definition()
	t.Paths = qoderUserEntries(root)
	return discoverDefinition(t)
}

func qoderUserEntries(root string) []Entry {
	return []Entry{
		{Path: root, Category: "unknown", Risk: "manual", Note: "user data root; child paths are classified separately"},
		{Path: root + "/settings.json", Category: "config", Risk: "never", Note: "user settings"},
		{Path: root + "/mcp.json", Category: "auth", Risk: "never", Note: "MCP server configuration may contain credentials"},
		{Path: root + "/cache", Category: "cache", Risk: "safe"},
		{Path: root + "/bin", Category: "cache", Risk: "safe", Note: "downloaded runtime"},
		{Path: root + "/plugins", Category: "cache", Risk: "manual", Note: "installed plugins and plugin cache"},
		{Path: root + "/extensions", Category: "cache", Risk: "manual", Note: "installed editor extensions"},
		{Path: root + "/projects", Category: "sessions", Risk: "archive", Note: "project-scoped Qoder data"},
		{Path: root + "/memories", Category: "sessions", Risk: "archive", Note: "Qoder memory"},
		{Path: root + "/session-env", Category: "auth", Risk: "never", Note: "session environment data"},
		{Path: root + "/logs", Category: "logs", Risk: "archive"},
		{Path: root + "/darwin-bundle-rename.log", Category: "logs", Risk: "archive"},
	}
}

func qoderPlatformEntries() map[string][]Entry {
	return map[string][]Entry{
		"darwin":  qoderAppEntries("{config}"),
		"windows": qoderAppEntries("{config}"),
		"linux":   qoderAppEntries("{config}"),
	}
}

func qoderAppEntries(root string) []Entry {
	return []Entry{
		{Path: root + "/Qoder", Category: "cache", Risk: "manual", Note: "Qoder IDE application data"},
		{Path: root + "/Qoder/logs", Category: "logs", Risk: "archive"},
		{Path: root + "/Qoder/SharedClientCache", Category: "cache", Risk: "manual", Note: "workspace indexes and local runtime state"},
		{Path: root + "/Qoder/SharedClientCache/logs", Category: "logs", Risk: "archive"},
		{Path: root + "/Qoder/User/workspaceStorage", Category: "sessions", Risk: "archive"},
		{Path: root + "/QoderCN", Category: "cache", Risk: "manual", Note: "Qoder China edition application data"},
		{Path: root + "/QoderCN/logs", Category: "logs", Risk: "archive"},
		{Path: root + "/QoderCN/SharedClientCache", Category: "cache", Risk: "manual", Note: "workspace indexes and local runtime state"},
		{Path: root + "/QoderCN/SharedClientCache/logs", Category: "logs", Risk: "archive"},
		{Path: root + "/QoderCN/User/workspaceStorage", Category: "sessions", Risk: "archive"},
	}
}
