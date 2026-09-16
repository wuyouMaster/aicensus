package registry

import (
	"os"
	"path/filepath"
	"strings"
)

type geminiCLIScanner struct{}

func (geminiCLIScanner) Definition() Tool {
	return Tool{
		ID:    "gemini-cli",
		Label: "Gemini CLI",
		Paths: geminiEntries("~/.gemini"),
	}
}

func (s geminiCLIScanner) Discover() ([]Entry, error) {
	root := "~/.gemini"
	if configured := strings.TrimSpace(os.Getenv("GEMINI_CLI_HOME")); configured != "" {
		root = filepath.Join(configured, ".gemini")
	}
	entries := expandList(geminiEntries(root))
	// ~/.gemini is shared by Gemini CLI and Antigravity. Require a Gemini
	// specific child before claiming the shared root for this scanner.
	for _, entry := range entries[1:] {
		if _, err := os.Stat(entry.Path); err == nil {
			return entries, nil
		}
	}
	return nil, nil
}

func geminiEntries(root string) []Entry {
	return []Entry{
		{Path: root, Category: "unknown", Risk: "manual", Note: "Gemini CLI data root; child paths are classified separately"},
		{Path: root + "/settings.json", Category: "config", Risk: "never", Note: "user settings"},
		{Path: root + "/GEMINI.md", Category: "config", Risk: "never", Note: "user instructions"},
		{Path: root + "/tmp", Category: "sessions", Risk: "archive", Note: "project sessions, chats, shell history, and checkpoints"},
		{Path: root + "/history", Category: "snapshots", Risk: "manual", Note: "checkpoint shadow repositories"},
		{Path: root + "/commands", Category: "config", Risk: "never"},
	}
}
