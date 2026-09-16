package server

import (
	"embed"
	"fmt"
	"html"
	"html/template"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/uwa/aisweep/internal/i18n"
	"github.com/uwa/aisweep/internal/registry"
	"github.com/uwa/aisweep/internal/snapshot"
)

//go:embed templates
var templatesFS embed.FS

var bundle = i18n.New()

var tpl = template.Must(template.New("").Funcs(template.FuncMap{
	"t":         bundle.T,
	"since":     bundle.Since,
	"bytes":     snapshot.FormatBytes,
	"pct":       pct,
	"shortPath": shortPath,
	"toolIcon":  toolIcon,
	"sparkline": sparkline,
	"sub":       func(a, b int64) int64 { return a - b },
	"subIdx":    func(a, b int) int { return a - b },
}).ParseFS(templatesFS, "templates/*.html"))

func Serve(host string, port int, scanFn func() error) error {
	addr := fmt.Sprintf("%s:%d", host, port)
	ln, err := net.Listen("tcp", addr)
	if err != nil {
		return err
	}
	if host != "127.0.0.1" && host != "localhost" && host != "::1" {
		fmt.Printf("WARNING: aisweep bound to %s (not loopback)\n", addr)
	}
	mux := http.NewServeMux()
	mux.HandleFunc("/", handleIndex)
	mux.HandleFunc("/tools", handleTools)
	mux.HandleFunc("/guide", handleGuide)
	mux.HandleFunc("/guide/", handleGuideTool)
	mux.HandleFunc("/tool/", handleTool)
	if scanFn != nil {
		mux.HandleFunc("/api/scan", func(w http.ResponseWriter, r *http.Request) {
			if r.Method != http.MethodPost {
				http.Error(w, "POST only", http.StatusMethodNotAllowed)
				return
			}
			if err := scanFn(); err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			w.WriteHeader(http.StatusNoContent)
		})
	}
	fmt.Printf("aisweep dashboard: http://%s\n", addr)
	return http.Serve(ln, mux)
}

func handleIndex(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}
	lang := detectLang(r)
	setLangCookie(w, lang)
	gran := detectGranularity(r)
	snaps, err := snapshot.List()
	if err != nil || len(snaps) == 0 {
		_ = tpl.ExecuteTemplate(w, "empty.html", map[string]any{"NavActive": "overview", "Lang": lang})
		return
	}
	latest := snaps[len(snaps)-1]
	summary := snapshot.Summarize(&latest)
	var prev *snapshot.Snapshot
	if len(snaps) >= 2 {
		p := snaps[len(snaps)-2]
		prev = &p
	}
	deltas := computeDeltas(&latest, prev)
	history := snapshot.BuildHistory(snaps)
	labels := map[string]string{}
	for _, e := range latest.Entries {
		labels[e.ToolID] = e.ToolLabel
	}
	data := map[string]any{
		"Snap":      latest,
		"Summary":   summary,
		"Deltas":    deltas,
		"Count":     len(snaps),
		"SnapList":  snaps,
		"History":   history,
		"Labels":    labels,
		"NavActive": "overview",
		"Lang":      lang,
		"Gran":      gran,
	}
	_ = tpl.ExecuteTemplate(w, "overview.html", data)
}

func handleTools(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/tools" {
		http.NotFound(w, r)
		return
	}
	lang := detectLang(r)
	setLangCookie(w, lang)
	snaps, err := snapshot.List()
	if err != nil || len(snaps) == 0 {
		_ = tpl.ExecuteTemplate(w, "empty.html", map[string]any{"NavActive": "tools", "Lang": lang})
		return
	}
	latest := snaps[len(snaps)-1]
	cards := snapshot.BuildToolCards(&latest)
	var total int64
	for _, c := range cards {
		total += c.Size
	}
	data := map[string]any{
		"Snap":      latest,
		"Cards":     cards,
		"Total":     total,
		"NavActive": "tools",
		"Lang":      lang,
	}
	_ = tpl.ExecuteTemplate(w, "tools.html", data)
}

type guidePath struct {
	Path           string
	Category       string
	Risk           string
	Note           string
	Found          bool
	Size           int64
	Files          int64
	ChildPathCount int
}

type guideTool struct {
	ID         string
	Label      string
	Homepage   string
	Entries    []guidePath
	FoundCount int
}

