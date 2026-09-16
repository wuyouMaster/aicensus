package registry

import (
	"os"
	"strings"
)

type kiroScanner struct{}

func (kiroScanner) Definition() Tool {
	return Tool{
		ID:            "kiro",
		Label:         "Kiro",
		Homepage:      "https://kiro.dev",
		Paths:         kiroUserEntries("~/.kiro"),
		PlatformPaths: kiroPlatformEntries(),
	}
}

func (s kiroScanner) Discover() ([]Entry, error) {
	root := "~/.kiro"
	if configured := strings.TrimSpace(os.Getenv("KIRO_HOME")); configured != "" {
		root = configured
	}
	t := s.Definition()
	t.Paths = kiroUserEntries(root)
	return discoverDefinition(t)
}

func kiroUserEntries(root string) []Entry {
	return []Entry{
		{Path: root, Category: "unknown", Risk: "manual", Note: "Kiro user data root; child paths are classified separately"},
		{Path: root + "/settings", Category: "config", Risk: "never", Note: "CLI and IDE settings"},
		{Path: root + "/agents", Category: "config", Risk: "never"},
		{Path: root + "/prompts", Category: "config", Risk: "never"},
		{Path: root + "/skills", Category: "config", Risk: "never"},
		{Path: root + "/steering", Category: "config", Risk: "never"},
		{Path: root + "/hooks", Category: "config", Risk: "never"},
		{Path: root + "/powers", Category: "config", Risk: "never"},
		{Path: root + "/sessions", Category: "sessions", Risk: "archive"},
		{Path: root + "/workspace-roots", Category: "config", Risk: "never", Note: "workspace trust state"},
	}
}

func kiroPlatformEntries() map[string][]Entry {
	return map[string][]Entry{
		"darwin": {
			{Path: "{config}/kiro-cli/knowledge_bases", Category: "cache", Risk: "safe", Note: "local code knowledge indexes"},
		},
		"windows": {
			{Path: "{config}/kiro-cli/knowledge_bases", Category: "cache", Risk: "safe", Note: "local code knowledge indexes"},
		},
		"linux": {
			{Path: "{config}/kiro-cli/knowledge_bases", Category: "cache", Risk: "safe", Note: "local code knowledge indexes"},
		},
	}
}
