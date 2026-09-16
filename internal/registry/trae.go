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
		MacOSPaths: []Entry{
			{Path: "~/Library/Application Support/Trae", Category: "cache", Risk: "safe"},
			{Path: "~/Library/Application Support/Trae/logs", Category: "logs", Risk: "archive"},
		},
	}
}

func (s traeScanner) Discover() ([]Entry, error) {
	return discoverDefinition(s.Definition())
}
