package main

import (
	"context"
	"fmt"
	"html/template"
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
		APIURL   string    `json:"url"`
	} `json:"pull_request"`

	Commits           int `json:"commits"`
	Additions         int `json:"additions"`
	Deletions         int `json:"deletions"`
	ChangedFiles      int `json:"changed_files"`
	updatedCommitInfo bool
}

const imageTemplate = `<picture><source media="(prefers-color-scheme: dark)" srcset="LOCATION" width="SIZE" height="SIZE"><source media="(prefers-color-scheme: light)" srcset="LOCATION" width="SIZE" height="SIZE"><img src="LOCATION" width="SIZE" height="SIZE" alt="STATUS"></picture> STATUS`

var statuses = map[string]string{
	"closed": `https://raw.githubusercontent.com/patrickdappollonio/patrickdappollonio/refs/heads/main/images/statuses/github-closed.png`,
	"merged": `https://raw.githubusercontent.com/patrickdappollonio/patrickdappollonio/refs/heads/main/images/statuses/github-merged.png`,
	"open":   `https://raw.githubusercontent.com/patrickdappollonio/patrickdappollonio/refs/heads/main/images/statuses/github-open.png`,
	"draft":  `https://raw.githubusercontent.com/patrickdappollonio/patrickdappollonio/refs/heads/main/images/statuses/github-draft.png`,
}

// StatusImageHTML returns an HTML image tag with the status of the pull request.
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

	b, ok := statuses[status]
	if !ok {
		return ""
	}

	return template.HTML(
		strings.NewReplacer(
			"LOCATION", b,
			"SIZE", fmt.Sprintf("%d", sizePixels),
			"STATUS", status,
		).Replace(imageTemplate),
	)
}

// ProjectOrg returns the organization of the project where the pull request was made.
func (p *PullRequest) ProjectOrg() string {
	return strings.TrimPrefix(p.RepositoryAPIURL, "https://api.github.com/repos/")
}

// RepositoryName returns the name of the repository where the pull request was made.
func (p *PullRequest) RepositoryName() string {
	pieces := strings.Split(p.ProjectOrg(), "/")
	if len(pieces) < 2 {
		return ""
	}

	return pieces[1]
}

// RepositoryURL returns the URL of the repository where the pull request was made.
func (p *PullRequest) RepositoryURL() string {
	return strings.ReplaceAll(p.RepositoryAPIURL, "https://api.github.com/repos", "https://github.com")
}

// Merged returns true if the pull request was merged.
func (p *PullRequest) Merged() bool {
	return !p.PullRequest.MergedAt.IsZero()
}

// Closed returns true if the pull request was closed.
func (p *PullRequest) Closed() bool {
	return !p.ClosedAt.IsZero() && !p.Merged()
}

// GetPRMetrics returns the additions and deletions of the pull request in a HTML format.
func (p *PullRequest) GetPRMetrics() (template.HTML, error) {
	if !p.updatedCommitInfo {
		return "", fmt.Errorf("PR information not updated")
	}

	imageURL := fmt.Sprintf("https://diff-counter.patrickdap.dev/?add=%d&del=%d&height=17", p.Additions, p.Deletions)

	return template.HTML(fmt.Sprintf(
		`<picture><source media="(prefers-color-scheme: dark)" srcset="%s"><source media="(prefers-color-scheme: light)" srcset="%s"><img src="%s" alt="+%s -%s"></picture>`,
		imageURL, imageURL, imageURL, formatNumber(p.Additions), formatNumber(p.Deletions),
	)), nil
}

var reExtractProject = regexp.MustCompile(`^https:\/\/api\.github\.com\/repos\/([^\/]+)\/.+$`)

// ContributedToOrg returns the organization of the project where the pull request was made.
func (p *PullRequest) ContributedToOrg() string {
	res := reExtractProject.FindStringSubmatch(p.RepositoryAPIURL)
	if len(res) != 2 {
		return ""
	}

	return res[1]
}

func getPullRequests(ctx context.Context, token, username string, maxItems, maxOrgs int) ([]PullRequest, []string, error) {
	u, _ := url.Parse("https://api.github.com/search/issues")

	vals := url.Values{}
	vals.Set("q", fmt.Sprintf("author:%s type:pr", username))
	vals.Set("per_page", "100")
	u.RawQuery = vals.Encode()

	var prs PullRequestResponse
	if err := doGet(ctx, &prs, token, u.String()); err != nil {
		return nil, nil, fmt.Errorf("failed to get pull requests for user %q: %w", username, err)
	}

	limited := make([]PullRequest, 0, len(prs.Items))
	contributionRepos := make(map[string]struct{})

	for _, pr := range prs.Items {
		if len(limited) < maxItems {
			if err := updatePRInformation(ctx, token, &pr); err != nil {
				return nil, nil, fmt.Errorf("failed to update PR information for PR %s: %w", pr.URL, err)
			}

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

func updatePRInformation(ctx context.Context, token string, pr *PullRequest) error {
	var updated PullRequest
	if err := doGet(ctx, &updated, token, pr.PullRequest.APIURL); err != nil {
		return fmt.Errorf("failed to get PR information for PR %s: %w", pr.PullRequest.APIURL, err)
	}

	pr.Additions = updated.Additions
	pr.ChangedFiles = updated.ChangedFiles
	pr.Commits = updated.Commits
	pr.Deletions = updated.Deletions
	pr.updatedCommitInfo = true
	return nil
}
