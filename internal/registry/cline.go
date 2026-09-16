package registry

import (
	"os"
	"strings"
)

type clineScanner struct{}

func (clineScanner) Definition() Tool {
	return Tool{
		ID:    "cline",
		Label: "Cline",
		Paths: clineEntries("~/.cline", "~/.cline/data"),
	}
}

func (s clineScanner) Discover() ([]Entry, error) {
	t := s.Definition()
	if configured := strings.TrimSpace(os.Getenv("CLINE_DATA_DIR")); configured != "" {
		t.Paths = clineEntries("", configured)
	}
	return discoverDefinition(t)
}

func clineEntries(root, dataRoot string) []Entry {
	entries := make([]Entry, 0, 9)
	if root != "" {
		entries = append(entries,
			Entry{Path: root, Category: "unknown", Risk: "manual", Note: "Cline local data root; child paths are classified separately"},
			Entry{Path: root + "/plugins", Category: "cache", Risk: "manual", Note: "installed plugins"},
			Entry{Path: root + "/hooks", Category: "config", Risk: "never"},
		)
	}
	entries = append(entries,
		Entry{Path: dataRoot, Category: "unknown", Risk: "manual", Note: "Cline data directory; child paths are classified separately"},
		Entry{Path: dataRoot + "/settings", Category: "config", Risk: "never"},
		Entry{Path: dataRoot + "/secrets", Category: "auth", Risk: "never", Note: "credentials and tokens"},
		Entry{Path: dataRoot + "/sessions", Category: "sessions", Risk: "archive"},
		Entry{Path: dataRoot + "/logs", Category: "logs", Risk: "archive"},
		Entry{Path: dataRoot + "/workspaces", Category: "sessions", Risk: "archive", Note: "shared chat and project workspaces"},
	)
	return entries
}
