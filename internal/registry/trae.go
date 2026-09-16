package registry

type traeScanner struct{}

func (traeScanner) Definition() Tool {
	return Tool{
		ID:       "trae",
		Label:    "Trae",
		Homepage: "https://www.trae.ai",
		Paths: []Entry{
			{Path: "~/.trae", Category: "cache", Risk: "safe"},
			{Path: "~/.trae/extensions", Category: "cache", Risk: "safe"},
		},
		PlatformPaths: map[string][]Entry{
			"darwin": {
				{Path: "{config}/Trae", Category: "cache", Risk: "safe"},
				{Path: "{config}/Trae/logs", Category: "logs", Risk: "archive"},
			},
			"windows": {
				{Path: "{config}/Trae", Category: "cache", Risk: "safe"},
				{Path: "{config}/Trae/logs", Category: "logs", Risk: "archive"},
				{Path: "{cache}/Trae", Category: "cache", Risk: "safe"},
			},
			"linux": {
				{Path: "{config}/Trae", Category: "cache", Risk: "safe"},
				{Path: "{config}/Trae/logs", Category: "logs", Risk: "archive"},
				{Path: "{cache}/Trae", Category: "cache", Risk: "safe"},
			},
		},
	}
}

func (s traeScanner) Discover() ([]Entry, error) {
	return discoverDefinition(s.Definition())
}
