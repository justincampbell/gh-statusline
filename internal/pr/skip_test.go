package pr

import "testing"

func TestShouldSkip(t *testing.T) {
	tests := []struct {
		name          string
		branch        string
		defaultBranch string
		want          bool
	}{
		{"on default", "main", "main", true},
		{"on feature", "feature-x", "main", false},
		{"default unknown", "main", "", false},
		{"master default", "master", "master", true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := ShouldSkip(tt.branch, tt.defaultBranch); got != tt.want {
				t.Errorf("ShouldSkip(%q, %q) = %v, want %v", tt.branch, tt.defaultBranch, got, tt.want)
			}
		})
	}
}
