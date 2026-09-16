package registry

import (
	"os"
	"strings"
)

type copilotCLIScanner struct{}

func (copilotCLIScanner) Definition() Tool {
	return Tool{
		ID:       "copilot-cli",
		Label:    "GitHub Copilot CLI",
		Homepage: "https://github.com/github/copilot-cli",
		Paths:    copilotEntries("~/.copilot"),
		PlatformPaths: map[string][]Entry{
			"darwin": {
				{Path: "{cache}/copilot", Category: "cache", Risk: "safe", Note: "marketplace and auto-update cache"},
			},
			"windows": {
				{Path: "{cache}/copilot", Category: "cache", Risk: "safe", Note: "marketplace and auto-update cache"},
			},
			"linux": {
				{Path: "{cache}/copilot", Category: "cache", Risk: "safe", Note: "marketplace and auto-update cache"},
			},
		},
	}
}

func (s copilotCLIScanner) Discover() ([]Entry, error) {
	root := "~/.copilot"
	if configured := strings.TrimSpace(os.Getenv("COPILOT_HOME")); configured != "" {
		root = configured
	}
	t := s.Definition()
	t.Paths = copilotEntries(root)
	return discoverDefinition(t)
}

func copilotEntries(root string) []Entry {
	return []Entry{
		{Path: root, Category: "unknown", Risk: "manual", Note: "Copilot CLI data root; child paths are classified separately"},
		{Path: root + "/settings.json", Category: "config", Risk: "never"},
		{Path: root + "/config.json", Category: "auth", Risk: "never", Note: "application state and authentication"},
		{Path: root + "/permissions-config.json", Category: "config", Risk: "never"},
		{Path: root + "/providers.json", Category: "auth", Risk: "never", Note: "provider and model configuration"},
		{Path: root + "/mcp-config.json", Category: "auth", Risk: "never", Note: "MCP configuration may contain credentials"},
		{Path: root + "/mcp-secrets", Category: "auth", Risk: "never"},
		{Path: root + "/mcp-oauth-config", Category: "auth", Risk: "never"},
		{Path: root + "/logs", Category: "logs", Risk: "archive"},
		{Path: root + "/session-state", Category: "sessions", Risk: "archive"},
		{Path: root + "/command-history-state", Category: "sessions", Risk: "archive"},
		{Path: root + "/session-store.db", Category: "sessions", Risk: "manual", Note: "cross-session index and search database"},
		{Path: root + "/installed-plugins", Category: "cache", Risk: "manual"},
		{Path: root + "/plugin-data", Category: "cache", Risk: "manual"},
		{Path: root + "/extensions", Category: "cache", Risk: "manual"},
		{Path: root + "/ide", Category: "cache", Risk: "safe"},
		{Path: root + "/agents", Category: "config", Risk: "never"},
		{Path: root + "/skills", Category: "config", Risk: "never"},
		{Path: root + "/hooks", Category: "config", Risk: "never"},
		{Path: root + "/instructions", Category: "config", Risk: "never"},
	}
}
