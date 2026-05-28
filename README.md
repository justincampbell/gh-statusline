# gh-statusline

A [gh](https://cli.github.com/) CLI extension that prints a compact, colored
status line for the current branch's PR — suitable for shell prompts, tmux
statuslines, and the Claude Code statusline.

```
2 ✓ #42 @justin bug
```

(unresolved comments • CI status • PR number with link • author • labels)

## Installation

```
gh extension install justincampbell/gh-statusline
```

## Usage

```
gh statusline [flags]
gh statusline pr [flags]
gh statusline fields [flags]
```

Bare `gh statusline` is shorthand for the `pr` subcommand, which prints the
GitHub PR for the current branch. The `fields` subcommand prints every
template variable alongside its current value — useful for designing custom
`--template` strings.

### Flags

| Flag | Default | Description |
|------|---------|-------------|
| `--template <go-template>` | (built-in) | Custom Go template for output |
| `--cache <duration>` | `30s` | TTL for the file-based output cache (`0` disables) |
| `--timeout <duration>` | `3s` | Deadline for the GraphQL request before falling back to stale cache (`0` disables) |
| `--no-color` | `false` | Strip ANSI colors (also honors the `NO_COLOR` env var) |
| `--no-hyperlinks` | `false` | Strip OSC 8 hyperlinks |

### Examples

Default statusline:

```
$ gh statusline
2 ✓ #42 @justin bug
```

Just the PR number:

```
$ gh statusline --template '{{.prLink}}'
#42
```

Pipe-safe machine output:

```
$ gh statusline --no-color --no-hyperlinks
2 ✓ #42 @justin bug
```

Embed in a tmux statusline:

```tmux
set -g status-right '#(cd #{pane_current_path} && gh statusline)'
```

For raw structured data, use `gh` itself:

```
$ gh pr view --json number,title,state,statusCheckRollup
```

## Template Fields

The `--template` flag accepts a [Go text/template](https://pkg.go.dev/text/template).
Two sets of fields are available: raw PR data, and pre-rendered helpers that
apply colors and OSC 8 hyperlinks the way the default template does.

### Raw fields

| Field | Type | Example |
|-------|------|---------|
| `.number` | `int` | `42` |
| `.title` | `string` | `"Fix login redirect"` |
| `.url` | `string` | `"https://github.com/o/r/pull/42"` |
| `.state` | `string` | `"OPEN"` / `"MERGED"` / `"CLOSED"` |
| `.isDraft` | `bool` | |
| `.mergeable` | `string` | `"MERGEABLE"` / `"CONFLICTING"` / `"UNKNOWN"` |
| `.reviewDecision` | `string` | `"APPROVED"` / `"CHANGES_REQUESTED"` / `"REVIEW_REQUIRED"` |
| `.autoMerge` | `bool` | Auto-merge requested |
| `.mergeQueueState` | `string` | `"QUEUED"` / `"AWAITING_CHECKS"` / `"MERGEABLE"` / `"UNMERGEABLE"` / `"LOCKED"` / `""` |
| `.mergeQueuePosition` | `int` | 1 = next up; 0 when not in queue |
| `.author` | `string` | `"justin"` |
| `.labels` | `[]Label` | `[{Name, Color}, …]` |
| `.unresolvedComments` | `int` | `2` |
| `.ciStatus` | `string` | `"passed"` / `"failed"` / `"pending"` / `"none"` |

### Pre-rendered helpers

| Helper | Output |
|--------|--------|
| `.ci` | Colored `✓` / `✗` / `●` / empty |
| `.mergeIndicator` | Red `!` (conflict), red `✕` (queue rejected), magenta `»` (next up in merge queue), green `»` (in merge queue), yellow `»` (auto-merge armed), or empty |
| `.prLink` | `#42` colored by review decision, wrapped in an OSC 8 hyperlink |
| `.commentIndicator` | Cyan unresolved-comments count, or empty |
| `.authorTag` | `@author`, magenta when it's not you, dim when it is |
| `.labelTags` | Space-joined, hex-colored label names |
| `.ciGroup` | `.ci` + `.mergeIndicator` + `.prLink` with the spacing the default template uses |

### Template functions

| Function | Description |
|----------|-------------|
| `join sep args…` | Joins non-empty strings with `sep` (skips empties) |
| `lower s` / `upper s` | Case conversion |

### Default template

```
{{join " " .commentIndicator .ciGroup .authorTag .labelTags}}
```

### Inspecting field values

```
$ gh statusline fields
ci                ✓
mergeIndicator    
prLink            #42
commentIndicator  2
authorTag         @justin
labelTags         bug
ciGroup           ✓ #42
```

## Output

When the current branch has no PR, output is a single empty line and the exit
code is `0` — a statusline must never break the prompt.

When there is no PR for the current branch — on the repo's default branch
(e.g. `main`), or on a feature branch that hasn't been opened yet — the
statusline falls back to a dim, hyperlinked `owner/repo`. A failing CI rollup
is prepended (no separator) so the prompt only lights up when something
actually needs attention:

```
$ gh statusline           # default branch / no PR, CI green
owner/repo

$ gh statusline           # default branch / no PR, CI failing
✗owner/repo
```

On a GitHub API failure (network, rate limit, transient error, or hitting
`--timeout`), the last cached output is reprinted regardless of its age. If no
cached value exists, output is empty and the exit code is still `0`.

The cache lives at `$TMPDIR/gh-statusline/pr-<sha>.json`, keyed by the current
working directory.

## Development

Build and install from local source:

```
go build -o ./gh-statusline . && gh extension install .
```

Run after making changes:

```
go build -o ./gh-statusline . && ./gh-statusline
```

Run tests:

```
go test ./...
```

## Releasing

Push a version tag — the release workflow cross-compiles binaries for all
platforms via `cli/gh-extension-precompile`:

```
git tag v0.1.0
git push origin v0.1.0
```

Tags with hyphens (e.g. `v0.1.0-rc.1`) are published as prereleases.

## License

MIT
