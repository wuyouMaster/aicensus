package registry

type chatGPTDesktopScanner struct{}

func (chatGPTDesktopScanner) Definition() Tool {
	return Tool{
		ID:    "chatgpt-desktop",
		Label: "ChatGPT Desktop",
		MacOSPaths: []Entry{
			{Path: "~/Library/Application Support/com.openai.chat", Category: "cache", Risk: "safe"},
		},
	}
}

func (s chatGPTDesktopScanner) Discover() ([]Entry, error) {
	return discoverDefinition(s.Definition())
}
