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
	raw := `{"viewer":{"login":"alice"},"repository":{"url":"https://github.com/o/r","pullRequests":{"nodes":[]}}}`
	var data gqlData
	if err := json.Unmarshal([]byte(raw), &data); err != nil {
		t.Fatal(err)
	}
	r := build("o", "r", &data)
	if r.PR.Number != 0 {
		t.Errorf("expected PR.Number=0, got %d", r.PR.Number)
	}
	if r.PR.Viewer != "alice" {
		t.Errorf("expected Viewer=alice, got %q", r.PR.Viewer)
	}
	if r.Branch.URL != "https://github.com/o/r" {
		t.Errorf("Branch.URL=%q", r.Branch.URL)
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
	r := build("o", "r", &data)
	if r.PR.Number != 2 {
		t.Errorf("expected OPEN PR #2 to win, got %d", r.PR.Number)
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
	r := build("o", "r", &data)
	if r.PR.UnresolvedComments != 2 {
		t.Errorf("expected 2 unresolved, got %d", r.PR.UnresolvedComments)
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
	r := build("o", "r", &data)
	if r.PR.CIStatus != "failed" {
		t.Errorf("expected failed, got %q", r.PR.CIStatus)
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
	r := build("o", "r", &data)
	if r.PR.CIStatus != "none" {
		t.Errorf("expected none, got %q", r.PR.CIStatus)
	}
}

func TestBuildMergeQueueEntry(t *testing.T) {
	raw := `{"viewer":{"login":"alice"},"repository":{"pullRequests":{"nodes":[
		{"number":42,"state":"OPEN","mergeQueueEntry":{"position":3,"state":"QUEUED"}}
	]}}}`
	var data gqlData
	if err := json.Unmarshal([]byte(raw), &data); err != nil {
		t.Fatal(err)
	}
	r := build("o", "r", &data)
	if !r.PR.InMergeQueue() {
		t.Error("expected InMergeQueue=true")
	}
	if r.PR.MergeQueueState != "QUEUED" {
		t.Errorf("MergeQueueState=%q, want QUEUED", r.PR.MergeQueueState)
	}
	if r.PR.MergeQueuePosition != 3 {
		t.Errorf("MergeQueuePosition=%d, want 3", r.PR.MergeQueuePosition)
	}
	if r.PR.MergeQueueBroken() {
		t.Error("expected MergeQueueBroken=false for QUEUED")
	}
}

func TestBuildMergeQueueBroken(t *testing.T) {
	raw := `{"viewer":{"login":"alice"},"repository":{"pullRequests":{"nodes":[
		{"number":42,"state":"OPEN","mergeQueueEntry":{"position":1,"state":"UNMERGEABLE"}}
	]}}}`
	var data gqlData
	if err := json.Unmarshal([]byte(raw), &data); err != nil {
		t.Fatal(err)
	}
	r := build("o", "r", &data)
	if !r.PR.MergeQueueBroken() {
		t.Error("expected MergeQueueBroken=true for UNMERGEABLE")
	}
}

func TestBuildNoMergeQueueEntry(t *testing.T) {
	raw := `{"viewer":{"login":"alice"},"repository":{"pullRequests":{"nodes":[
		{"number":42,"state":"OPEN","mergeQueueEntry":null}
	]}}}`
	var data gqlData
	if err := json.Unmarshal([]byte(raw), &data); err != nil {
		t.Fatal(err)
	}
	r := build("o", "r", &data)
	if r.PR.InMergeQueue() {
		t.Error("expected InMergeQueue=false when entry is null")
	}
}

func TestBuildBranchSuccess(t *testing.T) {
	raw := `{"repository":{"url":"https://github.com/o/r","ref":{"target":{"statusCheckRollup":{"state":"SUCCESS"}}},"pullRequests":{"nodes":[]}}}`
	var data gqlData
	if err := json.Unmarshal([]byte(raw), &data); err != nil {
		t.Fatal(err)
	}
	r := build("o", "r", &data)
	if r.Branch.CIStatus != "passed" {
		t.Errorf("Branch.CIStatus=%q, want passed", r.Branch.CIStatus)
	}
	if r.Branch.Owner != "o" || r.Branch.Repo != "r" {
		t.Errorf("owner/repo=%q/%q", r.Branch.Owner, r.Branch.Repo)
	}
}

func TestBuildBranchNoRef(t *testing.T) {
	raw := `{"repository":{"url":"https://github.com/o/r","ref":null,"pullRequests":{"nodes":[]}}}`
	var data gqlData
	if err := json.Unmarshal([]byte(raw), &data); err != nil {
		t.Fatal(err)
	}
	r := build("o", "r", &data)
	if r.Branch.CIStatus != "none" {
		t.Errorf("Branch.CIStatus=%q, want none", r.Branch.CIStatus)
	}
}

func TestBuildBranchURLFallback(t *testing.T) {
	raw := `{"repository":{"ref":{"target":{"statusCheckRollup":{"state":"FAILURE"}}},"pullRequests":{"nodes":[]}}}`
	var data gqlData
	if err := json.Unmarshal([]byte(raw), &data); err != nil {
		t.Fatal(err)
	}
	r := build("o", "r", &data)
	if r.Branch.URL != "https://github.com/o/r" {
		t.Errorf("URL fallback=%q", r.Branch.URL)
	}
	if r.Branch.CIStatus != "failed" {
		t.Errorf("Branch.CIStatus=%q, want failed", r.Branch.CIStatus)
	}
}
