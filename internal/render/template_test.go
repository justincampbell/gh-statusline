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

func TestRenderInvalidTemplate(t *testing.T) {
	s := &pr.State{Number: 42, State: "OPEN"}
	if _, err := Render(s, plain(), "{{.bogus"); err == nil {
		t.Error("expected parse error")
	}
}
