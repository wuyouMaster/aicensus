package registry

type openAIDesktopScanner struct{}

func (openAIDesktopScanner) Definition() Tool {
	return Tool{
		ID:    "openai-desktop",
		Label: "OpenAI Desktop",
		MacOSPaths: []Entry{
			{Path: "~/Library/Application Support/OpenAI", Category: "cache", Risk: "safe"},
		},
	}
}

func (s openAIDesktopScanner) Discover() ([]Entry, error) {
	return discoverDefinition(s.Definition())
}
