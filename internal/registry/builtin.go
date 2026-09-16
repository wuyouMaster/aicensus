package registry

// BuiltinScanners is the extension registry for first-party tool scanners.
// Add a concrete implementation here when adding a new AI tool.
func BuiltinScanners() []ToolScanner {
	return []ToolScanner{
		claudeCodeScanner{},
		codexCLIScanner{},
		qoderScanner{},
		kiroScanner{},
		clineScanner{},
		geminiCLIScanner{},
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
		workBuddyScanner{},
	}
}

func discoverDefinition(t Tool) ([]Entry, error) {
	return expandList(t.AllPaths()), nil
}
