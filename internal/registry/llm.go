package registry

type llmScanner struct{}

func (llmScanner) Definition() Tool {
	return Tool{
		ID:    "llm",
		Label: "Simon Willison llm",
		MacOSPaths: []Entry{
			{Path: "~/Library/Application Support/io.datasette.llm", Category: "cache", Risk: "manual"},
		},
	}
}

func (s llmScanner) Discover() ([]Entry, error) {
	return discoverDefinition(s.Definition())
}
