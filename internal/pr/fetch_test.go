package pr

import (
	"encoding/json"
	"testing"
)

func TestMapRollupState(t *testing.T) {
	tests := []struct {
		state string
		want  string
	}{
		{"SUCCESS", "passed"},
		{"FAILURE", "failed"},
		{"ERROR", "failed"},
		{"PENDING", "pending"},
		{"EXPECTED", "pending"},
		{"", "none"},
		{"WHATEVER", "none"},
	}
	for _, tt := range tests {
		t.Run(tt.state, func(t *testing.T) {
			if got := mapRollupState(tt.state); got != tt.want {
				t.Errorf("mapRollupState(%q) = %q, want %q", tt.state, got, tt.want)
			}
		})
	}
}

func TestBuildEmpty(t *testing.T) {
	raw := `{"viewer":{"login":"alice"},"repository":{"pullRequests":{"nodes":[]}}}`
	var data gqlData
	if err := json.Unmarshal([]byte(raw), &data); err != nil {
		t.Fatal(err)
	}
	s := build(&data)
	if s.Number != 0 {
		t.Errorf("expected Number=0, got %d", s.Number)
	}
	if s.Viewer != "alice" {
		t.Errorf("expected Viewer=alice, got %q", s.Viewer)
	}
}

func TestBuildPrefersOpenOverMerged(t *testing.T) {
	raw := `{"viewer":{"login":"alice"},"repository":{"pullRequests":{"nodes":[
		{"number":1,"state":"MERGED"},
		{"number":2,"state":"OPEN"}
	]}}}`
	var data gqlData
	if err := json.Unmarshal([]byte(raw), &data); err != nil {
		t.Fatal(err)
	}
	s := build(&data)
	if s.Number != 2 {
		t.Errorf("expected OPEN PR #2 to win, got %d", s.Number)
	}
}

func TestBuildUnresolvedCommentsCount(t *testing.T) {
	raw := `{"viewer":{"login":"alice"},"repository":{"pullRequests":{"nodes":[
		{"number":42,"state":"OPEN","reviewThreads":{"nodes":[
			{"isResolved":true},{"isResolved":false},{"isResolved":false}
		]}}
	]}}}`
	var data gqlData
	if err := json.Unmarshal([]byte(raw), &data); err != nil {
		t.Fatal(err)
	}
	s := build(&data)
	if s.UnresolvedComments != 2 {
		t.Errorf("expected 2 unresolved, got %d", s.UnresolvedComments)
	}
}

func TestBuildCIStatusFromRollup(t *testing.T) {
	raw := `{"viewer":{"login":"alice"},"repository":{"pullRequests":{"nodes":[
		{"number":42,"state":"OPEN","commits":{"nodes":[
			{"commit":{"statusCheckRollup":{"state":"FAILURE"}}}
		]}}
	]}}}`
	var data gqlData
	if err := json.Unmarshal([]byte(raw), &data); err != nil {
		t.Fatal(err)
	}
	s := build(&data)
	if s.CIStatus != "failed" {
		t.Errorf("expected failed, got %q", s.CIStatus)
	}
}

func TestBuildCIStatusNoRollup(t *testing.T) {
	raw := `{"viewer":{"login":"alice"},"repository":{"pullRequests":{"nodes":[
		{"number":42,"state":"OPEN","commits":{"nodes":[
			{"commit":{"statusCheckRollup":null}}
		]}}
	]}}}`
	var data gqlData
	if err := json.Unmarshal([]byte(raw), &data); err != nil {
		t.Fatal(err)
	}
	s := build(&data)
	if s.CIStatus != "none" {
		t.Errorf("expected none, got %q", s.CIStatus)
	}
}
