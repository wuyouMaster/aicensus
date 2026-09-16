package snapshot

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"time"

	"github.com/wuyouMaster/aicensus/internal/platform"
)

type Subdir struct {
	Path  string `json:"path"`
	Size  int64  `json:"size"`
	Files int64  `json:"files"`
}

type Entry struct {
	ToolID    string    `json:"tool_id"`
	ToolLabel string    `json:"tool_label"`
	Path      string    `json:"path"`
	Category  string    `json:"category"`
	Risk      string    `json:"risk"`
	Note      string    `json:"note,omitempty"`
	SizeBytes int64     `json:"size_bytes"`
	FileCount int64     `json:"file_count"`
	MTime     time.Time `json:"mtime"`
	TopSubs   []Subdir  `json:"top_subs,omitempty"`
}

type Snapshot struct {
	ID         string    `json:"id"`
	StartedAt  time.Time `json:"started_at"`
	FinishedAt time.Time `json:"finished_at"`
	Host       string    `json:"host"`
	Entries    []Entry   `json:"entries"`
}

type Summary struct {
	Total          int64
	ToolTotals     []ToolTotal
	CategoryTotals []CategoryTotal
	RiskTotals     []RiskTotal
}

type ToolTotal struct {
	ToolID    string `json:"tool_id"`
	ToolLabel string `json:"tool_label"`
	Size      int64  `json:"size"`
	Count     int    `json:"count"`
}

type CategoryTotal struct {
	Category string `json:"category"`
	Size     int64  `json:"size"`
}

type RiskTotal struct {
	Risk string `json:"risk"`
	Size int64  `json:"size"`
}

type HistoryPoint struct {
	ID        string           `json:"id"`
	StartedAt time.Time        `json:"started_at"`
	Total     int64            `json:"total"`
	ByTool    map[string]int64 `json:"by_tool"`
}

type History struct {
	ToolOrder []string       `json:"tool_order"`
	Points    []HistoryPoint `json:"points"`
}

// ToolCard is one row on the /tools list page: aggregated view of a single
// tool across all of its paths in one snapshot.
type ToolCard struct {
	ToolID            string          `json:"tool_id"`
	ToolLabel         string          `json:"tool_label"`
	Size              int64           `json:"size"`
	Count             int             `json:"count"`
	LatestMTime       time.Time       `json:"latest_mtime"`
	CategoryBreakdown []CategoryTotal `json:"category_breakdown"`
	RiskBreakdown     []RiskTotal     `json:"risk_breakdown"`
}

func Write(s *Snapshot) error {
	dir, err := dataDir()
	if err != nil {
		return err
	}
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return err
	}
	path := filepath.Join(dir, s.ID+".json")
	data, err := json.MarshalIndent(s, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0o644)
}

func List() ([]Snapshot, error) {
	dir, err := dataDir()
	if err != nil {
		return nil, err
	}
	files, err := filepath.Glob(filepath.Join(dir, "*.json"))
	if err != nil {
		return nil, err
	}
	sort.Strings(files)
	out := make([]Snapshot, 0, len(files))
	for _, f := range files {
		s, err := readFile(f)
		if err != nil {
			continue
		}
		out = append(out, *s)
	}
	return out, nil
}

func Get(id string) (*Snapshot, error) {
	dir, err := dataDir()
	if err != nil {
		return nil, err
	}
	return readFile(filepath.Join(dir, id+".json"))
}

func readFile(path string) (*Snapshot, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	s := &Snapshot{}
	if err := json.Unmarshal(data, s); err != nil {
		return nil, err
	}
	return s, nil
}

func dataDir() (string, error) {
	if p := os.Getenv("AISWEEP_DATA"); p != "" {
		return p, nil
	}
	ctx, err := platform.Current()
	if err != nil {
		return "", err
	}
	return filepath.Join(ctx.DataDir, "aisweep", "snapshots"), nil
}

func Summarize(s *Snapshot) Summary {
	byTool := map[string]*ToolTotal{}
	byCat := map[string]int64{}
	byRisk := map[string]int64{}
	var total int64
	for _, e := range s.Entries {
		total += e.SizeBytes
		t, ok := byTool[e.ToolID]
		if !ok {
			t = &ToolTotal{ToolID: e.ToolID, ToolLabel: e.ToolLabel}
			byTool[e.ToolID] = t
		}
		t.Size += e.SizeBytes
		t.Count++
		byCat[e.Category] += e.SizeBytes
		byRisk[e.Risk] += e.SizeBytes
	}
	tt := make([]ToolTotal, 0, len(byTool))
	for _, t := range byTool {
		tt = append(tt, *t)
	}
	sort.Slice(tt, func(i, j int) bool { return tt[i].Size > tt[j].Size })
	ct := make([]CategoryTotal, 0, len(byCat))
	for k, v := range byCat {
		ct = append(ct, CategoryTotal{Category: k, Size: v})
	}
	sort.Slice(ct, func(i, j int) bool { return ct[i].Size > ct[j].Size })
	rt := make([]RiskTotal, 0, len(byRisk))
	for k, v := range byRisk {
		rt = append(rt, RiskTotal{Risk: k, Size: v})
	}
	riskOrder := []string{"never", "manual", "archive", "safe", "unknown"}
	sort.SliceStable(rt, func(i, j int) bool {
		return indexOf(riskOrder, rt[i].Risk) < indexOf(riskOrder, rt[j].Risk)
	})
	return Summary{Total: total, ToolTotals: tt, CategoryTotals: ct, RiskTotals: rt}
}

