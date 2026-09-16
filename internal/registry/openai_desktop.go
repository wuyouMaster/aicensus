package registry

type openAIDesktopScanner struct{}

func (openAIDesktopScanner) Definition() Tool {
	return Tool{
		ID:    "openai-desktop",
		Label: "OpenAI Desktop",
		PlatformPaths: map[string][]Entry{
			"darwin": {
				{Path: "{config}/OpenAI", Category: "cache", Risk: "safe"},
			},
			"windows": {
				{Path: "{config}/OpenAI", Category: "cache", Risk: "safe"},
			},
			"linux": {
				{Path: "{config}/OpenAI", Category: "cache", Risk: "safe"},
			},
		},
	}
}

func (s openAIDesktopScanner) Discover() ([]Entry, error) {
	return discoverDefinition(s.Definition())
}