func handleGuide(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/guide" {
		http.NotFound(w, r)
		return
	}
	lang := detectLang(r)
	setLangCookie(w, lang)
	tools, hasSnapshot, latest := loadGuideTools()
	data := map[string]any{
		"Tools":       tools,
		"HasSnapshot": hasSnapshot,
		"Snap":        latest,
		"NavActive":   "guide",
		"Lang":        lang,
	}
	_ = tpl.ExecuteTemplate(w, "guide-list.html", data)
}

func handleGuideTool(w http.ResponseWriter, r *http.Request) {
	id := strings.TrimPrefix(r.URL.Path, "/guide/")
	if id == "" || strings.Contains(id, "/") {
		http.NotFound(w, r)
		return
	}
	lang := detectLang(r)
	setLangCookie(w, lang)
	tools, hasSnapshot, latest := loadGuideTools()
	for _, tool := range tools {
		if tool.ID != id {
			continue
		}
		data := map[string]any{
			"Tools":       []guideTool{tool},
			"HasSnapshot": hasSnapshot,
			"Snap":        latest,
			"NavActive":   "guide",
			"Lang":        lang,
		}
		_ = tpl.ExecuteTemplate(w, "guide.html", data)
		return
	}
	http.NotFound(w, r)
}

func loadGuideTools() ([]guideTool, bool, snapshot.Snapshot) {

	reg, err := registry.Load()
	if err != nil {
		reg = registry.New(registry.BuiltinScanners()...)
	}

	var latest snapshot.Snapshot
	hasSnapshot := false
	if snaps, listErr := snapshot.List(); listErr == nil && len(snaps) > 0 {
		latest = snaps[len(snaps)-1]
		hasSnapshot = true
	}
	observed := map[string]snapshot.Entry{}
	if hasSnapshot {
		for _, entry := range latest.Entries {
			observed[entry.ToolID+"\x00"+entry.Path] = entry
		}
	}

	tools := make([]guideTool, 0, len(reg.Tools))
	for _, tool := range reg.Tools {
		gt := guideTool{ID: tool.ID, Label: tool.Label, Homepage: tool.Homepage}
		for _, entry := range tool.AllPaths() {
			key := tool.ID + "\x00" + entry.Path
			current, found := observed[key]
			childPathCount := observedChildCount(tool.ID, entry.Path, observed)
			if !found && childPathCount > 0 {
				continue
			}
			gp := guidePath{
				Path:           entry.Path,
				Category:       entry.Category,
				Risk:           entry.Risk,
				Note:           entry.Note,
				ChildPathCount: childPathCount,
			}
			if found {
				gp.Found = true
				gp.Size = current.SizeBytes
				gp.Files = current.FileCount
				gt.FoundCount++
			}
			gt.Entries = append(gt.Entries, gp)
		}
		if len(gt.Entries) > 0 {
			tools = append(tools, gt)
		}
	}
	return tools, hasSnapshot, latest
}

func observedChildCount(toolID, parent string, observed map[string]snapshot.Entry) int {
	root := filepath.Clean(parent)
	count := 0
	for key := range observed {
		separator := strings.IndexByte(key, '\x00')
		if separator < 0 || key[:separator] != toolID {
			continue
		}
		child := filepath.Clean(key[separator+1:])
		rel, err := filepath.Rel(root, child)
		if err == nil && rel != "." && rel != ".." && !strings.HasPrefix(rel, ".."+string(os.PathSeparator)) && !filepath.IsAbs(rel) {
			count++
		}
	}
	return count
}

func handleTool(w http.ResponseWriter, r *http.Request) {
	id := strings.TrimPrefix(r.URL.Path, "/tool/")
	lang := detectLang(r)
	setLangCookie(w, lang)
	snaps, err := snapshot.List()
	if err != nil || len(snaps) == 0 {
		http.NotFound(w, r)
		return
	}
	latest := snaps[len(snaps)-1]
	var entries []snapshot.Entry
	var label string
	for _, e := range latest.Entries {
		if e.ToolID == id {
			entries = append(entries, e)
			label = e.ToolLabel
		}
	}
	if len(entries) == 0 {
		http.NotFound(w, r)
		return
	}
	sort.Slice(entries, func(i, j int) bool { return entries[i].SizeBytes > entries[j].SizeBytes })
	data := map[string]any{
		"ToolID":    id,
		"ToolLabel": label,
		"Snap":      latest,
		"Entries":   entries,
		"NavActive": "tool",
		"Lang":      lang,
	}
	_ = tpl.ExecuteTemplate(w, "tool.html", data)
}

