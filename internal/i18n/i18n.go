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
		"nav_label":                         "primary navigation",
		"nav_overview":                      "overview",
		"nav_tools":                         "tools",
		"nav_guide":                         "scan guide",
		"page_title":                        "aisweep",
		"page_tools_title":                  "tools",
		"page_tool_title":                   "%s - aisweep",
		"tile_total":                        "Total",
		"since_last_scan":                   "Since last scan",
		"no_change":                         "no change",
		"trend":                             "Trend",
		"gran_day":                          "Daily",
		"gran_hour":                         "Hourly",
		"scan_now":                          "scan now",
		"scan_scanning":                     "scanning…",
		"scan_done":                         "done — reloading",
		"scan_net_err":                      "network error",
		"scan_err":                          "error",
		"by_tool":                           "By tool",
		"by_category":                       "By category",
		"recent_snapshots":                  "Recent snapshots",
		"th_tool":                           "Tool",
		"th_size":                           "Size",
		"th_share":                          "Share",
		"th_paths":                          "Paths",
		"th_category":                       "Category",
		"th_risk":                           "Risk",
		"th_files":                          "Files",
		"th_top_subs":                       "Top subdirs",
		"subdir_files":                      "%d files",
		"th_id":                             "ID",
		"th_started":                        "Started",
		"th_total":                          "Total",
		"th_delta":                          "Delta",
		"th_entries":                        "Entries",
		"chart_need_2":                      "need at least 2 snapshots",
		"chart_empty":                       "all snapshots empty",
		"legend_total":                      "total",
		"meta_overview":                     "%d tools · %d snapshots · last scan %s",
		"meta_tools":                        "%d tools · total %s · last scan %s",
		"meta_tool":                         "id %s · last scan %s",
		"pct_of_total":                      "%s%% of total",
		"x_paths":                           "%d paths",
		"latest":                            "latest %s",
		"empty_body":                        "No snapshots yet. Run `aisweep scan` or restart `aisweep serve` to create the first one.",
		"risk_safe":                         "safe",
		"risk_archive":                      "archive",
		"risk_manual":                       "manual",
		"risk_never":                        "never",
		"tip_total":                         "Sum of every scanned path across all tools.",
		"tip_risk_safe":                     "Regeneratable caches. Safe to delete — the tool re-downloads on demand.",
		"tip_risk_archive":                  "Old logs / sessions. Archive before delete if you want a backup.",
		"tip_risk_manual":                   "Looks important enough that we won’t auto-clean. Inspect before deciding.",
		"tip_risk_never":                    "Auth tokens or key material. Never auto-delete.",
		"tip_category_cache":                "Regeneratable files the tool can recreate or download when needed.",
		"tip_category_snapshots":            "Point-in-time recovery data or local history used to restore earlier states.",
		"tip_category_sessions":             "Working state from conversations, tasks, or projects.",
		"tip_category_logs":                 "Diagnostic and activity records used for troubleshooting and history.",
		"tip_category_transcripts":          "Saved text or audio transcriptions from past interactions.",
		"tip_category_models":               "Downloaded model files used to run AI features locally.",
		"tip_category_config":               "Preferences and local settings that control tool behavior.",
		"tip_category_auth":                 "Sign-in tokens and credentials. Treat these files as sensitive.",
		"tip_category_unknown":              "Files whose purpose was not identified by the scanner.",
		"show_explanation":                  "Show explanation",
		"guide_title":                       "Scan guide",
		"guide_detail_title":                "Tool scan details",
		"guide_intro":                       "See which paths aisweep checks, why each path has its category, and what the tool uses the data for.",
		"guide_card_intro":                  "Paths, category rules, and tool usage",
		"guide_path_count":                  "%d scan paths",
		"guide_found_count":                 "%d found",
		"guide_open":                        "View scan details",
		"guide_back":                        "Back to scan guide",
		"guide_platform":                    "Paths are resolved for the current operating system. A missing path is kept here so the scanner’s coverage stays visible.",
		"guide_scope_note":                  "Each path is reported as a non-overlapping scope. Separately scanned child paths are excluded from the parent, and a parent with no remaining files is hidden.",
		"guide_remaining_scope":             "Remaining scope: only files outside %d separately scanned child path(s) are counted here; child paths are shown separately.",
		"guide_path":                        "Scan path",
		"guide_category_basis":              "Why this category",
		"guide_usage":                       "How the tool uses it",
		"guide_found":                       "found",
		"guide_missing":                     "not found",
		"guide_files":                       "%d files",
		"guide_no_snapshot":                 "No snapshot has been collected yet, so current size and file count are unavailable.",
		"guide_note":                        "Scanner note",
		"guide_homepage":                    "website",
		"guide_tool_default":                "This section explains the local files that support this AI tool and how aisweep groups them.",
		"guide_category_reason_cache":       "Generated or downloaded data that supports faster operation and can usually be recreated.",
		"guide_category_reason_snapshots":   "Point-in-time copies kept so the tool can restore or compare an earlier state.",
		"guide_category_reason_sessions":    "Persistent conversation or task state that lets the tool continue previous work.",
		"guide_category_reason_logs":        "Diagnostic and activity records written while the tool runs.",
		"guide_category_reason_transcripts": "Saved text or audio-to-text records from previous interactions.",
		"guide_category_reason_models":      "Downloaded model artifacts that are read when local AI features run.",
		"guide_category_reason_config":      "Preferences and local settings read to configure the tool’s behavior.",
		"guide_category_reason_auth":        "Tokens or credentials used to authenticate with the tool or its services.",
		"guide_category_reason_unknown":     "The scanner found the path but does not know enough about its contents to classify it.",
		"guide_category_use_cache":          "The tool reads these files as reusable local data, indexes, extensions, or temporary results.",
		"guide_category_use_snapshots":      "The tool reads these files for recovery, time travel, or local history.",
		"guide_category_use_sessions":       "The tool reads and updates them to restore conversations, tasks, or project context.",
		"guide_category_use_logs":           "The tool appends diagnostic events and reads them when troubleshooting or showing activity.",
		"guide_category_use_transcripts":    "The tool reads them to display or search previous interaction content.",
		"guide_category_use_models":         "The tool loads these artifacts when it needs the corresponding local model.",
		"guide_category_use_config":         "The tool reads them at startup or while running to apply preferences and integrations.",
		"guide_category_use_auth":           "The tool reads them to sign requests or establish an authenticated session.",
		"guide_category_use_unknown":        "The tool’s use is not documented by the scanner yet; inspect before making changes.",
		"cat_cache":                         "cache",
		"cat_snapshots":                     "snapshots",
		"cat_sessions":                      "sessions",
		"cat_logs":                          "logs",
		"cat_transcripts":                   "transcripts",
		"cat_models":                        "models",
		"cat_config":                        "config",
		"cat_auth":                          "auth",
		"cat_unknown":                       "unknown",
		"since_fmt":                         "%d%s ago",
		"u_sec":                             "s",
		"u_min":                             "m",
		"u_hr":                              "h",
		"u_day":                             "d",
		"lang_label":                        "language",
	}
}

