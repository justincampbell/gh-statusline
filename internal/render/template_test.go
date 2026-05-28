package render

import (
	"strings"
	"testing"

	"github.com/justincampbell/gh-statusline/internal/pr"
)

func plain() Mode { return Mode{NoColor: true, NoHyperlinks: true} }

func TestRenderNoPR(t *testing.T) {
	got, err := Render(nil, plain(), "")
	if err != nil {
		t.Fatal(err)
	}
	if got != "" {
		t.Errorf("expected empty, got %q", got)
	}

	got, err = Render(&pr.State{Number: 0}, plain(), "")
	if err != nil {
		t.Fatal(err)
	}
	if got != "" {
		t.Errorf("expected empty for Number=0, got %q", got)
	}
}

func TestRenderDefaultTemplate(t *testing.T) {
	s := &pr.State{
		Number:             42,
		URL:                "https://github.com/o/r/pull/42",
		State:              "OPEN",
		CIStatus:           "passed",
		ReviewDecision:     "APPROVED",
		Author:             "alice",
		Viewer:             "bob",
		UnresolvedComments: 2,
		Labels:             []pr.Label{{Name: "bug", Color: "d73a4a"}},
	}
	got, err := Render(s, plain(), "")
	if err != nil {
		t.Fatal(err)
	}
	want := "2 ✓ #42 @alice bug"
	if got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}

func TestRenderConflictTightens(t *testing.T) {
	s := &pr.State{
		Number:    42,
		State:     "OPEN",
		CIStatus:  "passed",
		Mergeable: "CONFLICTING",
	}
	got, _ := Render(s, plain(), "")
	if !strings.Contains(got, "✓!#42") {
		t.Errorf("expected '✓!#42' grouping in %q", got)
	}
}

func TestRenderAutoMergeIndicator(t *testing.T) {
	s := &pr.State{Number: 42, State: "OPEN", CIStatus: "pending", AutoMerge: true}
	got, _ := Render(s, plain(), "")
	if !strings.Contains(got, "●»#42") {
		t.Errorf("expected '●»#42' grouping in %q", got)
	}
}

