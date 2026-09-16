package server

import (
	"embed"
	"fmt"
	"html"
	"html/template"
	"net"
	"net/http"
	"os"
	"sort"
	"strings"

	"github.com/uwa/aisweep/internal/i18n"
	"github.com/uwa/aisweep/internal/snapshot"
)

//go:embed templates/*
var templatesFS embed.FS

var bundle = i18n.New()

var tpl = template.Must(template.New("").Funcs(template.FuncMap{
	"t":         bundle.T,
	"since":     bundle.Since,
	"bytes":     snapshot.FormatBytes,
	"pct":       pct,
	"shortPath": shortPath,
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

func shortPath(p string) string {
	home, _ := os.UserHomeDir()
	if home != "" && strings.HasPrefix(p, home) {
		return "~" + p[len(home):]
	}
	return p
}

// sparkline renders an inline SVG line chart of total size across history
// points, with stacked translucent areas for the top 5 tools (by latest size).
// Each data point also gets an invisible hit-zone rect carrying data-* attrs
// so the browser tooltip on the overview page can show the exact values on
// hover. Returns template.HTML so the rendered SVG is injected verbatim.
func sparkline(lang string, h snapshot.History, labels map[string]string) template.HTML {
	if len(h.Points) < 2 {
		return template.HTML(fmt.Sprintf(`<span class="meta">%s</span>`, bundle.T(lang, "chart_need_2")))
	}
	const (
		viewW  = 720
		viewH  = 180
		padL   = 48
		padR   = 8
		padT   = 12
		padB   = 24
	)
	n := len(h.Points)
	plotW := float64(viewW - padL - padR)
	plotH := float64(viewH - padT - padB)

	type ts struct {
		x    float64
		y    float64
		tot  float64
	}
	pts := make([]ts, n)
	var maxTotal int64
	for _, p := range h.Points {
		if p.Total > maxTotal {
			maxTotal = p.Total
		}
	}
	if maxTotal == 0 {
		return template.HTML(fmt.Sprintf(`<span class="meta">%s</span>`, bundle.T(lang, "chart_empty")))
	}
	for i, p := range h.Points {
		x := padL + float64(i)/float64(n-1)*plotW
		y := padT + plotH - float64(p.Total)/float64(maxTotal)*plotH
		pts[i] = ts{x: x, y: y, tot: float64(p.Total)}
	}

	// Pick top 5 tools by latest size for stacked areas.
	topN := 5
	if len(h.ToolOrder) < topN {
		topN = len(h.ToolOrder)
	}
	topTools := h.ToolOrder[:topN]
	palette := []string{"#4f8cf7", "#7cc4ff", "#4ade80", "#f5c451", "#f97373"}

	type layer struct {
		id   string
		col  string
		poly string
	}
	layers := make([]layer, 0, topN)
	for li, id := range topTools {
		col := palette[li%len(palette)]
		var path strings.Builder
		// top edge: cumulative sum from bottom layer up
		cumUp := make([]float64, n)
		for k := 0; k <= li; k++ {
			for j, p := range h.Points {
				cumUp[j] += float64(p.ByTool[topTools[k]])
			}
		}
		for j := 0; j < n; j++ {
			x := padL + float64(j)/float64(n-1)*plotW
			y := padT + plotH - cumUp[j]/float64(maxTotal)*plotH
			if j == 0 {
				path.WriteString(fmt.Sprintf("M%.2f,%.2f", x, y))
			} else {
				path.WriteString(fmt.Sprintf(" L%.2f,%.2f", x, y))
			}
		}
		// close down to baseline
		path.WriteString(fmt.Sprintf(" L%.2f,%.2f", padL+plotW, padT+plotH))
		path.WriteString(fmt.Sprintf(" L%.2f,%.2f", float64(padL), padT+plotH))
		path.WriteString(" Z")
		layers = append(layers, layer{id: id, col: col, poly: path.String()})
	}

	var totalPath strings.Builder
	for i, p := range pts {
		if i == 0 {
			totalPath.WriteString(fmt.Sprintf("M%.2f,%.2f", p.x, p.y))
		} else {
			totalPath.WriteString(fmt.Sprintf(" L%.2f,%.2f", p.x, p.y))
		}
	}

	var b strings.Builder
	fmt.Fprintf(&b, "<svg viewBox=\"0 0 %d %d\" width=\"100%%\" height=\"180\" preserveAspectRatio=\"none\" class=\"chart\">", viewW, viewH)
	// y axis gridlines at 0/25/50/75/100% with byte labels
	for i := 0; i <= 4; i++ {
		frac := float64(i) / 4.0
		y := padT + plotH - frac*plotH
		v := int64(float64(maxTotal) * frac)
		fmt.Fprintf(&b, "<line x1=\"%d\" y1=\"%.2f\" x2=\"%d\" y2=\"%.2f\" stroke=\"#232831\" stroke-width=\"1\"/>", padL, y, int(padL+plotW), y)
		fmt.Fprintf(&b, "<text x=\"%d\" y=\"%.2f\" fill=\"#8a93a2\" font-size=\"10\" text-anchor=\"end\" dominant-baseline=\"middle\">%s</text>", padL-6, y, snapshot.FormatBytes(v))
	}
	// tool stacked areas (bottom-up)
	for _, l := range layers {
		fmt.Fprintf(&b, "<path d=\"%s\" fill=\"%s\" fill-opacity=\"0.18\" stroke=\"%s\" stroke-width=\"0.5\"/>", l.poly, l.col, l.col)
	}
	// total line (top, bold)
	fmt.Fprintf(&b, "<path d=\"%s\" fill=\"none\" stroke=\"#d6dde6\" stroke-width=\"2\"/>", totalPath.String())
	// point markers + x labels (first/middle/last)
	pick := []int{0, n / 2, n - 1}
	seen := map[int]bool{}
	for _, idx := range pick {
		if seen[idx] {
			continue
		}
		seen[idx] = true
		p := pts[idx]
		fmt.Fprintf(&b, "<circle cx=\"%.2f\" cy=\"%.2f\" r=\"3\" fill=\"#d6dde6\"/>", p.x, p.y)
		label := h.Points[idx].StartedAt.Format("01-02 15:04")
		fmt.Fprintf(&b, "<text x=\"%.2f\" y=\"%d\" fill=\"#8a93a2\" font-size=\"10\" text-anchor=\"middle\">%s</text>", p.x, int(padT+plotH+14), label)
	}
	// invisible hit zones — one per data point — so the JS layer can show a
	// tooltip for any point on hover. Each rect carries the per-point values
	// in data-* attributes (HTML-escaped labels, pipe-separated details).
	stepW := plotW / float64(n-1)
	var prevTot int64
	for i, p := range pts {
		var hx, hw float64
		switch i {
		case 0:
			hx = float64(padL)
			hw = stepW / 2
		case n - 1:
			hx = float64(padL) + plotW - stepW/2
			hw = stepW / 2
		default:
			hx = p.x - stepW/2
			hw = stepW
		}
		var parts []string
		for _, id := range topTools {
			label := labels[id]
			if label == "" {
				label = id
			}
			sz := h.Points[i].ByTool[id]
			parts = append(parts, html.EscapeString(label)+":"+snapshot.FormatBytes(sz))
		}
		detail := strings.Join(parts, "|")
		tot := int64(p.tot)
		var delta string
		switch {
		case i == 0:
			delta = "—"
		case tot-prevTot > 0:
			delta = "+" + snapshot.FormatBytes(tot - prevTot)
		case tot-prevTot < 0:
			delta = snapshot.FormatBytes(tot - prevTot)
		default:
			delta = "0"
		}
		prevTot = tot
		fmt.Fprintf(&b, `<rect class="chart-hit" data-i="%d" data-t="%s" data-total="%s" data-delta="%s" data-detail="%s" x="%.2f" y="0" width="%.2f" height="%d" fill="transparent" pointer-events="all"/>`,
			i,
			h.Points[i].StartedAt.Format("01-02 15:04"),
			snapshot.FormatBytes(tot),
			delta,
			detail,
			hx, hw, viewH,
		)
	}
	b.WriteString("</svg>")
	// legend
	b.WriteString("<div class=\"chart-legend\">")
	for _, l := range layers {
		label := labels[l.id]
		if label == "" {
			label = l.id
		}
		fmt.Fprintf(&b, "<span><i style=\"background:%s\"></i>%s</span>", l.col, html.EscapeString(label))
	}
	fmt.Fprintf(&b, "<span><i style=\"background:#d6dde6\"></i>%s</span>", bundle.T(lang, "legend_total"))
	b.WriteString("</div>")
	return template.HTML(b.String())
}
