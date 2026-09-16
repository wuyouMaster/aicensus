package registry

type lmStudioScanner struct{}

func (lmStudioScanner) Definition() Tool {
	return Tool{
		ID:    "lm-studio",
		Label: "LM Studio",
		Paths: []Entry{{Path: "~/.lmstudio", Category: "models", Risk: "manual", Note: "delegate to LM Studio UI"}},
		MacOSPaths: []Entry{
			{Path: "~/Library/Application Support/lm-studio", Category: "models", Risk: "manual"},
			{Path: "~/Library/Application Support/LM Studio", Category: "models", Risk: "manual"},
		},
	}
}

func (s lmStudioScanner) Discover() ([]Entry, error) {
	return discoverDefinition(s.Definition())
}