func TestMergeIndicatorColors(t *testing.T) {
	colorMode := Mode{NoHyperlinks: true}
	const (
		yellow  = "\033[33m"
		green   = "\033[32m"
		magenta = "\033[35m"
		red     = "\033[31m"
		reset   = "\033[0m"
	)

	tests := []struct {
		name  string
		state *pr.State
		want  string
	}{
		{
			name:  "auto-merge armed (not in queue)",
			state: &pr.State{AutoMerge: true},
			want:  yellow + "»" + reset,
		},
		{
			name:  "in queue, position > 1",
			state: &pr.State{AutoMerge: true, MergeQueueState: "QUEUED", MergeQueuePosition: 3},
			want:  green + "»" + reset,
		},
		{
			name:  "in queue, position 1 = next up",
			state: &pr.State{AutoMerge: true, MergeQueueState: "AWAITING_CHECKS", MergeQueuePosition: 1},
			want:  magenta + "»" + reset,
		},
		{
			name:  "queue unmergeable",
			state: &pr.State{MergeQueueState: "UNMERGEABLE", MergeQueuePosition: 1},
			want:  red + "✕" + reset,
		},
		{
			name:  "queue locked",
			state: &pr.State{MergeQueueState: "LOCKED", MergeQueuePosition: 2},
			want:  red + "✕" + reset,
		},
		{
			name:  "conflict beats everything",
			state: &pr.State{Mergeable: "CONFLICTING", AutoMerge: true, MergeQueueState: "QUEUED", MergeQueuePosition: 1},
			want:  red + "!" + reset,
		},
		{
			name:  "nothing armed",
			state: &pr.State{},
			want:  "",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := colorMode.MergeIndicator(tt.state); got != tt.want {
				t.Errorf("MergeIndicator = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestRenderCustomTemplate(t *testing.T) {
	s := &pr.State{Number: 42, State: "OPEN", Author: "alice"}
	got, err := Render(s, plain(), `{{.author}}/{{.number}}`)
	if err != nil {
		t.Fatal(err)
	}
	if got != "alice/42" {
		t.Errorf("got %q, want %q", got, "alice/42")
	}
}

func TestFieldsPopulated(t *testing.T) {
	s := &pr.State{
		Number:             42,
		Title:              "Fix login redirect",
		URL:                "https://github.com/o/r/pull/42",
		State:              "OPEN",
		IsDraft:            false,
		Mergeable:          "MERGEABLE",
		ReviewDecision:     "APPROVED",
		AutoMerge:          true,
		Author:             "alice",
		Viewer:             "bob",
		UnresolvedComments: 2,
		CIStatus:           "passed",
		Labels:             []pr.Label{{Name: "bug", Color: "d73a4a"}, {Name: "p1", Color: "0075ca"}},
	}
	got := Fields(s, plain())

	want := map[string]string{
		"number":             "42",
		"title":              "Fix login redirect",
		"url":                "https://github.com/o/r/pull/42",
		"state":              "OPEN",
		"isDraft":            "false",
		"mergeable":          "MERGEABLE",
		"reviewDecision":     "APPROVED",
		"autoMerge":          "true",
		"mergeQueueState":    "",
		"mergeQueuePosition": "0",
		"author":             "alice",
		"labels":             "[bug p1]",
		"unresolvedComments": "2",
		"ciStatus":           "passed",
		"ci":                 "✓",
		"mergeIndicator":     "»",
		"prLink":             "#42",
		"commentIndicator":   "2",
		"authorTag":          "@alice",
		"labelTags":          "bug p1",
		"ciGroup":            "✓»#42",
	}

	if len(got) != len(want) {
		t.Fatalf("got %d fields, want %d", len(got), len(want))
	}
	for _, f := range got {
		w, ok := want[f.Name]
		if !ok {
			t.Errorf("unexpected field %q", f.Name)
			continue
		}
		if f.Value != w {
			t.Errorf("field %q: got %q, want %q", f.Name, f.Value, w)
		}
	}

	// Categories are split: raw fields first, helpers second.
	var seenHelper bool
	for _, f := range got {
		if f.Category == "helper" {
			seenHelper = true
		} else if seenHelper {
			t.Errorf("raw field %q appeared after a helper — order should be raw then helper", f.Name)
		}
	}
}

func TestFieldsEmptyState(t *testing.T) {
	got := Fields(nil, plain())
	if len(got) == 0 {
		t.Fatal("expected fields even for nil state")
	}
	for _, f := range got {
		if f.Name == "number" && f.Value != "0" {
			t.Errorf("expected number=0 for nil state, got %q", f.Value)
		}
		if f.Name == "labels" && f.Value != "[]" {
			t.Errorf("expected labels=[] for nil state, got %q", f.Value)
		}
	}
}

func TestRepoStatusFailingShowsXNoSpace(t *testing.T) {
	bs := &pr.BranchStatus{Owner: "o", Repo: "r", URL: "https://github.com/o/r", CIStatus: "failed"}
	got := plain().RepoStatus(bs)
	if got != "✗o/r" {
		t.Errorf("got %q, want %q", got, "✗o/r")
	}
}

func TestRepoStatusOmitsPassing(t *testing.T) {
	for _, status := range []string{"passed", "pending", "none", ""} {
		bs := &pr.BranchStatus{Owner: "o", Repo: "r", URL: "https://github.com/o/r", CIStatus: status}
		got := plain().RepoStatus(bs)
		if got != "o/r" {
			t.Errorf("CIStatus=%q: got %q, want %q", status, got, "o/r")
		}
	}
}

func TestRepoStatusNil(t *testing.T) {
	if got := plain().RepoStatus(nil); got != "" {
		t.Errorf("got %q, want empty", got)
	}
}

func TestRenderInvalidTemplate(t *testing.T) {
	s := &pr.State{Number: 42, State: "OPEN"}
	if _, err := Render(s, plain(), "{{.bogus"); err == nil {
		t.Error("expected parse error")
	}
}
