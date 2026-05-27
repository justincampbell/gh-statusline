package pr

// BranchStatus is the minimal info shown when there is no PR to display for
// the current branch: the owner/repo link plus the branch tip's CI rollup.
type BranchStatus struct {
	Owner    string
	Repo     string
	URL      string
	CIStatus string
}
