package registry

type continueScanner struct{}

func (continueScanner) Definition() Tool {
	return Tool{
		ID:    "continue",
		Label: "Continue",
		Paths: []Entry{{Path: "~/.continue", Category: "config", Risk: "never", Note: "extension config"}},
	}
}

func (s continueScanner) Discover() ([]Entry, error) {
	return discoverDefinition(s.Definition())
}
