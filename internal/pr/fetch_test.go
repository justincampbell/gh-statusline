package pr

import "testing"

func TestRollupStatus(t *testing.T) {
	tests := []struct {
		name   string
		checks []StatusCheck
		want   string
	}{
		{"no checks", nil, "none"},
		{"all CheckRun success", []StatusCheck{
			{Typename: "CheckRun", Status: "COMPLETED", Conclusion: "SUCCESS"},
			{Typename: "CheckRun", Status: "COMPLETED", Conclusion: "SUCCESS"},
		}, "passed"},
		{"CheckRun failure beats success", []StatusCheck{
			{Typename: "CheckRun", Status: "COMPLETED", Conclusion: "SUCCESS"},
			{Typename: "CheckRun", Status: "COMPLETED", Conclusion: "FAILURE"},
		}, "failed"},
		{"CheckRun in progress is pending", []StatusCheck{
			{Typename: "CheckRun", Status: "COMPLETED", Conclusion: "SUCCESS"},
			{Typename: "CheckRun", Status: "IN_PROGRESS"},
		}, "pending"},
		{"failed beats pending", []StatusCheck{
			{Typename: "CheckRun", Status: "IN_PROGRESS"},
			{Typename: "CheckRun", Status: "COMPLETED", Conclusion: "FAILURE"},
		}, "failed"},
		{"StatusContext success", []StatusCheck{
			{Typename: "StatusContext", State: "SUCCESS"},
		}, "passed"},
		{"StatusContext pending", []StatusCheck{
			{Typename: "StatusContext", State: "PENDING"},
		}, "pending"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := rollupStatus(tt.checks); got != tt.want {
				t.Errorf("got %q, want %q", got, tt.want)
			}
		})
	}
}

func TestParseEmpty(t *testing.T) {
	raw := []byte(`{"data":{"viewer":{"login":"alice"},"repository":{"pullRequests":{"nodes":[]}}}}`)
	s, err := parse(raw)
	if err != nil {
		t.Fatal(err)
	}
	if s.Number != 0 {
		t.Errorf("expected Number=0, got %d", s.Number)
	}
	if s.Viewer != "alice" {
		t.Errorf("expected Viewer=alice, got %q", s.Viewer)
	}
}

func TestParsePrefersOpenOverMerged(t *testing.T) {
	raw := []byte(`{"data":{"viewer":{"login":"alice"},"repository":{"pullRequests":{"nodes":[
		{"number":1,"state":"MERGED"},
		{"number":2,"state":"OPEN"}
	]}}}}`)
	s, err := parse(raw)
	if err != nil {
		t.Fatal(err)
	}
	if s.Number != 2 {
		t.Errorf("expected OPEN PR #2 to win, got %d", s.Number)
	}
}

func TestParseUnresolvedCommentsCount(t *testing.T) {
	raw := []byte(`{"data":{"viewer":{"login":"alice"},"repository":{"pullRequests":{"nodes":[
		{"number":42,"state":"OPEN","reviewThreads":{"nodes":[
			{"isResolved":true},{"isResolved":false},{"isResolved":false}
		]}}
	]}}}}`)
	s, err := parse(raw)
	if err != nil {
		t.Fatal(err)
	}
	if s.UnresolvedComments != 2 {
		t.Errorf("expected 2 unresolved, got %d", s.UnresolvedComments)
	}
}
