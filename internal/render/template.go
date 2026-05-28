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

// Field is one template variable with its rendered value, used by the
// `fields` subcommand to enumerate everything a template can reference.
type Field struct {
	Name     string
	Value    string
	Category string // "raw" or "helper"
}

// Fields returns every template variable and its current value, in stable
// display order: raw PR fields first, then pre-rendered helpers.
func Fields(s *pr.State, m Mode) []Field {
	if s == nil {
		s = &pr.State{}
	}
	data := dataFor(s, m)

	order := []struct{ name, category string }{
		{"number", "raw"},
		{"title", "raw"},
		{"url", "raw"},
		{"state", "raw"},
		{"isDraft", "raw"},
		{"mergeable", "raw"},
		{"reviewDecision", "raw"},
		{"autoMerge", "raw"},
		{"mergeQueueState", "raw"},
		{"mergeQueuePosition", "raw"},
		{"author", "raw"},
		{"labels", "raw"},
		{"unresolvedComments", "raw"},
		{"ciStatus", "raw"},
		{"ci", "helper"},
		{"mergeIndicator", "helper"},
		{"prLink", "helper"},
		{"commentIndicator", "helper"},
		{"authorTag", "helper"},
		{"labelTags", "helper"},
		{"ciGroup", "helper"},
	}

	out := make([]Field, 0, len(order))
	for _, o := range order {
		out = append(out, Field{
			Name:     o.name,
			Value:    formatValue(data[o.name]),
			Category: o.category,
		})
	}
	return out
}

func formatValue(v any) string {
	switch x := v.(type) {
	case string:
		return x
	case int:
		return fmt.Sprintf("%d", x)
	case bool:
		return fmt.Sprintf("%t", x)
	case []pr.Label:
		if len(x) == 0 {
			return "[]"
		}
		names := make([]string, 0, len(x))
		for _, l := range x {
			names = append(names, l.Name)
		}
		return "[" + strings.Join(names, " ") + "]"
	default:
		return fmt.Sprintf("%v", v)
	}
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
		"mergeQueueState":    s.MergeQueueState,
		"mergeQueuePosition": s.MergeQueuePosition,
		"author":             s.Author,
		"labels":             s.Labels,
		"unresolvedComments": s.UnresolvedComments,
		"ciStatus":           s.CIStatus,
		// pre-rendered helpers
		"ci":               m.CI(s.CIStatus),
		"mergeIndicator":   m.MergeIndicator(s),
		"prLink":           m.Number(s),
		"commentIndicator": m.CommentIndicator(s.UnresolvedComments),
		"authorTag":        m.AuthorTag(s.Author, s.Viewer),
		"labelTags":        m.Labels(s.Labels),
		"ciGroup":          ciGroup(m, s),
	}
}

// ciGroup renders CI + mergeIndicator + prLink with the spacing recon uses:
//
//	"✓ #42"   (CI present, no merge indicator)
//	"✓!#42"   (CI + conflict marker, no space)
//	"!#42"    (merge indicator only)
//	"#42"     (just the link)
func ciGroup(m Mode, s *pr.State) string {
	ci := m.CI(s.CIStatus)
	merge := m.MergeIndicator(s)
	link := m.Number(s)

	sep := ""
	if ci != "" && merge == "" {
		sep = " "
	}
	return ci + sep + merge + link
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
