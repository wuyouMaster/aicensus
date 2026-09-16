package registry

type copilotCLIScanner struct{}

func (copilotCLIScanner) Definition() Tool {
	return Tool{
		ID:    "copilot-cli",
		Label: "GitHub Copilot CLI",
		Paths: []Entry{{Path: "~/.copilot", Category: "cache", Risk: "safe"}},
	}
}

func (s copilotCLIScanner) Discover() ([]Entry, error) {
	return discoverDefinition(s.Definition())
}