func computeDeltas(latest, prev *snapshot.Snapshot) map[string]int64 {
	out := map[string]int64{"total": 0}
	if prev == nil {
		return out
	}
	prevByPath := map[string]int64{}
	for _, e := range prev.Entries {
		prevByPath[e.Path] = e.SizeBytes
	}
	var d int64
	for _, e := range latest.Entries {
		d += e.SizeBytes - prevByPath[e.Path]
	}
	out["total"] = d
	return out
}

func pct(n, total int64) string {
	if total == 0 {
		return "0"
	}
	return fmt.Sprintf("%.1f", float64(n)*100/float64(total))
}

func detectLang(r *http.Request) string {
	if c, err := r.Cookie("aisweep_lang"); err == nil {
		if c.Value == i18n.LangEN || c.Value == i18n.LangZH {
			return c.Value
		}
	}
	if q := r.URL.Query().Get("lang"); q == i18n.LangEN || q == i18n.LangZH {
		return q
	}
	if strings.HasPrefix(r.Header.Get("Accept-Language"), "zh") {
		return i18n.LangZH
	}
	return i18n.LangEN
}

func setLangCookie(w http.ResponseWriter, lang string) {
	http.SetCookie(w, &http.Cookie{
		Name:     "aisweep_lang",
		Value:    lang,
		Path:     "/",
		MaxAge:   365 * 24 * 3600,
		HttpOnly: false,
		SameSite: http.SameSiteLaxMode,
	})
}

func detectGranularity(r *http.Request) string {
	if q := r.URL.Query().Get("gran"); q == "day" || q == "hour" {
		return q
	}
	if c, err := r.Cookie("aisweep_gran"); err == nil {
		if c.Value == "day" || c.Value == "hour" {
			return c.Value
		}
	}
	return "day"
}

func shortPath(p string) string {
	home, _ := os.UserHomeDir()
	if home != "" && strings.HasPrefix(p, home) {
		return "~" + p[len(home):]
	}
	return p
}

var toolIconFiles = map[string]string{
	"cursor":          "cursor.svg",
	"claude-code":     "claude-code.svg",
	"codex-cli":       "openai.svg",
	"windsurf":        "windsurf.svg",
	"trae":            "trae.svg",
	"antigravity":     "antigravity.svg",
	"copilot-cli":     "copilot-cli.svg",
	"huggingface":     "huggingface.svg",
	"continue":        "continue.svg",
	"openai-desktop":  "openai.svg",
	"chatgpt-desktop": "openai.svg",
	"lm-studio":       "lm-studio.svg",
	"ollama":          "ollama.svg",
}

// toolIcon returns a bundled brand mark for known tools. The fallback is a
// safely escaped initial so user-defined registry entries never inject markup
// into the page.
func toolIcon(id, label string) template.HTML {
	if file, ok := toolIconFiles[id]; ok {
		if data, err := templatesFS.ReadFile("templates/icons/" + file); err == nil {
			return template.HTML(`<span class="tool-icon" aria-hidden="true">` + string(data) + `</span>`)
		}
	}
	{
		initial := "?"
		if r := []rune(strings.TrimSpace(label)); len(r) > 0 {
			initial = html.EscapeString(string(r[0]))
		}
		return template.HTML(`<span class="tool-icon tool-icon-fallback" aria-hidden="true">` + initial + `</span>`)
	}
}

// sparkline renders an inline SVG bar chart of total size aggregated by the
// selected time bucket, with stacked segments for the top 5 tools (by latest
// size). Each bucket gets an invisible hit-zone rect carrying data-*
// attrs so the browser tooltip on the overview page can show the exact
// values on hover. Returns template.HTML so the rendered SVG is injected
// verbatim.
func axisTick(v int64) string {
	const k = int64(1024)
	switch {
	case v >= k*k*k:
		return fmt.Sprintf("%.0f GB", float64(v)/float64(k*k*k))
	case v >= k*k:
		return fmt.Sprintf("%.0f MB", float64(v)/float64(k*k))
	case v >= k:
		return fmt.Sprintf("%.0f KB", float64(v)/float64(k))
	default:
		return fmt.Sprintf("%d B", v)
	}
}

