package registry

type ollamaScanner struct{}

func (ollamaScanner) Definition() Tool {
	return Tool{
		ID:    "ollama",
		Label: "Ollama",
		Paths: []Entry{{Path: "~/.ollama", Category: "models", Risk: "manual", Note: "delegate to ollama rm"}},
	}
}

func (s ollamaScanner) Discover() ([]Entry, error) {
	return discoverDefinition(s.Definition())
}
