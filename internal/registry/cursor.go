package registry

type cursorScanner struct{}

func (cursorScanner) Definition() Tool {
	return Tool{
		ID:       "cursor",
		Label:    "Cursor",
		Homepage: "https://www.cursor.com",
		Paths: []Entry{
			{Path: "~/.cursor", Category: "cache", Risk: "safe"},
			{Path: "~/.cursor/extensions", Category: "cache", Risk: "safe", Note: "reinstallable"},
			{Path: "~/.cursor/ai-tracking", Category: "logs", Risk: "manual"},
			{Path: "~/.cursor/worktrees", Category: "cache", Risk: "safe"},
		},
		MacOSPaths: []Entry{
			{Path: "~/Library/Application Support/Cursor", Category: "cache", Risk: "safe"},
			{Path: "~/Library/Application Support/Cursor/snapshots", Category: "snapshots", Risk: "manual", Note: "time-travel snapshots"},
			{Path: "~/Library/Application Support/Cursor/logs", Category: "logs", Risk: "safe"},
			{Path: "~/Library/Caches/Cursor", Category: "cache", Risk: "safe"},
		},
	}
}

func (s cursorScanner) Discover() ([]Entry, error) {
	return discoverDefinition(s.Definition())
}
