package pr

import (
	"context"
	"fmt"

	"github.com/cli/go-gh/v2/pkg/api"
)

// Result bundles the PR for the current branch (zero PR when none exists) and
// the repo/CI info used as a fallback when there is no PR to show.
type Result struct {
	PR     *State
	Branch *BranchStatus
}

const query = `query($owner: String!, $repo: String!, $branch: String!) {
  viewer { login }
  repository(owner: $owner, name: $repo) {
    url
    ref(qualifiedName: $branch) {
      target {
        ... on Commit {
          statusCheckRollup { state }
        }
      }
    }
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
        mergeQueueEntry { position state }
        labels(first: 10) { nodes { name color } }
        commits(last: 1) {
          nodes {
            commit {
              statusCheckRollup { state }
            }
          }
        }
        reviewThreads(first: 100) { nodes { isResolved } }
      }
    }
  }
}`

// Fetch issues a single GraphQL request that returns both the PR for the
// branch (if any) and the repo's URL plus the branch's CI rollup. ctx applies
// to the HTTP call; pass a deadline to avoid blocking past the caller's
// tolerance.
func Fetch(ctx context.Context, owner, repo, branch string) (*Result, error) {
	client, err := api.DefaultGraphQLClient()
	if err != nil {
		return nil, fmt.Errorf("graphql client: %w", err)
	}
	vars := map[string]interface{}{
		"owner":  owner,
		"repo":   repo,
		"branch": branch,
	}
	var data gqlData
	if err := client.DoWithContext(ctx, query, vars, &data); err != nil {
		return nil, fmt.Errorf("graphql query: %w", err)
	}
	return build(owner, repo, &data), nil
}

type gqlData struct {
	Viewer struct {
		Login string `json:"login"`
	} `json:"viewer"`
	Repository struct {
		URL string `json:"url"`
		Ref *struct {
			Target struct {
				StatusCheckRollup *struct {
					State string `json:"state"`
				} `json:"statusCheckRollup"`
			} `json:"target"`
		} `json:"ref"`
		PullRequests struct {
			Nodes []gqlPR `json:"nodes"`
		} `json:"pullRequests"`
	} `json:"repository"`
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
	MergeQueueEntry *struct {
		Position int    `json:"position"`
		State    string `json:"state"`
	} `json:"mergeQueueEntry"`
	Labels struct {
		Nodes []Label `json:"nodes"`
	} `json:"labels"`
	Commits struct {
		Nodes []struct {
			Commit struct {
				StatusCheckRollup *struct {
					State string `json:"state"`
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

func build(owner, repo string, data *gqlData) *Result {
	return &Result{
		PR:     buildPR(data),
		Branch: buildBranch(owner, repo, data),
	}
}

func buildPR(data *gqlData) *State {
	nodes := data.Repository.PullRequests.Nodes
	if len(nodes) == 0 {
		return &State{Viewer: data.Viewer.Login}
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
		Viewer:         data.Viewer.Login,
	}

	if gpr.MergeQueueEntry != nil {
		s.MergeQueueState = gpr.MergeQueueEntry.State
		s.MergeQueuePosition = gpr.MergeQueueEntry.Position
	}

	var rollupState string
	if len(gpr.Commits.Nodes) > 0 && gpr.Commits.Nodes[0].Commit.StatusCheckRollup != nil {
		rollupState = gpr.Commits.Nodes[0].Commit.StatusCheckRollup.State
	}
	s.CIStatus = mapRollupState(rollupState)

	for _, t := range gpr.ReviewThreads.Nodes {
		if !t.IsResolved {
			s.UnresolvedComments++
		}
	}

	return s
}

func buildBranch(owner, repo string, data *gqlData) *BranchStatus {
	bs := &BranchStatus{
		Owner: owner,
		Repo:  repo,
		URL:   data.Repository.URL,
	}
	if bs.URL == "" {
		bs.URL = fmt.Sprintf("https://github.com/%s/%s", owner, repo)
	}
	var state string
	if data.Repository.Ref != nil && data.Repository.Ref.Target.StatusCheckRollup != nil {
		state = data.Repository.Ref.Target.StatusCheckRollup.State
	}
	bs.CIStatus = mapRollupState(state)
	return bs
}

// mapRollupState normalizes GitHub's StatusState enum into the four values
// the rest of the codebase uses.
func mapRollupState(state string) string {
	switch state {
	case "SUCCESS":
		return "passed"
	case "FAILURE", "ERROR":
		return "failed"
	case "PENDING", "EXPECTED":
		return "pending"
	default:
		return "none"
	}
}
