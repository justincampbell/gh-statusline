# gh-statusline

A `gh` CLI extension that prints a compact, colored status line for the
current branch's PR. See README.md for usage, flags, template fields, and
output behavior.

## Architecture

- `main.go` — CLI entry point (cobra). Root command + `pr` subcommand share the same handler; bare `gh statusline` runs `pr`.
- `internal/pr/` — PR state struct and GraphQL fetcher. Uses go-gh's `pkg/api.GraphQLClient` (direct HTTPS POST, no `gh` fork); auth is inherited from `gh` config. Accepts a `context.Context` so callers can impose a deadline.
- `internal/render/` — ANSI color helpers, `Mode` (color/hyperlink gating via flags + `NO_COLOR` env var, but **not** TTY detection — statusline consumers capture stdout yet still expect ANSI escapes), template engine, and pre-rendered helpers (`ci`, `mergeIndicator`, `prLink`, `authorTag`, `labelTags`, `ciGroup`).
- `internal/cache/` — Tiny file-based output cache at `$TMPDIR/gh-statusline/`, keyed by SHA256 of the current working directory.

## Key design decisions

- One-shot CLI: every invocation reads/writes the cache and returns. No daemon, no in-memory state between calls.
- Cache the **rendered output**, not the parsed PR state. Statuslines call us every few seconds; the cheapest path is "read string from disk, print it".
- On API failure, fall back to the **last cached value regardless of age**. A statusline must never break the prompt.
- Empty output is a valid result (no PR for this branch, no network, not in a repo) and always exits 0.
- No `--json` flag — raw structured data is already available via `gh pr view --json …`. This extension's value is the formatted, cached statusline.

## Dependencies

- `cli/go-gh/v2` — `pkg/api.GraphQLClient` for the GraphQL call, `repository.Current()` for owner/repo
- `spf13/cobra` — CLI structure

Auth is inherited from `gh` automatically.

## Development

Install as a local extension via symlink — `gh extension list` shows the local
copy, `gh statusline` invokes it from this dir, no GitHub fetch:

```
go build -o ./gh-statusline . && gh extension install .
```

**Always reinstall after changes** so `gh statusline` picks up the new
binary in this and every other worktree:

```
go build -o ./gh-statusline . && gh extension install .
```

The second `gh extension install .` is a no-op against the existing symlink
("already an installed extension that provides the 'statusline' command")
but is safe to run every time — it guarantees the symlink exists when
working in a fresh clone.

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
