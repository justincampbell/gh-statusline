# gh-statusline

A `gh` CLI extension that prints a compact, colored status line for the
current branch's PR. See README.md for usage, flags, template fields, and
output behavior.

## Architecture

- `main.go` — CLI entry point (cobra). Root command + `pr` subcommand share the same handler; bare `gh statusline` runs `pr`.
- `internal/pr/` — PR state struct and GraphQL fetcher. Uses go-gh's `gh.Exec` to call `gh api graphql`; auth is inherited from `gh`.
- `internal/render/` — ANSI color helpers, `Mode` (color/hyperlink gating from TTY / `NO_COLOR`), template engine, and pre-rendered helpers (`ci`, `midIndicator`, `prLink`, `authorTag`, `labelTags`, `ciGroup`).
- `internal/cache/` — Tiny file-based output cache at `$TMPDIR/gh-statusline/`, keyed by SHA256 of the current working directory.

## Key design decisions

- One-shot CLI: every invocation reads/writes the cache and returns. No daemon, no in-memory state between calls.
- Cache the **rendered output**, not the parsed PR state. Statuslines call us every few seconds; the cheapest path is "read string from disk, print it".
- On API failure, fall back to the **last cached value regardless of age**. A statusline must never break the prompt.
- Empty output is a valid result (no PR for this branch, no network, not in a repo) and always exits 0.
- No `--json` flag — raw structured data is already available via `gh pr view --json …`. This extension's value is the formatted, cached statusline.

## Dependencies

- `cli/go-gh/v2` — `gh.Exec` for the GraphQL call, `repository.Current()` for owner/repo
- `spf13/cobra` — CLI structure
- `golang.org/x/term` — TTY detection for color/hyperlink gating

Auth is inherited from `gh` automatically.

## Development

Install as a local extension via symlink — `gh extension list` shows the local
copy, `gh statusline` invokes it from this dir, no GitHub fetch:

```
go build -o ./gh-statusline . && gh extension install .
```

Rebuild after changes:

```
go build -o ./gh-statusline .
```

## Testing

```
go test ./...
```

Tests live in `internal/pr/` (parsing, CI rollup), `internal/render/` (default
and custom templates, conflict/auto-merge grouping), and `internal/cache/`
(file roundtrip, stale fallback).

## Releasing

Tag-triggered via `cli/gh-extension-precompile`:

```
git tag vX.Y.Z && git push origin vX.Y.Z
```
