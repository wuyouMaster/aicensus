// Package i18n provides language-aware message lookup and time formatting
// for templates rendered by the aisweep dashboard.
package i18n

import (
	"fmt"
	"time"
)

const (
	LangEN = "en"
	LangZH = "zh"
)

// Bundle holds message catalogs for every supported language.
type Bundle struct {
	msgs map[string]map[string]string
}

// New returns a Bundle preloaded with the built-in English and Chinese
// catalogs. Additional languages can be added later via Register.
func New() *Bundle {
	return &Bundle{msgs: map[string]map[string]string{
		LangEN: en(),
		LangZH: zh(),
	}}
}

// Register adds or replaces a language catalog at runtime. Existing keys are
// overwritten; new keys are appended.
func (b *Bundle) Register(lang string, msgs map[string]string) {
	if b.msgs[lang] == nil {
		b.msgs[lang] = map[string]string{}
	}
	for k, v := range msgs {
		b.msgs[lang][k] = v
	}
}

// T returns the message for the given key in the chosen language. Missing
// translations fall back to English and finally to the key itself so a
// template never crashes on a missing entry.
func (b *Bundle) T(lang, key string, args ...any) string {
	m := b.msgs[lang]
	s, ok := m[key]
	if !ok {
		m = b.msgs[LangEN]
		s, ok = m[key]
		if !ok {
			return key
		}
	}
	if len(args) == 0 {
		return s
	}
	return fmt.Sprintf(s, args...)
}

// Since formats a duration relative to now in the chosen language. Units are
// coarse (seconds/minutes/hours/days) and the format string lives in the
// catalog so each language can control its own phrasing.
func (b *Bundle) Since(lang string, t time.Time) string {
	d := time.Since(t)
	var n int
	unitKey := "u_sec"
	switch {
	case d < time.Minute:
		n = int(d.Seconds())
		unitKey = "u_sec"
	case d < time.Hour:
		n = int(d.Minutes())
		unitKey = "u_min"
	case d < 24*time.Hour:
		n = int(d.Hours())
		unitKey = "u_hr"
	default:
		n = int(d.Hours() / 24)
		unitKey = "u_day"
	}
	return b.T(lang, "since_fmt", n, b.T(lang, unitKey))
}

func en() map[string]string {
	return map[string]string{
		"nav_overview":     "overview",
		"nav_tools":        "tools",
		"page_title":       "aisweep",
		"page_tools_title": "tools",
		"page_tool_title":  "%s - aisweep",
		"tile_total":       "Total",
		"since_last_scan":  "Since last scan",
		"no_change":        "no change",
		"trend":            "Trend",
		"gran_day":         "Daily",
		"gran_hour":        "Hourly",
		"scan_now":         "scan now",
		"scan_scanning":    "scanning…",
		"scan_done":        "done — reloading",
		"scan_net_err":     "network error",
		"scan_err":         "error",
		"by_tool":          "By tool",
		"by_category":      "By category",
		"recent_snapshots": "Recent snapshots",
		"th_tool":          "Tool",
		"th_size":          "Size",
		"th_share":         "Share",
		"th_paths":         "Paths",
		"th_category":      "Category",
		"th_risk":          "Risk",
		"th_files":         "Files",
		"th_top_subs":      "Top subdirs",
		"th_id":            "ID",
		"th_started":       "Started",
		"th_total":         "Total",
		"th_delta":         "Delta",
		"th_entries":       "Entries",
		"chart_need_2":     "need at least 2 snapshots",
		"chart_empty":      "all snapshots empty",
		"legend_total":     "total",
		"meta_overview":    "%d tools · %d snapshots · last scan %s",
		"meta_tools":       "%d tools · total %s · last scan %s",
		"meta_tool":        "id %s · last scan %s",
		"pct_of_total":     "%.1f%% of total",
		"x_paths":          "%d paths",
		"latest":           "latest %s",
		"empty_body":       "No snapshots yet. Run `aisweep scan` or restart `aisweep serve` to create the first one.",
		"risk_safe":        "safe",
		"risk_archive":     "archive",
		"risk_manual":      "manual",
		"risk_never":       "never",
		"tip_total":        "Sum of every scanned path across all tools.",
		"tip_risk_safe":    "Regeneratable caches. Safe to delete — the tool re-downloads on demand.",
		"tip_risk_archive": "Old logs / sessions. Archive before delete if you want a backup.",
		"tip_risk_manual":  "Looks important enough that we won’t auto-clean. Inspect before deciding.",
		"tip_risk_never":   "Auth tokens or key material. Never auto-delete.",
		"cat_cache":        "cache",
		"cat_snapshots":    "snapshots",
		"cat_sessions":     "sessions",
		"cat_logs":         "logs",
		"cat_transcripts":  "transcripts",
		"cat_models":       "models",
		"cat_config":       "config",
		"cat_auth":         "auth",
		"cat_unknown":      "unknown",
		"since_fmt":        "%d%s ago",
		"u_sec":            "s",
		"u_min":            "m",
		"u_hr":             "h",
		"u_day":            "d",
		"lang_label":       "language",
	}
}

