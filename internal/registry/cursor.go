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
		PlatformPaths: map[string][]Entry{
			"darwin": {
				{Path: "{config}/Cursor", Category: "cache", Risk: "safe"},
				{Path: "{config}/Cursor/snapshots", Category: "snapshots", Risk: "manual", Note: "time-travel snapshots"},
				{Path: "{config}/Cursor/logs", Category: "logs", Risk: "safe"},
				{Path: "{cache}/Cursor", Category: "cache", Risk: "safe"},
			},
			"windows": {
				{Path: "{config}/Cursor", Category: "cache", Risk: "safe"},
				{Path: "{config}/Cursor/logs", Category: "logs", Risk: "safe"},
				{Path: "{cache}/Cursor", Category: "cache", Risk: "safe"},
			},
			"linux": {
				{Path: "{config}/Cursor", Category: "cache", Risk: "safe"},
				{Path: "{config}/Cursor/logs", Category: "logs", Risk: "safe"},
				{Path: "{cache}/Cursor", Category: "cache", Risk: "safe"},
			},
		},
	}
}

func (s cursorScanner) Discover() ([]Entry, error) {
	return discoverDefinition(s.Definition())
}