func indexOf(s []string, t string) int {
	for i, v := range s {
		if v == t {
			return i
		}
	}
	return 999
}

// BuildHistory flattens a snapshot series into per-snapshot total + per-tool
// size points. Tool order is derived from the latest snapshot, then by
// descending peak size for tools absent in latest.
func BuildHistory(snaps []Snapshot) History {
	out := History{}
	if len(snaps) == 0 {
		return out
	}
	latest := snaps[len(snaps)-1]
	latestByTool := map[string]int64{}
	for _, e := range latest.Entries {
		latestByTool[e.ToolID] += e.SizeBytes
	}
	type sized struct {
		id   string
		size int64
	}
	var primary []sized
	for id, sz := range latestByTool {
		primary = append(primary, sized{id, sz})
	}
	sort.Slice(primary, func(i, j int) bool { return primary[i].size > primary[j].size })
	seen := map[string]bool{}
	for _, p := range primary {
		out.ToolOrder = append(out.ToolOrder, p.id)
		seen[p.id] = true
	}
	peak := map[string]int64{}
	for _, s := range snaps {
		byTool := map[string]int64{}
		for _, e := range s.Entries {
			if !seen[e.ToolID] {
				if cur, ok := peak[e.ToolID]; !ok || e.SizeBytes > cur {
					peak[e.ToolID] = e.SizeBytes
				}
			}
			byTool[e.ToolID] += e.SizeBytes
		}
		var total int64
		for _, v := range byTool {
			total += v
		}
		out.Points = append(out.Points, HistoryPoint{
			ID:        s.ID,
			StartedAt: s.StartedAt,
			Total:     total,
			ByTool:    byTool,
		})
	}
	if len(peak) > 0 {
		var tail []sized
		for id, sz := range peak {
			tail = append(tail, sized{id, sz})
		}
		sort.Slice(tail, func(i, j int) bool { return tail[i].size > tail[j].size })
		for _, t := range tail {
			out.ToolOrder = append(out.ToolOrder, t.id)
		}
	}
	return out
}

// BuildToolCards aggregates per-tool view for the /tools list page.
func BuildToolCards(s *Snapshot) []ToolCard {
	type acc struct {
		card   *ToolCard
		byCat  map[string]int64
		byRisk map[string]int64
	}
	m := map[string]*acc{}
	for _, e := range s.Entries {
		a, ok := m[e.ToolID]
		if !ok {
			a = &acc{card: &ToolCard{ToolID: e.ToolID, ToolLabel: e.ToolLabel}, byCat: map[string]int64{}, byRisk: map[string]int64{}}
			m[e.ToolID] = a
		}
		a.card.Size += e.SizeBytes
		a.card.Count++
		if e.MTime.After(a.card.LatestMTime) {
			a.card.LatestMTime = e.MTime
		}
		a.byCat[e.Category] += e.SizeBytes
		a.byRisk[e.Risk] += e.SizeBytes
	}
	out := make([]ToolCard, 0, len(m))
	for _, a := range m {
		for k, v := range a.byCat {
			a.card.CategoryBreakdown = append(a.card.CategoryBreakdown, CategoryTotal{Category: k, Size: v})
		}
		for k, v := range a.byRisk {
			a.card.RiskBreakdown = append(a.card.RiskBreakdown, RiskTotal{Risk: k, Size: v})
		}
		sort.Slice(a.card.CategoryBreakdown, func(i, j int) bool { return a.card.CategoryBreakdown[i].Size > a.card.CategoryBreakdown[j].Size })
		riskOrder := []string{"never", "manual", "archive", "safe", "unknown"}
		sort.SliceStable(a.card.RiskBreakdown, func(i, j int) bool {
			return indexOf(riskOrder, a.card.RiskBreakdown[i].Risk) < indexOf(riskOrder, a.card.RiskBreakdown[j].Risk)
		})
		out = append(out, *a.card)
	}
	sort.Slice(out, func(i, j int) bool { return out[i].Size > out[j].Size })
	return out
}

func FormatBytes(n int64) string {
	return formatBytes(n, 2)
}

// FormatBytesCompact formats a size with adaptive units and no decimal places.
// It is intended for compact labels such as chart axes.
func FormatBytesCompact(n int64) string {
	return formatBytes(n, 0)
}

func formatBytes(n int64, decimals int) string {
	const base = 1024.0
	units := [...]string{"B", "KB", "MB", "GB", "TB", "PB", "EB"}

	value := float64(n)
	sign := ""
	if value < 0 {
		sign = "-"
		value = -value
	}

	unit := 0
	for value >= base && unit < len(units)-1 {
		value /= base
		unit++
	}
	if unit == 0 {
		return fmt.Sprintf("%s%.0f B", sign, value)
	}
	return fmt.Sprintf("%s%.*f %s", sign, decimals, value, units[unit])
}
