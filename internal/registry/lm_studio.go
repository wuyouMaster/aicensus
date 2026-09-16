package registry

type lmStudioScanner struct{}

func (lmStudioScanner) Definition() Tool {
	return Tool{
		ID:    "lm-studio",
		Label: "LM Studio",
		Paths: []Entry{{Path: "~/.lmstudio", Category: "models", Risk: "manual", Note: "delegate to LM Studio UI"}},
		PlatformPaths: map[string][]Entry{
			"darwin": {
				{Path: "{config}/lm-studio", Category: "models", Risk: "manual"},
				{Path: "{config}/LM Studio", Category: "models", Risk: "manual"},
			},
			"windows": {
				{Path: "{config}/lm-studio", Category: "models", Risk: "manual"},
				{Path: "{config}/LM Studio", Category: "models", Risk: "manual"},
			},
			"linux": {
				{Path: "{config}/lm-studio", Category: "models", Risk: "manual"},
				{Path: "{config}/LM Studio", Category: "models", Risk: "manual"},
			},
		},
	}
}

func (s lmStudioScanner) Discover() ([]Entry, error) {
	return discoverDefinition(s.Definition())
}
