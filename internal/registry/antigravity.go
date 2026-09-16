package registry

type antigravityScanner struct{}

func (antigravityScanner) Definition() Tool {
	return Tool{
		ID:    "antigravity",
		Label: "Antigravity",
		Paths: []Entry{
			{Path: "~/.gemini/antigravity", Category: "cache", Risk: "safe"},
		},
		MacOSPaths: []Entry{
			{Path: "~/Library/Application Support/Antigravity", Category: "cache", Risk: "safe"},
		},
	}
}

func (s antigravityScanner) Discover() ([]Entry, error) {
	return discoverDefinition(s.Definition())
}
