package render

import (
	"fmt"
	"strings"

	"github.com/justincampbell/gh-statusline/internal/pr"
)

// CI returns a colored CI indicator: ✓ / ✗ / ● / "".
func (m Mode) CI(status string) string {
	switch status {
	case "passed":
		return m.green("✓")
	case "failed":
		return m.red("✗")
	case "pending":
		return m.yellow("●")
	default:
		return ""
	}
}

// MergeIndicator returns a single-character mergeability hint:
//
//	red "!"     — branch has merge conflicts (wins over everything below)
//	red "✕"     — merge queue entry is UNMERGEABLE or LOCKED
//	magenta "»" — at position 1 in the merge queue (next up)
//	green "»"   — in the merge queue, position > 1
//	yellow "»"  — auto-merge armed, not yet in the queue
//	""          — none of the above
func (m Mode) MergeIndicator(s *pr.State) string {
	if s == nil {
		return ""
	}
	if s.Conflicting() {
		return m.red("!")
	}
	if s.MergeQueueBroken() {
		return m.red("✕")
	}
	if s.InMergeQueue() {
		if s.MergeQueuePosition == 1 {
			return m.magenta("»")
		}
		return m.green("»")
	}
	if s.AutoMerge {
		return m.yellow("»")
	}
	return ""
}

// Number renders the PR number as a colored, hyperlinked "#NNN".
// Color reflects the review decision (or merged/draft).
func (m Mode) Number(s *pr.State) string {
	if s == nil || s.Number == 0 {
		return ""
	}
	label := fmt.Sprintf("#%d", s.Number)
	colored := m.numberColor(s, label)
	return m.hyperlink(colored, s.URL)
}

func (m Mode) numberColor(s *pr.State, text string) string {
	if s.State == "MERGED" {
		return m.magenta(text)
	}
	if s.IsDraft {
		return m.gray(text)
	}
	switch s.ReviewDecision {
	case "APPROVED":
		return m.green(text)
	case "CHANGES_REQUESTED":
		return m.red(text)
	default:
		return m.yellow(text)
	}
}

// CommentIndicator returns a cyan count when there are unresolved review
// threads, otherwise an empty string.
func (m Mode) CommentIndicator(n int) string {
	if n == 0 {
		return ""
	}
	return m.cyan(fmt.Sprintf("%d", n))
}

// AuthorTag returns "@login" colored magenta when it's not the viewer,
// dimmed when it is. Empty if author is unknown.
func (m Mode) AuthorTag(author, viewer string) string {
	if author == "" {
		return ""
	}
	tag := "@" + author
	if viewer != "" && author == viewer {
		return m.dim(tag)
	}
	return m.magenta(tag)
}

// RepoStatus renders a dimmed, hyperlinked "owner/repo". A failing CI rollup
// for the current branch is prepended (no separator); passing/pending/none
// stays quiet so the prompt only lights up when something actually needs
// attention.
func (m Mode) RepoStatus(bs *pr.BranchStatus) string {
	if bs == nil {
		return ""
	}
	repo := m.hyperlink(m.dim(bs.Owner+"/"+bs.Repo), bs.URL)
	if bs.CIStatus != "failed" {
		return repo
	}
	return m.CI(bs.CIStatus) + repo
}

// Labels returns space-joined, hex-colored label tags.
func (m Mode) Labels(labels []pr.Label) string {
	if len(labels) == 0 {
		return ""
	}
	var parts []string
	for _, l := range labels {
		parts = append(parts, m.hex(l.Name, l.Color))
	}
	return strings.Join(parts, " ")
}
