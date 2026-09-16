# AICensus

### AI tools keep crapping all over your hard drive

AI tools keep getting installed one after another. They leave caches, logs,
sessions, snapshots, and models all over your drive, then walk away like
nothing happened. The worst part? The mess hides well, so by the time your disk
is full, you may not even know which tool did it.

AICensus hunts down the leftovers: who is taking space, where they left it, and
whether the mess is still growing. Find it all, then clean it up one tool at a
time.

> AI tools keep crapping. AICensus keeps catching them. Your hard drive needs a cleanup.

> The executable is still named `aisweep` for compatibility with existing commands, environment variables, and snapshots.

[中文 README](README.md) · [GitHub Actions](https://github.com/wuyouMaster/aicensus/actions)

## Features

- Overview of total storage, risk distribution, tools, categories, and snapshots.
- Daily and hourly trend granularity with hover data for each chart point.
- A card for every AI tool with its icon, storage usage, categories, and last scan time.
- Scan guide explaining every path, its category, and how the tool typically uses it.
- macOS, Windows, and Linux support.
- Read-only Phase 1: inventory and trends first; cleanup is planned for Phase 2.
- Local-only by default on `127.0.0.1`; no database and no telemetry upload.
- Nested path deduplication: when a parent is fully covered by registered children, only the child results are displayed.

## Screenshots

These screenshots were captured from the local page content only. They do not include a browser address bar, bookmarks, or browser chrome.

### Overview

![AICensus overview](docs/images/overview.png)

### Tool list

![AICensus tool list](docs/images/tools.png)

### Scan guide

![AICensus scan guide](docs/images/guide.png)

## Quick start

Requires Go 1.22 or later.

```bash
go build -o bin/aisweep ./cmd/aisweep
./bin/aisweep serve --interval=1h
```

Open <http://127.0.0.1:7890>. The overview also has an action for an on-demand scan.

```bash
aisweep scan                         # scan once and write a snapshot
aisweep serve                        # start the local dashboard
aisweep path                         # print the effective path registry
aisweep doctor                       # show paths found on this host
aisweep serve --port 7890            # change the port
aisweep serve --host 127.0.0.1       # change the bind address
aisweep serve --interval=1h          # rescan in the background every hour
aisweep serve --scan=false            # skip the initial scan
```

Override the snapshot directory with `AISWEEP_DATA=/path/to/snapshots`.

## Publishing a release

Push a version tag beginning with `v` to GitHub. Actions will run the tests and
static checks on all three desktop platforms, build amd64 and arm64 packages
for macOS, Linux, and Windows, and create a GitHub Release automatically.

```bash
git tag v0.1.0
git push origin v0.1.0
```

Each release contains platform archives and a `sha256sums.txt` checksum file.
Rerunning the workflow for the same tag updates attachments with the same names.

## Platforms and builds

The scanner resolves platform-specific paths at runtime. GitHub Actions runs
tests and static checks on Ubuntu, macOS, and Windows and builds:

- macOS: amd64, arm64
- Linux: amd64, arm64
- Windows: amd64, arm64

## Scanner architecture

Each AI tool can implement its own `registry.ToolScanner`:

```go
type ToolScanner interface {
    Definition() registry.Tool
    Discover() ([]registry.Entry, error)
}
```

`Definition` provides metadata and default paths. `Discover` resolves actual
paths for the current machine. Generic filesystem traversal remains in
`internal/scanner`, while tool-specific path knowledge lives in
`internal/registry`.

Built-in paths use `platform_paths` with `darwin`, `windows`, and `linux` keys.
Templates support `{home}`, `{config}`, `{data}`, `{cache}`, and `{state}`.
The legacy `macos_paths` key remains supported.

The built-in registry currently covers Claude Code, Codex CLI, Qoder, Kiro,
Cline, Gemini CLI, Cursor, Windsurf, TRAE, Antigravity, GitHub Copilot CLI,
Hugging Face, Continue, llm, OpenAI Desktop, ChatGPT Desktop, LM Studio,
Ollama, and WorkBuddy.

## Data and development

Each path has a category such as cache, snapshot, session, log, transcript,
model, config, credential, or unknown, plus a risk level: safe, archive,
manual, or never. Snapshots are JSON files in the platform data directory.
There is no database and no telemetry upload.

```bash
go test ./...
go vet ./...
go build -o bin/aisweep ./cmd/aisweep
```

Cleanup is intentionally deferred to Phase 2. The planned flow is dry-run
preview followed by selective archive actions gated by the cleanup registry.
