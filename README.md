# aisweep

Read-only inventory + trend tracker for AI CLI / app storage on macOS, Windows,
and Linux.

Phase 1: scan -> classify -> trend. No deletion yet (phase 2).

## Build

    go build -o bin/aisweep ./cmd/aisweep

## Quick start

    aisweep serve --interval=1h        # scan + serve http://127.0.0.1:7890
    # click "scan now" on the dashboard for an on-demand rescan

## Subcommands

    aisweep scan     # one scan, writes to the platform data directory
    aisweep serve    # local dashboard at http://127.0.0.1:7890
    aisweep path     # print effective registry
    aisweep doctor   # show which registered paths exist on this host

Flags:

    aisweep serve --port 7890 --host 127.0.0.1
    aisweep serve --interval=1h         # background rescan every 1h
    aisweep serve --scan=false          # skip the initial scan on startup

Override the snapshot directory with `AISWEEP_DATA=/path/to/snapshots`.

Cross-build the same binary for the supported desktop platforms:

    GOOS=darwin  GOARCH=arm64 go build -o bin/aisweep-darwin ./cmd/aisweep
    GOOS=linux   GOARCH=amd64 go build -o bin/aisweep-linux ./cmd/aisweep
    GOOS=windows GOARCH=amd64 go build -o bin/aisweep-windows.exe ./cmd/aisweep

## Endpoints

    GET  /              overview: totals, sparkline, history table
    GET  /tool/<id>     one tool's paths + top subdirs
    POST /api/scan      trigger an immediate rescan (returns 204)

The dashboard binds 127.0.0.1 only by default. Bind to a public address at
your own risk -- a warning is printed.

## Trend

Each snapshot records total bytes per tool. The overview renders a stacked
SVG sparkline (top 5 tools as translucent areas, total as a bold line) and a
history table with total + delta vs the previous snapshot. At least 2
snapshots are required; with `--interval` the dashboard fills in over time.

No external JS/CSS/font, single static HTML, dark theme.

## Registry

Built-in scanners live in `internal/registry/`, with one implementation per
AI tool. Each implementation satisfies `registry.ToolScanner`:

- `Definition() registry.Tool` returns the tool metadata and default paths.
- `Discover() ([]registry.Entry, error)` resolves the paths for the current
  machine. A tool can use dynamic discovery here when its storage layout needs
  more than a static list of paths.

To add a tool, add a concrete scanner in `internal/registry/`, register it in
`BuiltinScanners`, and add tests or documentation as needed. The generic
filesystem traversal remains in `internal/scanner`, so tool-specific path
knowledge stays isolated and new contributions do not require changes to the
scan pipeline.

Built-in platform-specific paths are declared with `platform_paths` using
`darwin`, `windows`, and `linux` keys. Path templates can use `{home}`,
`{config}`, `{data}`, `{cache}`, and `{state}`; they are expanded to absolute
paths before scanning. The older `macos_paths` key remains supported for
existing overrides.

The registry override is stored at `~/.config/aisweep/registry.yaml` on macOS,
`$XDG_CONFIG_HOME/aisweep/registry.yaml` or `~/.config/aisweep/registry.yaml`
on Linux, and `%APPDATA%\aisweep\registry.yaml` on Windows. Snapshots use
`AISWEEP_DATA` when set, otherwise macOS `~/.local/share/aisweep/snapshots`,
Linux `$XDG_DATA_HOME/aisweep/snapshots` or `~/.local/share/aisweep/snapshots`,
and Windows `%LOCALAPPDATA%\aisweep\snapshots`.

Each path declares a `category` (cache, snapshots, sessions, logs, transcripts,
models, config, auth, unknown) and a `risk` (safe, archive, manual, never).

The built-in registry currently covers Claude Code, Codex CLI, Qoder, Kiro,
Cline, Gemini CLI, Cursor, Windsurf, TRAE, Antigravity, GitHub Copilot CLI,
Hugging Face, Continue, llm, OpenAI Desktop, ChatGPT Desktop, LM Studio,
Ollama, and WorkBuddy. Qoder, Kiro, Cline, Gemini CLI, and Copilot CLI honor
their documented home-directory environment variables when present. WorkBuddy
uses a finite set of discovered application and log locations for each platform
because its official help flow does not publish one stable absolute path.

Override paths are merged on top of the builtin (matched by tool `id`); only
non-empty fields win. New tool ids are appended.

## Data

Snapshots are JSON files under the platform data directory described above.
Override with `AISWEEP_DATA`. No database, no telemetry.

## Phase 2

Cleanup is intentionally deferred. Planned: dry-run + selective archive by
`risk` and `category`, gated on the cleanup registry.
