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

// MidIndicator returns the single-character indicator between CI and PR
// number. Conflict (!) wins over auto-merge (»).
func (m Mode) MidIndicator(s *pr.State) string {
	if s == nil {
		return ""
	}
	if s.Conflicting() {
		return m.red("!")
	}
	if s.AutoMerge {
		return m.magenta("»")
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
