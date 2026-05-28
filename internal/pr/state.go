package pr

type Label struct {
	Name  string `json:"name"`
	Color string `json:"color"`
}

type State struct {
	Number             int     `json:"number"`
	Title              string  `json:"title"`
	URL                string  `json:"url"`
	State              string  `json:"state"`
	IsDraft            bool    `json:"isDraft"`
	Mergeable          string  `json:"mergeable"`
	ReviewDecision     string  `json:"reviewDecision"`
	AutoMerge          bool    `json:"autoMerge"`
	MergeQueueState    string  `json:"mergeQueueState"`
	MergeQueuePosition int     `json:"mergeQueuePosition"`
	Author             string  `json:"author"`
	Labels             []Label `json:"labels"`
	UnresolvedComments int     `json:"unresolvedComments"`
	CIStatus           string  `json:"ciStatus"`
	Viewer             string  `json:"viewer"`
}

// InMergeQueue is true when this PR has an active merge queue entry, in any
// state (queued, awaiting checks, mergeable, unmergeable, locked).
func (s *State) InMergeQueue() bool {
	return s != nil && s.MergeQueueState != ""
}

// MergeQueueBroken is true when the queue entry can't proceed — either it's
// flagged unmergeable or the queue itself is locked.
func (s *State) MergeQueueBroken() bool {
	return s != nil && (s.MergeQueueState == "UNMERGEABLE" || s.MergeQueueState == "LOCKED")
}

func (s *State) Conflicting() bool {
	return s != nil && s.Mergeable == "CONFLICTING"
}
