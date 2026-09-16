package registry

type windsurfScanner struct{}

func (windsurfScanner) Definition() Tool {
	return Tool{
		ID:    "windsurf",
		Label: "Windsurf",
		Paths: []Entry{
			{Path: "~/.windsurf", Category: "cache", Risk: "safe"},
			{Path: "~/.windsurf/extensions", Category: "cache", Risk: "safe"},
		},
		MacOSPaths: []Entry{
			{Path: "~/Library/Application Support/Windsurf", Category: "cache", Risk: "safe"},
			{Path: "~/Library/Caches/Windsurf", Category: "cache", Risk: "safe"},
		},
	}
}

func (s windsurfScanner) Discover() ([]Entry, error) {
	return discoverDefinition(s.Definition())
}
