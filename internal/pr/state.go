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
	Author             string  `json:"author"`
	Labels             []Label `json:"labels"`
	UnresolvedComments int     `json:"unresolvedComments"`
	CIStatus           string  `json:"ciStatus"`
	Viewer             string  `json:"viewer"`
}

func (s *State) Conflicting() bool {
	return s != nil && s.Mergeable == "CONFLICTING"
}
