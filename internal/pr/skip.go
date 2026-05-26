package pr

// ShouldSkip reports whether PR lookup should be skipped for the current
// branch. The default branch (e.g. main, master) is virtually never a real
// PR head — matching merged PRs from there surfaces stale results. An empty
// defaultBranch means "unknown" and never skips.
func ShouldSkip(branch, defaultBranch string) bool {
	return defaultBranch != "" && branch == defaultBranch
}
