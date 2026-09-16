package registry

type llmScanner struct{}

func (llmScanner) Definition() Tool {
	return Tool{
		ID:    "llm",
		Label: "Simon Willison llm",
		PlatformPaths: map[string][]Entry{
			"darwin": {
				{Path: "{config}/io.datasette.llm", Category: "cache", Risk: "manual"},
			},
			"windows": {
				{Path: "{config}/io.datasette.llm", Category: "cache", Risk: "manual"},
			},
			"linux": {
				{Path: "{config}/io.datasette.llm", Category: "cache", Risk: "manual"},
			},
		},
	}
}

func (s llmScanner) Discover() ([]Entry, error) {
	return discoverDefinition(s.Definition())
}
