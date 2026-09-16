package scanner

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/uwa/aisweep/internal/registry"
	"github.com/uwa/aisweep/internal/snapshot"
)

type dirAgg struct {
	size  int64
	files int64
	mtime time.Time
}

type subdirInfo struct {
	Path  string
	Size  int64
	Files int64
}

type pathResult struct {
	Entry     registry.Entry
	ToolID    string
	ToolLabel string
	Found     bool
	Size      int64
	Files     int64
	MTime     time.Time
	TopSubs   []subdirInfo
}

type resolvedTool struct {
	tool    registry.Tool
	entries []registry.Entry
}

func Scan(reg *registry.Registry, progress io.Writer) (*snapshot.Snapshot, error) {
	startedAt := time.Now()
	host, _ := os.Hostname()

	skip := map[string]bool{}
	sources := reg.Scanners()
	resolved := make([]resolvedTool, 0, len(sources))
	for _, source := range sources {
		t := source.Definition()
		entries, err := source.Discover()
		if err != nil {
			return nil, fmt.Errorf("discover %s: %w", t.ID, err)
		}
		for _, e := range entries {
			if e.Path != "" {
				skip[filepath.Clean(e.Path)] = true
			}
		}
		resolved = append(resolved, resolvedTool{tool: t, entries: entries})
	}

	var results []pathResult
	for _, item := range resolved {
		for _, e := range item.entries {
			if e.Path == "" {
				continue
			}
			if progress != nil {
				fmt.Fprintf(progress, "scan %s ... ", shortPath(e.Path))
			}
			r := scanPath(item.tool.ID, item.tool.Label, e, skip)
			if progress != nil {
				if r.Found {
					fmt.Fprintf(progress, "%s\n", snapshot.FormatBytes(r.Size))
				} else {
					fmt.Fprintln(progress, "missing")
				}
			}
			results = append(results, r)
		}
	}

	finishedAt := time.Now()
	entries := make([]snapshot.Entry, 0, len(results))
	for _, r := range results {
		if !r.Found {
			continue
		}
		subs := make([]snapshot.Subdir, 0, len(r.TopSubs))
		for _, s := range r.TopSubs {
			subs = append(subs, snapshot.Subdir{Path: s.Path, Size: s.Size, Files: s.Files})
		}
		entries = append(entries, snapshot.Entry{
			ToolID:    r.ToolID,
			ToolLabel: r.ToolLabel,
			Path:      r.Entry.Path,
			Category:  r.Entry.Category,
			Risk:      r.Entry.Risk,
			Note:      r.Entry.Note,
			SizeBytes: r.Size,
			FileCount: r.Files,
			MTime:     r.MTime,
			TopSubs:   subs,
		})
	}

	return &snapshot.Snapshot{
		ID:         startedAt.UTC().Format("20060102T150405Z"),
		StartedAt:  startedAt,
		FinishedAt: finishedAt,
		Host:       host,
		Entries:    entries,
	}, nil
}

func scanPath(toolID, toolLabel string, e registry.Entry, skip map[string]bool) pathResult {
	res := pathResult{Entry: e, ToolID: toolID, ToolLabel: toolLabel}
	info, err := os.Stat(e.Path)
	if err != nil {
		return res
	}
	res.Found = true
	if !info.IsDir() {
		res.Size = info.Size()
		res.Files = 1
		res.MTime = info.ModTime()
		return res
	}

	rootKey := filepath.Clean(e.Path)
	agg := map[string]*dirAgg{rootKey: {}}
	_ = filepath.WalkDir(e.Path, func(p string, d os.DirEntry, werr error) error {
		if werr != nil {
			return nil
		}
		if d.IsDir() {
			key := filepath.Clean(p)
			if p != e.Path && skip[key] {
				return filepath.SkipDir
			}
			if _, ok := agg[key]; !ok {
				agg[key] = &dirAgg{}
			}
			return nil
		}
		if d.Type()&os.ModeSymlink != 0 {
			return nil
		}
		fi, ferr := d.Info()
		if ferr != nil {
			return nil
		}
		parent := filepath.Clean(filepath.Dir(p))
		a, ok := agg[parent]
		if !ok {
			a = &dirAgg{}
			agg[parent] = a
		}
		a.size += fi.Size()
		a.files++
		if fi.ModTime().After(a.mtime) {
			a.mtime = fi.ModTime()
		}
		return nil
	})

	// Bottom-up: each dir's recursive size = own direct size + sum of child recursive sizes.
	recursive := map[string]int64{}
	for p, a := range agg {
		recursive[p] = a.size
	}
	keys := make([]string, 0, len(agg))
	for k := range agg {
		keys = append(keys, k)
	}
	sort.Slice(keys, func(i, j int) bool { return pathDepth(keys[i]) > pathDepth(keys[j]) })
	for _, p := range keys {
		if p == rootKey {
			continue
		}
		recursive[filepath.Dir(p)] += recursive[p]
	}

	res.Size = recursive[rootKey]
	var totalFiles int64
	var latest time.Time
	for _, a := range agg {
		totalFiles += a.files
		if a.mtime.After(latest) {
			latest = a.mtime
		}
	}
	res.Files = totalFiles
	res.MTime = latest

	var subs []subdirInfo
	for p, a := range agg {
		if filepath.Dir(p) == rootKey && p != rootKey && recursive[p] > 0 {
			subs = append(subs, subdirInfo{Path: p, Size: recursive[p], Files: a.files})
		}
	}
	sort.Slice(subs, func(i, j int) bool { return subs[i].Size > subs[j].Size })
	if len(subs) > 10 {
		subs = subs[:10]
	}
	res.TopSubs = subs
	return res
}

func shortPath(p string) string {
	home, _ := os.UserHomeDir()
	if home != "" && strings.HasPrefix(p, home) {
		return "~" + p[len(home):]
	}
	return p
}

func pathDepth(p string) int {
	return strings.Count(p, string(os.PathSeparator))
}
