package registry

type chatGPTDesktopScanner struct{}

func (chatGPTDesktopScanner) Definition() Tool {
	return Tool{
		ID:    "chatgpt-desktop",
		Label: "ChatGPT Desktop",
		PlatformPaths: map[string][]Entry{
			"darwin": {
				{Path: "{config}/com.openai.chat", Category: "cache", Risk: "safe"},
			},
			"windows": {
				{Path: "{config}/com.openai.chat", Category: "cache", Risk: "safe"},
			},
			"linux": {
				{Path: "{config}/com.openai.chat", Category: "cache", Risk: "safe"},
			},
		},
	}
}

func (s chatGPTDesktopScanner) Discover() ([]Entry, error) {
	return discoverDefinition(s.Definition())
}
