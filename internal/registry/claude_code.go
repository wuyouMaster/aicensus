package registry

type claudeCodeScanner struct{}

func (claudeCodeScanner) Definition() Tool {
	return Tool{
		ID:       "claude-code",
		Label:    "Claude Code",
		Homepage: "https://docs.claude.com/en/docs/claude-code",
		Paths: []Entry{
			{Path: "~/.claude", Category: "cache", Risk: "safe", Note: "sessions + cache, jaro covers cleanup"},
			{Path: "~/.claude/projects", Category: "transcripts", Risk: "archive", Note: "JSONL transcripts"},
			{Path: "~/.claude/transcripts", Category: "transcripts", Risk: "archive"},
			{Path: "~/.claude.json", Category: "config", Risk: "never", Note: "main config"},
		},
		MacOSPaths: []Entry{
			{Path: "~/Library/Application Support/Claude", Category: "cache", Risk: "safe"},
			{Path: "~/Library/Caches/Claude", Category: "cache", Risk: "safe"},
		},
	}
}

func (s claudeCodeScanner) Discover() ([]Entry, error) {
	return discoverDefinition(s.Definition())
}
