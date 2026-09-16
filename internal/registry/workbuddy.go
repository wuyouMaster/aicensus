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
	home, err := os.UserHomeDir()
	if err != nil {
		return nil, nil
	}
	candidates := []Entry{
		{Path: filepath.Join(home, "Library/Application Support/WorkBuddy AI"), Category: "unknown", Risk: "manual", Note: "WorkBuddy desktop application data"},
		{Path: filepath.Join(home, "Library/Application Support/WorkBuddy"), Category: "unknown", Risk: "manual", Note: "WorkBuddy desktop application data"},
		{Path: filepath.Join(home, "Library/Application Support/CodeBuddyExtension"), Category: "cache", Risk: "manual", Note: "WorkBuddy companion extension data"},
		{Path: filepath.Join(home, "Library/Logs/WorkBuddy AI"), Category: "logs", Risk: "archive", Note: "open from WorkBuddy Help → Open Logs"},
		{Path: filepath.Join(home, "Library/Logs/WorkBuddy"), Category: "logs", Risk: "archive", Note: "open from WorkBuddy Help → Open Logs"},
		{Path: filepath.Join(home, "Library/Logs/CodeBuddy"), Category: "logs", Risk: "archive", Note: "open from WorkBuddy Help → Open Logs"},
	}
	found := make([]Entry, 0, len(candidates))
	for _, entry := range candidates {
		if _, err := os.Stat(entry.Path); err == nil {
			found = append(found, entry)
		}
	}
	return found, nil
}
