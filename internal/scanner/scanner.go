package scanner

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"runtime"
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

type dirTotals struct {
	size  int64
	files int64
}

type subdirInfo struct {
	Path  string
	Size  int64
	Files int64
}

type pathResult struct {
	Entry               registry.Entry
	ToolID              string
	ToolLabel           string
	Found               bool
	Size                int64
	Files               int64
	MTime               time.Time
	TopSubs             []subdirInfo
	CoveredByChildPaths bool
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
				skip[pathKey(e.Path)] = true
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
			if r.CoveredByChildPaths {
				if progress != nil {
					fmt.Fprintln(progress, "covered by child paths")
				}
				continue
			}
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

	rootKey := pathKey(e.Path)
	hasChildPath := hasNestedPath(rootKey, skip)
	agg := map[string]*dirAgg{rootKey: {}}
	displayPaths := map[string]string{rootKey: filepath.Clean(e.Path)}
	_ = filepath.WalkDir(e.Path, func(p string, d os.DirEntry, werr error) error {
		if werr != nil {
			return nil
		}
		key := pathKey(p)
		// A separately registered path owns its complete subtree. Skip both
		// nested directories and nested files so parent and child entries never
		// count the same bytes twice.
		if key != rootKey && skip[key] {
			if d.IsDir() {
				return filepath.SkipDir
			}
			return nil
		}
		if d.IsDir() {
			if _, ok := agg[key]; !ok {
				agg[key] = &dirAgg{}
			}
			displayPaths[key] = filepath.Clean(p)
			return nil
		}
		if d.Type()&os.ModeSymlink != 0 {
			return nil
		}
		fi, ferr := d.Info()
		if ferr != nil {
			return nil
		}
		parent := pathKey(filepath.Dir(p))
		a, ok := agg[parent]
		if !ok {
			a = &dirAgg{}
			agg[parent] = a
			displayPaths[parent] = filepath.Clean(filepath.Dir(p))
		}
		a.size += fi.Size()
		a.files++
		if fi.ModTime().After(a.mtime) {
			a.mtime = fi.ModTime()
		}
		return nil
	})

	// Bottom-up: each dir's recursive size and file count include its direct
	// contents plus all descendants.
	recursive := make(map[string]dirTotals, len(agg))
	for p, a := range agg {
		recursive[p] = dirTotals{size: a.size, files: a.files}
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
		parent := filepath.Dir(p)
		child := recursive[p]
		parentTotal := recursive[parent]
		parentTotal.size += child.size
		parentTotal.files += child.files
		recursive[parent] = parentTotal
	}

	rootTotal := recursive[rootKey]
	res.Size = rootTotal.size
	res.Files = rootTotal.files
	// A directory with no remaining files after nested registered paths have
	// been skipped is only a container for those child entries. Do not emit a
	// redundant parent result in that case.
	res.CoveredByChildPaths = hasChildPath && res.Files == 0
	var latest time.Time
	for _, a := range agg {
		if a.mtime.After(latest) {
			latest = a.mtime
		}
	}
	res.MTime = latest

	var subs []subdirInfo
	for p := range agg {
		totals := recursive[p]
		if filepath.Dir(p) == rootKey && p != rootKey && totals.size > 0 {
			subs = append(subs, subdirInfo{Path: displayPaths[p], Size: totals.size, Files: totals.files})
		}
	}
	sort.Slice(subs, func(i, j int) bool { return subs[i].Size > subs[j].Size })
	if len(subs) > 10 {
		subs = subs[:10]
	}
	res.TopSubs = subs
	return res
}

func hasNestedPath(root string, paths map[string]bool) bool {
	root = pathKey(root)
	for candidate := range paths {
		candidate = pathKey(candidate)
		if candidate == root {
			continue
		}
		rel, err := filepath.Rel(root, candidate)
		if err != nil || rel == "." || rel == ".." || strings.HasPrefix(rel, ".."+string(os.PathSeparator)) || filepath.IsAbs(rel) {
			continue
		}
		return true
	}
	return false
}

func pathKey(path string) string {
	return pathKeyFor(path, runtime.GOOS)
}

func pathKeyFor(path, goos string) string {
	if absolute, err := filepath.Abs(path); err == nil {
		path = absolute
	}
	path = filepath.Clean(path)
	if goos == "windows" {
		return strings.ToLower(path)
	}
	return path
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
