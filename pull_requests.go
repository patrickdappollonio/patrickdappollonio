package main

import (
	"embed"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"html/template"
	"net/http"
	"net/url"
	"regexp"
	"strings"
	"time"
)

type PullRequestResponse struct {
	TotalCount int           `json:"total_count"`
	Items      []PullRequest `json:"items"`
}

type PullRequest struct {
	URL              string    `json:"html_url"`
	RepositoryAPIURL string    `json:"repository_url"`
	ID               int64     `json:"number"`
	Title            string    `json:"title"`
	State            string    `json:"state"`
	Locked           bool      `json:"locked"`
	Comments         int       `json:"comments"`
	CreatedAt        time.Time `json:"created_at"`
	UpdatedAt        time.Time `json:"updated_at"`
	ClosedAt         time.Time `json:"closed_at"`
	Draft            bool      `json:"draft"`
	Body             string    `json:"body"`
	PullRequest      struct {
		MergedAt time.Time `json:"merged_at"`
	} `json:"pull_request"`
}

//go:embed images/statuses/*.png
var statusImages embed.FS

const imageTemplate = `<img src="data:image/png;base64,B64" width="SIZE" height="SIZE" alt="ALT"> STATUS`

func (p *PullRequest) StatusImageHTML(sizePixels int) template.HTML {
	status := "open"
	if p.Closed() {
		status = "closed"
	} else if p.Merged() {
		status = "merged"
	} else if p.Draft {
		status = "draft"
	}

	if sizePixels > 128 {
		sizePixels = 128
	}

	b, err := statusImages.ReadFile(fmt.Sprintf("images/statuses/github-%s.png", status))
	if err != nil {
		return ""
	}

	return template.HTML(
		strings.NewReplacer(
			"B64", base64.StdEncoding.EncodeToString(b),
			"SIZE", fmt.Sprintf("%d", sizePixels),
			"ALT", status,
			"STATUS", status,
		).Replace(imageTemplate),
	)
}

func (p *PullRequest) ProjectOrg() string {
	return strings.TrimPrefix(p.RepositoryAPIURL, "https://api.github.com/repos/")
}

func (p *PullRequest) RepositoryName() string {
	pieces := strings.Split(p.ProjectOrg(), "/")
	if len(pieces) < 2 {
		return ""
	}

	return pieces[1]
}

func (p *PullRequest) RepositoryURL() string {
	return strings.ReplaceAll(p.RepositoryAPIURL, "https://api.github.com/repos", "https://github.com")
}

func (p *PullRequest) Merged() bool {
	return !p.PullRequest.MergedAt.IsZero()
}

func (p *PullRequest) Closed() bool {
	return !p.ClosedAt.IsZero() && !p.Merged()
}

var reExtractProject = regexp.MustCompile(`^https:\/\/api\.github\.com\/repos\/([^\/]+)\/.+$`)

func (p *PullRequest) ContributedToOrg() string {
	res := reExtractProject.FindStringSubmatch(p.RepositoryAPIURL)
	if len(res) != 2 {
		return ""
	}

	return res[1]
}

func getPullRequests(username string, maxItems, maxOrgs int) ([]PullRequest, []string, error) {
	u, _ := url.Parse("https://api.github.com/search/issues")

	vals := url.Values{}
	vals.Set("q", fmt.Sprintf("author:%s type:pr", username))
	vals.Set("per_page", "100")
	u.RawQuery = vals.Encode()

	resp, err := http.Get(u.String())
	if err != nil {
		return nil, nil, fmt.Errorf("failed to get pull requests for user %q: %w", username, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, nil, fmt.Errorf("failed to get pull requests for user %q: API returned non-200 status code: %s", username, resp.Status)
	}

	var prs PullRequestResponse
	if err := json.NewDecoder(resp.Body).Decode(&prs); err != nil {
		return nil, nil, fmt.Errorf("failed to decode pull requests for user %q: %w", username, err)
	}

	limited := make([]PullRequest, 0, len(prs.Items))
	contributionRepos := make(map[string]struct{})

	for _, pr := range prs.Items {
		if len(limited) < maxItems {
			limited = append(limited, pr)
		}

		// Only count contributions to other orgs if the PR is merged or still open
		if org := pr.ContributedToOrg(); org != username && org != "" && (pr.Merged() || !pr.Closed()) {
			contributionRepos[org] = struct{}{}
		}
	}

	var repos []string
	for repo := range contributionRepos {
		if len(repos) >= maxOrgs {
			break
		}

		repos = append(repos, repo)
	}

	return limited, repos, nil
}