func sparkline(lang, gran string, h snapshot.History, labels map[string]string) template.HTML {
	if len(h.Points) < 2 {
		return template.HTML(fmt.Sprintf(`<span class="meta">%s</span>`, bundle.T(lang, "chart_need_2")))
	}
	const (
		viewW = 720
		viewH = 200
		padL  = 56
		padR  = 8
		padT  = 12
		padB  = 28
	)
	// Bucket points by the selected granularity. A storage trend is a point-in-
	// time measurement, so the most recent snapshot in each bucket represents
	// that bucket instead of summing repeated measurements.
	type timeBucket struct {
		date    time.Time
		total   int64
		byTool  map[string]int64
		present bool
	}
	bucketMap := map[string]timeBucket{}
	bucketDuration := 24 * time.Hour
	if gran == "hour" {
		bucketDuration = time.Hour
	}
	for _, p := range h.Points {
		d := p.StartedAt.Truncate(bucketDuration)
		key := d.Format("2006-01-02 15:04")
		byTool := make(map[string]int64, len(p.ByTool))
		for tid, sz := range p.ByTool {
			byTool[tid] = sz
		}
		if _, ok := bucketMap[key]; ok {
			bucketMap[key] = timeBucket{date: d, total: p.Total, byTool: byTool, present: true}
			continue
		}
		bucketMap[key] = timeBucket{date: d, total: p.Total, byTool: byTool, present: true}
	}
	if len(bucketMap) == 0 {
		return template.HTML(fmt.Sprintf(`<span class="meta">%s</span>`, bundle.T(lang, "chart_empty")))
	}
	maxBuckets := 60
	if gran == "hour" {
		maxBuckets = 48
	}
	var first, last time.Time
	for _, bucket := range bucketMap {
		if first.IsZero() || bucket.date.Before(first) {
			first = bucket.date
		}
		if last.IsZero() || bucket.date.After(last) {
			last = bucket.date
		}
	}
	span := int(last.Sub(first)/bucketDuration) + 1
	if span > maxBuckets {
		first = last.Add(-time.Duration(maxBuckets-1) * bucketDuration)
		span = maxBuckets
	}
	buckets := make([]timeBucket, span)
	for i := range buckets {
		d := first.Add(time.Duration(i) * bucketDuration)
		bucket := bucketMap[d.Format("2006-01-02 15:04")]
		if !bucket.present {
			bucket.date = d
		}
		buckets[i] = bucket
	}
	// Max total across buckets for the y-axis scale.
	var maxTotal int64
	for _, d := range buckets {
		if d.present && d.total > maxTotal {
			maxTotal = d.total
		}
	}
	if maxTotal == 0 {
		return template.HTML(fmt.Sprintf(`<span class="meta">%s</span>`, bundle.T(lang, "chart_empty")))
	}

	// Pick top N tools by latest size for the stacked segments.
	topN := 5
	if len(h.ToolOrder) < topN {
		topN = len(h.ToolOrder)
	}
	topTools := h.ToolOrder[:topN]
	// A brighter Apple-inspired palette for the light Liquid Glass surface.
	// The hues move from cool to warm so neighboring stacked segments transition
	// smoothly instead of placing complementary colors side by side.
	palette := []string{"#2f80ed", "#31a8c7", "#4db879", "#e9a23b", "#e87578"}

	n := len(buckets)
	plotW := float64(viewW - padL - padR)
	plotH := float64(viewH - padT - padB)
	slotW := plotW / float64(n)
	barW := slotW * 0.55
	if barW > 56 {
		barW = 56
	}
	if barW < 6 {
		barW = 6
	}

	var b strings.Builder
	// Hover band rect — positioned by JS to the hovered slot's x.
	fmt.Fprintf(&b, "<svg viewBox=\"0 0 %d %d\" width=\"100%%\" height=\"200\" preserveAspectRatio=\"none\" class=\"chart\">", viewW, viewH)
	fmt.Fprintf(&b, "<rect class=\"chart-band\" x=\"%d\" y=\"0\" width=\"%.2f\" height=\"%d\" fill=\"rgba(0,122,255,0.07)\" style=\"display:none\"/>", padL, slotW, viewH)

	// Y-axis dashed gridlines + size labels (0/25/50/75/100%).
	for i := 0; i <= 4; i++ {
		frac := float64(i) / 4.0
		y := padT + plotH - frac*plotH
		v := int64(float64(maxTotal) * frac)
		fmt.Fprintf(&b, "<line x1=\"%d\" y1=\"%.2f\" x2=\"%d\" y2=\"%.2f\" stroke=\"rgba(92,110,150,0.16)\" stroke-width=\"1\" stroke-dasharray=\"2 4\"/>", padL, y, int(padL+plotW), y)
		fmt.Fprintf(&b, "<text x=\"%d\" y=\"%.2f\" fill=\"#9299ae\" font-size=\"10\" text-anchor=\"end\" dominant-baseline=\"middle\">%s</text>", padL-6, y, axisTick(v))
	}

	// Decide which buckets show an x-axis label: first, last, middle, and
	// ~evenly spaced in between for longer histories.
	labelIdxs := map[int]bool{0: true, n - 1: true}
	if n > 2 {
		labelIdxs[n/2] = true
	}
	if n > 6 {
		step := n / 5
		if step < 1 {
			step = 1
		}
		for i := 0; i < n; i += step {
			labelIdxs[i] = true
		}
	}

	var prevTotal int64
	var havePrev bool
	for i, d := range buckets {
		slotX := padL + float64(i)*slotW
		barX := slotX + (slotW-barW)/2

		// Stacked segments from bottom layer up.
		cum := int64(0)
		for li, tid := range topTools {
			sz := d.byTool[tid]
			if sz <= 0 {
				continue
			}
			segH := float64(sz) / float64(maxTotal) * plotH
			segY := padT + plotH - float64(cum+sz)/float64(maxTotal)*plotH
			col := palette[li%len(palette)]
			fmt.Fprintf(&b, "<rect x=\"%.2f\" y=\"%.2f\" width=\"%.2f\" height=\"%.2f\" fill=\"%s\"/>", barX, segY, barW, segH, col)
			cum += sz
		}
		if d.present && cum == 0 {
			fmt.Fprintf(&b, "<rect x=\"%.2f\" y=\"%.2f\" width=\"%.2f\" height=\"2\" fill=\"rgba(92,110,150,0.16)\" rx=\"1\" ry=\"1\"/>", barX, padT+plotH-2, barW)
		}

		if labelIdxs[i] {
			label := d.date.Format("01-02")
			if gran == "hour" {
				label = d.date.Format("15:04")
			}
			fmt.Fprintf(&b, "<text x=\"%.2f\" y=\"%d\" fill=\"#858ca1\" font-size=\"10\" text-anchor=\"middle\">%s</text>", slotX+slotW/2, int(padT+plotH+18), label)
		}

		if !d.present {
			continue
		}

		// Hit zone for tooltip (covers full slot width/height).
		var parts []string
		for _, tid := range topTools {
			sz := d.byTool[tid]
			if sz == 0 {
				continue
			}
			lbl := labels[tid]
			if lbl == "" {
				lbl = tid
			}
			parts = append(parts, html.EscapeString(lbl)+":"+snapshot.FormatBytes(sz))
		}
		detail := strings.Join(parts, "|")
		var delta string
		switch {
		case !havePrev:
			delta = "—"
		case d.total-prevTotal > 0:
			delta = "+" + snapshot.FormatBytes(d.total-prevTotal)
		case d.total-prevTotal < 0:
			delta = snapshot.FormatBytes(d.total - prevTotal)
		default:
			delta = "0"
		}
		prevTotal = d.total
		havePrev = true
		stamp := d.date.Format("01-02")
		if gran == "hour" {
			stamp = d.date.Format("01-02 15:04")
		}
		fmt.Fprintf(&b, `<rect class="chart-hit" data-i="%d" data-t="%s" data-total="%s" data-delta="%s" data-detail="%s" x="%.2f" y="0" width="%.2f" height="%d" fill="transparent" pointer-events="all"/>`, i, stamp, snapshot.FormatBytes(d.total), delta, detail, slotX, slotW, viewH)
	}
	b.WriteString("</svg>")

	// Legend — top N tools. Total is the sum of segments so it has no separate
	// swatch.
	b.WriteString("<div class=\"chart-legend\">")
	for li, tid := range topTools {
		lbl := labels[tid]
		if lbl == "" {
			lbl = tid
		}
		col := palette[li%len(palette)]
		fmt.Fprintf(&b, "<span><i style=\"background:%s\"></i>%s</span>", col, html.EscapeString(lbl))
	}
	b.WriteString("</div>")
	return template.HTML(b.String())
}