func zh() map[string]string {
	return map[string]string{
		"nav_overview":     "总览",
		"nav_tools":        "工具",
		"page_title":       "aisweep",
		"page_tools_title": "工具列表",
		"page_tool_title":  "%s - aisweep",
		"tile_total":       "总占用",
		"since_last_scan":  "相比上次扫描",
		"no_change":        "无变化",
		"trend":            "趋势",
		"gran_day":         "按天",
		"gran_hour":        "按小时",
		"scan_now":         "立即扫描",
		"scan_scanning":    "扫描中…",
		"scan_done":        "完成，正在刷新",
		"scan_net_err":     "网络错误",
		"scan_err":         "错误",
		"by_tool":          "按工具",
		"by_category":      "按分类",
		"recent_snapshots": "最近快照",
		"th_tool":          "工具",
		"th_size":          "大小",
		"th_share":         "占比",
		"th_paths":         "路径",
		"th_category":      "分类",
		"th_risk":          "风险",
		"th_files":         "文件数",
		"th_top_subs":      "主要子目录",
		"th_id":            "编号",
		"th_started":       "开始时间",
		"th_total":         "总大小",
		"th_delta":         "变化",
		"th_entries":       "条目数",
		"chart_need_2":     "需要至少 2 个快照",
		"chart_empty":      "所有快照为空",
		"legend_total":     "总计",
		"meta_overview":    "%d 个工具 · %d 个快照 · 上次扫描 %s",
		"meta_tools":       "%d 个工具 · 总计 %s · 上次扫描 %s",
		"meta_tool":        "编号 %s · 上次扫描 %s",
		"pct_of_total":     "占总计 %.1f%%",
		"x_paths":          "%d 个路径",
		"latest":           "最近 %s",
		"empty_body":       "暂无快照。运行 `aisweep scan` 或重启 `aisweep serve` 来生成第一份快照。",
		"risk_safe":        "安全",
		"risk_archive":     "可归档",
		"risk_manual":      "需手动",
		"risk_never":       "永不清理",
		"tip_total":        "所有 AI 工具被扫描路径的总占用。",
		"tip_risk_safe":    "可重新生成的缓存，删了工具会自己再下回来。",
		"tip_risk_archive": "老的日志 / 会话，建议先归档再删。",
		"tip_risk_manual":  "看着重要，不自动清理。需要你自己判断。",
		"tip_risk_never":   "凭证 / 密钥类，绝对不自动清理。",
		"cat_cache":        "缓存",
		"cat_snapshots":    "快照",
		"cat_sessions":     "会话",
		"cat_logs":         "日志",
		"cat_transcripts":  "转写",
		"cat_models":       "模型",
		"cat_config":       "配置",
		"cat_auth":         "凭证",
		"cat_unknown":      "未知",
		"since_fmt":        "%d%s前",
		"u_sec":            "秒",
		"u_min":            "分钟",
		"u_hr":             "小时",
		"u_day":            "天",
		"lang_label":       "语言",
	}
}
