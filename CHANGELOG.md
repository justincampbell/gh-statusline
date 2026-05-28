# Changelog

All notable changes to this project are documented here. The format is based
on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/), and this project
adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [0.3.0] - 2026-05-28

### Added
- Merge queue awareness. The `»` indicator is now color-coded by state:
  yellow for auto-merge armed (not yet enqueued), green for in the merge
  queue, and magenta for next up (position 1). A red `✕` appears when the
  queue entry is `UNMERGEABLE` or `LOCKED`. Conflict still wins with red `!`.
- New raw template fields `.mergeQueueState`
  (`QUEUED`/`AWAITING_CHECKS`/`MERGEABLE`/`UNMERGEABLE`/`LOCKED`) and
  `.mergeQueuePosition` (1 = next up; 0 when not in queue).

### Changed
- Auto-merge armed (without a queue entry) renders yellow instead of magenta,
  so magenta is reserved for "next up in the queue."

## [0.2.0] - 2026-05-27

### Added
- When there is no PR for the current branch — on the repo's default branch,
  or on a feature branch that hasn't been opened yet — fall back to a dim,
  hyperlinked `owner/repo` instead of empty output. A failing CI rollup is
  prepended (no separator) so the prompt only lights up when something needs
  attention.

### Changed
- Collapsed PR and repo lookups into a single GraphQL query (`pr.Fetch` now
  returns `Result{PR, Branch}`), so no-PR feature branches don't pay an extra
  round trip.
- `defaultBranch()` falls back to a local `main`/`master` lookup when
  `refs/remotes/origin/HEAD` is unset, so the default-branch view works on
  repos that were never `git clone`d (or had their origin/HEAD wiped).
