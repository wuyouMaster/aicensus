package registry

type huggingFaceScanner struct{}

func (huggingFaceScanner) Definition() Tool {
	return Tool{
		ID:    "huggingface",
		Label: "Hugging Face",
		Paths: []Entry{{Path: "~/.cache/huggingface", Category: "models", Risk: "manual"}},
	}
}

func (s huggingFaceScanner) Discover() ([]Entry, error) {
	return discoverDefinition(s.Definition())
}
