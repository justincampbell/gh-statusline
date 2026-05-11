package pr

import (
	"encoding/json"
	"fmt"

	"github.com/cli/go-gh/v2"
)

const query = `query($owner: String!, $repo: String!, $branch: String!) {
  viewer { login }
  repository(owner: $owner, name: $repo) {
    pullRequests(headRefName: $branch, first: 2, states: [OPEN, MERGED], orderBy: {field: UPDATED_AT, direction: DESC}) {
      nodes {
        number
        title
        url
        state
        isDraft
        mergeable
        reviewDecision
        author { login }
        autoMergeRequest { enabledAt }
        labels(first: 10) { nodes { name color } }
        commits(last: 1) {
          nodes {
            commit {
              statusCheckRollup {
                contexts(first: 100) {
                  nodes {
                    __typename
                    ... on CheckRun { status conclusion }
                    ... on StatusContext { state }
                  }
                }
              }
            }
          }
        }
        reviewThreads(first: 100) { nodes { isResolved } }
      }
    }
  }
}`

// Fetch returns the PR for the given branch. Returns (nil, nil) when the
// branch has no associated PR — that is not an error condition for a statusline.
func Fetch(owner, repo, branch string) (*State, error) {
	stdout, _, err := gh.Exec(
		"api", "graphql",
		"-f", "query="+query,
		"-f", "owner="+owner,
		"-f", "repo="+repo,
		"-f", "branch="+branch,
	)
	if err != nil {
		return nil, fmt.Errorf("gh api graphql: %w", err)
	}
	return parse(stdout.Bytes())
}

type gqlResponse struct {
	Data struct {
		Viewer struct {
			Login string `json:"login"`
		} `json:"viewer"`
		Repository struct {
			PullRequests struct {
				Nodes []gqlPR `json:"nodes"`
			} `json:"pullRequests"`
		} `json:"repository"`
	} `json:"data"`
}

type gqlPR struct {
	Number         int    `json:"number"`
	Title          string `json:"title"`
	URL            string `json:"url"`
	State          string `json:"state"`
	IsDraft        bool   `json:"isDraft"`
	Mergeable      string `json:"mergeable"`
	ReviewDecision string `json:"reviewDecision"`
	Author         struct {
		Login string `json:"login"`
	} `json:"author"`
	AutoMergeRequest *struct {
		EnabledAt string `json:"enabledAt"`
	} `json:"autoMergeRequest"`
	Labels struct {
		Nodes []Label `json:"nodes"`
	} `json:"labels"`
	Commits struct {
		Nodes []struct {
			Commit struct {
				StatusCheckRollup struct {
					Contexts struct {
						Nodes []StatusCheck `json:"nodes"`
					} `json:"contexts"`
				} `json:"statusCheckRollup"`
			} `json:"commit"`
		} `json:"nodes"`
	} `json:"commits"`
	ReviewThreads struct {
		Nodes []struct {
			IsResolved bool `json:"isResolved"`
		} `json:"nodes"`
	} `json:"reviewThreads"`
}

func parse(raw []byte) (*State, error) {
	var resp gqlResponse
	if err := json.Unmarshal(raw, &resp); err != nil {
		return nil, fmt.Errorf("parse graphql response: %w", err)
	}

	nodes := resp.Data.Repository.PullRequests.Nodes
	if len(nodes) == 0 {
		s := &State{Viewer: resp.Data.Viewer.Login}
		// Empty result — Number stays 0, indicating "no PR for this branch".
		return s, nil
	}

	// Prefer OPEN over MERGED if both exist for the same branch.
	gpr := nodes[0]
	for _, n := range nodes {
		if n.State == "OPEN" {
			gpr = n
			break
		}
	}

	s := &State{
		Number:         gpr.Number,
		Title:          gpr.Title,
		URL:            gpr.URL,
		State:          gpr.State,
		IsDraft:        gpr.IsDraft,
		Mergeable:      gpr.Mergeable,
		ReviewDecision: gpr.ReviewDecision,
		AutoMerge:      gpr.AutoMergeRequest != nil,
		Author:         gpr.Author.Login,
		Labels:         gpr.Labels.Nodes,
		Viewer:         resp.Data.Viewer.Login,
	}

	if len(gpr.Commits.Nodes) > 0 {
		s.StatusCheckRollup = gpr.Commits.Nodes[0].Commit.StatusCheckRollup.Contexts.Nodes
	}

	for _, t := range gpr.ReviewThreads.Nodes {
		if !t.IsResolved {
			s.UnresolvedComments++
		}
	}

	s.CIStatus = rollupStatus(s.StatusCheckRollup)

	return s, nil
}

// rollupStatus aggregates individual check states into "passed", "failed",
// "pending", or "none".
func rollupStatus(checks []StatusCheck) string {
	if len(checks) == 0 {
		return "none"
	}
	var hasRunning, hasFailed, hasSuccess bool
	for _, c := range checks {
		if c.Typename == "StatusContext" {
			switch c.State {
			case "FAILURE", "ERROR":
				hasFailed = true
			case "PENDING":
				hasRunning = true
			case "SUCCESS":
				hasSuccess = true
			}
		} else {
			switch c.Status {
			case "IN_PROGRESS", "QUEUED":
				hasRunning = true
			}
			switch c.Conclusion {
			case "FAILURE", "ERROR", "TIMED_OUT", "ACTION_REQUIRED":
				hasFailed = true
			case "SUCCESS":
				hasSuccess = true
			}
		}
	}
	switch {
	case hasFailed:
		return "failed"
	case hasRunning:
		return "pending"
	case hasSuccess:
		return "passed"
	default:
		return "none"
	}
}
