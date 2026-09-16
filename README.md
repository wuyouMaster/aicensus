# aisweep

Read-only inventory + trend tracker for AI CLI / app storage on macOS.

Phase 1: scan -> classify -> trend. No deletion yet (phase 2).

## Build

    go build -o bin/aisweep ./cmd/aisweep

## Quick start

    aisweep serve --interval=1h        # scan + serve http://127.0.0.1:7890
    # click "scan now" on the dashboard for an on-demand rescan

## Subcommands

    aisweep scan     # one scan, writes ~/.local/share/aisweep/snapshots/<id>.json
    aisweep serve    # local dashboard at http://127.0.0.1:7890
    aisweep path     # print effective registry
    aisweep doctor   # show which registered paths exist on this host

Flags:

    aisweep serve --port 7890 --host 127.0.0.1
    aisweep serve --interval=1h         # background rescan every 1h
    aisweep serve --scan=false          # skip the initial scan on startup

Override data dir with `AISWEEP_DATA=/path/to/snapshots`.

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

Builtin: `internal/registry/registry.yaml`. 15 tools, ~40 paths, all observed
on the author's machine. Override at `~/.config/aisweep/registry.yaml` or
point `AISWEEP_REGISTRY` at a file of the same shape.

Each path declares a `category` (cache, snapshots, sessions, logs, transcripts,
models, config, auth, unknown) and a `risk` (safe, archive, manual, never).

Override paths are merged on top of the builtin (matched by tool `id`); only
non-empty fields win. New tool ids are appended.

## Data

Snapshots are JSON files under `~/.local/share/aisweep/snapshots/`. Override
with `AISWEEP_DATA`. No database, no telemetry.

## Phase 2

Cleanup is intentionally deferred. Planned: dry-run + selective archive by
`risk` and `category`, gated on the cleanup registry.
