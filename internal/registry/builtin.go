package registry

// BuiltinScanners is the extension registry for first-party tool scanners.
// Add a concrete implementation here when adding a new AI tool.
func BuiltinScanners() []ToolScanner {
	return []ToolScanner{
		claudeCodeScanner{},
		codexCLIScanner{},
		cursorScanner{},
		windsurfScanner{},
		traeScanner{},
		antigravityScanner{},
		copilotCLIScanner{},
		huggingFaceScanner{},
		continueScanner{},
		llmScanner{},
		openAIDesktopScanner{},
		chatGPTDesktopScanner{},
		lmStudioScanner{},
		ollamaScanner{},
	}
}

func discoverDefinition(t Tool) ([]Entry, error) {
	return expandList(t.AllPaths()), nil
}
