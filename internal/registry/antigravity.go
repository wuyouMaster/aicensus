package registry

type antigravityScanner struct{}

func (antigravityScanner) Definition() Tool {
	return Tool{
		ID:    "antigravity",
		Label: "Antigravity",
		Paths: []Entry{
			{Path: "~/.gemini/antigravity", Category: "cache", Risk: "safe"},
		},
		PlatformPaths: map[string][]Entry{
			"darwin": {
				{Path: "{config}/Antigravity", Category: "cache", Risk: "safe"},
			},
			"windows": {
				{Path: "{config}/Antigravity", Category: "cache", Risk: "safe"},
			},
			"linux": {
				{Path: "{config}/Antigravity", Category: "cache", Risk: "safe"},
			},
		},
	}
}

func (s antigravityScanner) Discover() ([]Entry, error) {
	return discoverDefinition(s.Definition())
}
