package render

import (
	"bytes"
	"fmt"
	"strings"
	"text/template"

	"github.com/justincampbell/gh-statusline/internal/pr"
)

// DefaultTemplate renders, byte-for-byte, the same statusline that
// `recon statusline pr` emits.
const DefaultTemplate = `{{join " " .commentIndicator .ciGroup .authorTag .labelTags}}`

// Render applies tpl (defaults to DefaultTemplate) to the PR state.
// Returns "" when there is no PR for the current branch.
func Render(s *pr.State, mode Mode, tpl string) (string, error) {
	if s == nil || s.Number == 0 {
		return "", nil
	}
	if tpl == "" {
		tpl = DefaultTemplate
	}

	t, err := template.New("statusline").Funcs(funcMap()).Parse(tpl)
	if err != nil {
		return "", fmt.Errorf("parse template: %w", err)
	}

	var buf bytes.Buffer
	if err := t.Execute(&buf, dataFor(s, mode)); err != nil {
		return "", fmt.Errorf("execute template: %w", err)
	}
	return buf.String(), nil
}

func dataFor(s *pr.State, m Mode) map[string]any {
	return map[string]any{
		// raw fields
		"number":             s.Number,
		"title":              s.Title,
		"url":                s.URL,
		"state":              s.State,
		"isDraft":            s.IsDraft,
		"mergeable":          s.Mergeable,
		"reviewDecision":     s.ReviewDecision,
		"autoMerge":          s.AutoMerge,
		"author":             s.Author,
		"labels":             s.Labels,
		"unresolvedComments": s.UnresolvedComments,
		"ciStatus":           s.CIStatus,
		// pre-rendered helpers
		"ci":               m.CI(s.CIStatus),
		"midIndicator":     m.MidIndicator(s),
		"prLink":           m.Number(s),
		"commentIndicator": m.CommentIndicator(s.UnresolvedComments),
		"authorTag":        m.AuthorTag(s.Author, s.Viewer),
		"labelTags":        m.Labels(s.Labels),
		"ciGroup":          ciGroup(m, s),
	}
}

// ciGroup renders CI + midIndicator + prLink with the spacing recon uses:
//
//	"✓ #42"   (CI present, no mid indicator)
//	"✓!#42"   (CI + conflict marker, no space)
//	"!#42"    (mid indicator only)
//	"#42"     (just the link)
func ciGroup(m Mode, s *pr.State) string {
	ci := m.CI(s.CIStatus)
	mid := m.MidIndicator(s)
	link := m.Number(s)

	sep := ""
	if ci != "" && mid == "" {
		sep = " "
	}
	return ci + sep + mid + link
}

func funcMap() template.FuncMap {
	return template.FuncMap{
		"join": func(sep string, args ...string) string {
			var parts []string
			for _, a := range args {
				if a != "" {
					parts = append(parts, a)
				}
			}
			return strings.Join(parts, sep)
		},
		"lower": strings.ToLower,
		"upper": strings.ToUpper,
	}
}
