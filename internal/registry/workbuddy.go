package registry

import (
	"os"
	"path/filepath"
)

type workBuddyScanner struct{}

func (workBuddyScanner) Definition() Tool {
	return Tool{
		ID:       "workbuddy",
		Label:    "WorkBuddy",
		Homepage: "https://www.workbuddy.cn",
	}
}

func (s workBuddyScanner) Discover() ([]Entry, error) {
	ctx, err := CurrentPlatformContext()
	if err != nil {
		return nil, nil
	}
	candidates := workBuddyCandidates(ctx)
	found := make([]Entry, 0, len(candidates))
	seen := map[string]bool{}
	for _, entry := range candidates {
		if seen[entry.Path] {
			continue
		}
		seen[entry.Path] = true
		if _, err := os.Stat(entry.Path); err == nil {
			found = append(found, entry)
		}
	}
	return found, nil
}

func workBuddyCandidates(ctx PlatformContext) []Entry {
	entries := []Entry{
		{Path: filepath.Join(ctx.ConfigDir, "WorkBuddy AI"), Category: "unknown", Risk: "manual", Note: "WorkBuddy desktop application data"},
		{Path: filepath.Join(ctx.ConfigDir, "WorkBuddy"), Category: "unknown", Risk: "manual", Note: "WorkBuddy desktop application data"},
		{Path: filepath.Join(ctx.ConfigDir, "CodeBuddyExtension"), Category: "cache", Risk: "manual", Note: "WorkBuddy companion extension data"},
		{Path: filepath.Join(ctx.DataDir, "WorkBuddy AI"), Category: "unknown", Risk: "manual", Note: "WorkBuddy desktop application data"},
		{Path: filepath.Join(ctx.DataDir, "WorkBuddy"), Category: "unknown", Risk: "manual", Note: "WorkBuddy desktop application data"},
		{Path: filepath.Join(ctx.DataDir, "CodeBuddyExtension"), Category: "cache", Risk: "manual", Note: "WorkBuddy companion extension data"},
	}
	if ctx.GOOS == "darwin" {
		entries = append(entries,
			Entry{Path: filepath.Join(ctx.HomeDir, "Library", "Logs", "WorkBuddy AI"), Category: "logs", Risk: "archive", Note: "open from WorkBuddy Help → Open Logs"},
			Entry{Path: filepath.Join(ctx.HomeDir, "Library", "Logs", "WorkBuddy"), Category: "logs", Risk: "archive", Note: "open from WorkBuddy Help → Open Logs"},
			Entry{Path: filepath.Join(ctx.HomeDir, "Library", "Logs", "CodeBuddy"), Category: "logs", Risk: "archive", Note: "open from WorkBuddy Help → Open Logs"},
		)
	} else {
		entries = append(entries,
			Entry{Path: filepath.Join(ctx.DataDir, "WorkBuddy AI", "logs"), Category: "logs", Risk: "archive", Note: "WorkBuddy diagnostic logs"},
			Entry{Path: filepath.Join(ctx.DataDir, "WorkBuddy", "logs"), Category: "logs", Risk: "archive", Note: "WorkBuddy diagnostic logs"},
			Entry{Path: filepath.Join(ctx.ConfigDir, "WorkBuddy AI", "logs"), Category: "logs", Risk: "archive", Note: "WorkBuddy diagnostic logs"},
			Entry{Path: filepath.Join(ctx.ConfigDir, "WorkBuddy", "logs"), Category: "logs", Risk: "archive", Note: "WorkBuddy diagnostic logs"},
		)
	}
	return entries
}