func zh() map[string]string {
	return map[string]string{
		"nav_label":                         "主导航",
		"nav_overview":                      "总览",
		"nav_tools":                         "工具",
		"nav_guide":                         "扫描说明",
		"page_title":                        "aisweep",
		"page_tools_title":                  "工具列表",
		"page_tool_title":                   "%s - aisweep",
		"tile_total":                        "总占用",
		"since_last_scan":                   "相比上次扫描",
		"no_change":                         "无变化",
		"trend":                             "趋势",
		"gran_day":                          "按天",
		"gran_hour":                         "按小时",
		"scan_now":                          "立即扫描",
		"scan_scanning":                     "扫描中…",
		"scan_done":                         "完成，正在刷新",
		"scan_net_err":                      "网络错误",
		"scan_err":                          "错误",
		"by_tool":                           "按工具",
		"by_category":                       "按分类",
		"recent_snapshots":                  "最近快照",
		"th_tool":                           "工具",
		"th_size":                           "大小",
		"th_share":                          "占比",
		"th_paths":                          "路径",
		"th_category":                       "分类",
		"th_risk":                           "风险",
		"th_files":                          "文件数",
		"th_top_subs":                       "主要子目录",
		"subdir_files":                      "%d 个文件",
		"th_id":                             "编号",
		"th_started":                        "开始时间",
		"th_total":                          "总大小",
		"th_delta":                          "变化",
		"th_entries":                        "条目数",
		"chart_need_2":                      "需要至少 2 个快照",
		"chart_empty":                       "所有快照为空",
		"legend_total":                      "总计",
		"meta_overview":                     "%d 个工具 · %d 个快照 · 上次扫描 %s",
		"meta_tools":                        "%d 个工具 · 总计 %s · 上次扫描 %s",
		"meta_tool":                         "编号 %s · 上次扫描 %s",
		"pct_of_total":                      "占总计 %s%%",
		"x_paths":                           "%d 个路径",
		"latest":                            "最近 %s",
		"empty_body":                        "暂无快照。运行 `aisweep scan` 或重启 `aisweep serve` 来生成第一份快照。",
		"risk_safe":                         "安全",
		"risk_archive":                      "可归档",
		"risk_manual":                       "需手动",
		"risk_never":                        "永不清理",
		"tip_total":                         "所有 AI 工具被扫描路径的总占用。",
		"tip_risk_safe":                     "可重新生成的缓存，删了工具会自己再下回来。",
		"tip_risk_archive":                  "老的日志 / 会话，建议先归档再删。",
		"tip_risk_manual":                   "看着重要，不自动清理。需要你自己判断。",
		"tip_risk_never":                    "凭证 / 密钥类，绝对不自动清理。",
		"tip_category_cache":                "可以重新生成或重新下载的临时数据。",
		"tip_category_snapshots":            "用于恢复历史状态的时间点副本或本地历史记录。",
		"tip_category_sessions":             "对话、任务或项目产生的工作状态。",
		"tip_category_logs":                 "用于排查问题和回看活动的诊断记录。",
		"tip_category_transcripts":          "保存的对话文本或音频转写内容。",
		"tip_category_models":               "本地 AI 功能使用的已下载模型文件。",
		"tip_category_config":               "控制工具行为的偏好和本地设置。",
		"tip_category_auth":                 "登录令牌和凭证，属于敏感数据。",
		"tip_category_unknown":              "扫描器暂时无法识别用途的文件。",
		"show_explanation":                  "显示说明",
		"guide_title":                       "扫描说明",
		"guide_detail_title":                "工具扫描详情",
		"guide_intro":                       "查看 aisweep 扫描哪些路径、为什么归到这个分类，以及工具通常如何使用这些数据。",
		"guide_card_intro":                  "扫描路径、分类依据和工具用途",
		"guide_path_count":                  "%d 条扫描路径",
		"guide_found_count":                 "%d 条已找到",
		"guide_open":                        "查看扫描详情",
		"guide_back":                        "返回扫描说明",
		"guide_platform":                    "路径会根据当前操作系统解析。即使路径不存在也会保留在这里，方便查看扫描覆盖范围。",
		"guide_scope_note":                  "每条路径按不重叠范围统计；单独扫描的子路径会从父路径中排除，父路径没有剩余文件时会自动隐藏。",
		"guide_remaining_scope":             "剩余范围：这里只统计 %d 个单独扫描子路径之外的文件；子路径占用会单独展示。",
		"guide_path":                        "扫描路径",
		"guide_category_basis":              "分类依据",
		"guide_usage":                       "工具用途",
		"guide_found":                       "已找到",
		"guide_missing":                     "未找到",
		"guide_files":                       "%d 个文件",
		"guide_no_snapshot":                 "还没有采集快照，因此暂时没有当前大小和文件数。",
		"guide_note":                        "扫描器备注",
		"guide_homepage":                    "官方网站",
		"guide_tool_default":                "这里说明支持该 AI 工具运行的本地文件，以及 aisweep 如何对它们分类。",
		"guide_category_reason_cache":       "支持工具运行速度的生成数据或下载数据，通常可以重新生成。",
		"guide_category_reason_snapshots":   "按时间点保存的副本，用于恢复或比较之前的状态。",
		"guide_category_reason_sessions":    "保存对话或任务状态，让工具可以继续之前的工作。",
		"guide_category_reason_logs":        "工具运行期间写入的诊断记录和活动记录。",
		"guide_category_reason_transcripts": "之前交互产生的文本记录或音频转写内容。",
		"guide_category_reason_models":      "本地 AI 功能运行时读取的已下载模型文件。",
		"guide_category_reason_config":      "工具读取的偏好和本地设置，用于配置运行行为。",
		"guide_category_reason_auth":        "用于登录工具或其服务的令牌和凭证。",
		"guide_category_reason_unknown":     "扫描器找到了这个路径，但暂时无法根据内容确定分类。",
		"guide_category_use_cache":          "工具会把它们作为可复用的本地数据、索引、扩展或临时结果来读取。",
		"guide_category_use_snapshots":      "工具会读取它们来完成恢复、时间回溯或查看本地历史。",
		"guide_category_use_sessions":       "工具会读取和更新它们，以恢复对话、任务或项目上下文。",
		"guide_category_use_logs":           "工具会追加诊断事件，并在排查问题或展示活动时读取它们。",
		"guide_category_use_transcripts":    "工具会读取它们来展示或搜索之前的交互内容。",
		"guide_category_use_models":         "工具需要对应的本地模型时，会加载这些文件。",
		"guide_category_use_config":         "工具会在启动或运行过程中读取它们，应用偏好和集成设置。",
		"guide_category_use_auth":           "工具会读取它们来签名请求或建立已认证的会话。",
		"guide_category_use_unknown":        "扫描器还没有记录工具如何使用它们，修改前请先自行检查。",
		"cat_cache":                         "缓存",
		"cat_snapshots":                     "快照",
		"cat_sessions":                      "会话",
		"cat_logs":                          "日志",
		"cat_transcripts":                   "转写",
		"cat_models":                        "模型",
		"cat_config":                        "配置",
		"cat_auth":                          "凭证",
		"cat_unknown":                       "未知",
		"since_fmt":                         "%d%s前",
		"u_sec":                             "秒",
		"u_min":                             "分钟",
		"u_hr":                              "小时",
		"u_day":                             "天",
		"lang_label":                        "语言",
	}
}
