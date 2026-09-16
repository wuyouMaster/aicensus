package registry

type codexCLIScanner struct{}

func (codexCLIScanner) Definition() Tool {
	return Tool{
		ID:       "codex-cli",
		Label:    "Codex CLI",
		Homepage: "https://github.com/openai/codex",
		Paths: []Entry{
			{Path: "~/.codex", Category: "cache", Risk: "safe"},
			{Path: "~/.codex/sessions", Category: "sessions", Risk: "archive", Note: "session history; archive before delete"},
			{Path: "~/.codex/archived_sessions", Category: "sessions", Risk: "archive"},
			{Path: "~/.codex/thread_history_1.sqlite", Category: "sessions", Risk: "manual", Note: "thread history SQLite"},
			{Path: "~/.codex/logs_2.sqlite", Category: "logs", Risk: "archive"},
		},
		PlatformPaths: map[string][]Entry{
			"darwin": {
				{Path: "{config}/Codex", Category: "cache", Risk: "safe"},
			},
			"windows": {
				{Path: "{config}/Codex", Category: "cache", Risk: "safe"},
				{Path: "{cache}/Codex", Category: "cache", Risk: "safe"},
			},
			"linux": {
				{Path: "{config}/Codex", Category: "cache", Risk: "safe"},
				{Path: "{cache}/Codex", Category: "cache", Risk: "safe"},
			},
		},
	}
}

func (s codexCLIScanner) Discover() ([]Entry, error) {
	return discoverDefinition(s.Definition())
}
