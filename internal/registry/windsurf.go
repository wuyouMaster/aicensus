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
		PlatformPaths: map[string][]Entry{
			"darwin": {
				{Path: "{config}/Windsurf", Category: "cache", Risk: "safe"},
				{Path: "{cache}/Windsurf", Category: "cache", Risk: "safe"},
			},
			"windows": {
				{Path: "{config}/Windsurf", Category: "cache", Risk: "safe"},
				{Path: "{config}/Windsurf/logs", Category: "logs", Risk: "archive"},
				{Path: "{cache}/Windsurf", Category: "cache", Risk: "safe"},
			},
			"linux": {
				{Path: "{config}/Windsurf", Category: "cache", Risk: "safe"},
				{Path: "{config}/Windsurf/logs", Category: "logs", Risk: "archive"},
				{Path: "{cache}/Windsurf", Category: "cache", Risk: "safe"},
			},
		},
	}
}

func (s windsurfScanner) Discover() ([]Entry, error) {
	return discoverDefinition(s.Definition())
}
