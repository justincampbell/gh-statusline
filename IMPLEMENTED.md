# Template fields

User-facing capabilities for `--template`. Checked items are exposed today;
unchecked items are candidates we could add.

- [ ] Age — created / updated (e.g., `2d` since opened, `1h` since last update)
- [ ] Assigned users
- [x] Author — `.author` (login), `.authorTag` (colored `@login`, dim when it's you)
- [x] Auto-merge enabled — `.autoMerge` (bool), `.mergeIndicator` (`»`)
- [ ] Auto-merge method — squash / merge / rebase
- [x] CI status — `.ciStatus` (`passed`/`failed`/`pending`/`none`), `.ci` (glyph)
- [ ] Changed file count
- [ ] Commit count
- [ ] Commit SHA (short, from head ref)
- [ ] Diff size — additions / deletions (`+123/-45`)
- [x] Draft state — `.isDraft`
- [ ] Fork-PR indicator (cross-repository)
- [x] Labels — `.labels` (raw), `.labelTags` (hex-colored)
- [ ] Linked closing issues (e.g., `→ #100`)
- [ ] Locked conversation indicator
- [ ] Merge state details — granular beyond `CONFLICTING` (`BEHIND`, `BLOCKED`, `DIRTY`, `UNSTABLE`, …)
- [ ] Merge-queue indicator
- [x] Mergeable / conflict — `.mergeable`, `.mergeIndicator` (`!` when conflicting)
- [ ] Merged-at timestamp / merged-by user (only meaningful when `state == MERGED`)
- [ ] Milestone title
- [x] PR number — `.number` (raw), `.prLink` (colored, hyperlinked `#NNN`)
- [x] PR state — `.state` (`OPEN`/`MERGED`/`CLOSED`)
- [x] PR title — `.title`
- [x] PR URL — `.url`
- [ ] Pending requested reviewers ("waiting on @x")
- [ ] Per-reviewer latest review states
- [ ] Reactions on the PR body
- [x] Review decision — `.reviewDecision` (`APPROVED`/`CHANGES_REQUESTED`/`REVIEW_REQUIRED`)
- [ ] Source branch name
- [ ] Target branch name (e.g., when not the repo default)
- [ ] Total comment count (all conversation comments)
- [ ] Total review thread count (resolved + unresolved)
- [x] Unresolved review threads — `.unresolvedComments` (count), `.commentIndicator` (cyan count)
- [ ] Unread indicator
- [ ] Viewer-authored boolean (cleaner than `eq .author .viewer`)
- [ ] Viewer can merge / update branch / enable auto-merge
- [ ] Viewer's own latest review state
- [ ] Watching / subscription state
