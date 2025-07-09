package main

import (
	"context"
	"fmt"
)

// StarredRepo represents a GitHub starred repository
type StarredRepo struct {
	Name    string `json:"full_name"`
	Private bool   `json:"private"`
	URL     string `json:"html_url"`
	Stars   int    `json:"stargazers_count"`
	Owner   struct {
		User string `json:"login"`
	} `json:"owner"`
}

// IsValid checks if the starred repository has valid data
func (s *StarredRepo) IsValid() bool {
	return s.Name != "" && s.URL != "" && s.Owner.User != ""
}

// Validate validates the starred repository data
func (s *StarredRepo) Validate() error {
	if s.Name == "" {
		return fmt.Errorf("repository name cannot be empty")
	}

	if s.URL == "" {
		return fmt.Errorf("repository URL cannot be empty")
	}

	if s.Owner.User == "" {
		return fmt.Errorf("repository owner cannot be empty")
	}

	if s.Stars < 0 {
		return fmt.Errorf("star count cannot be negative")
	}

	return nil
}

// IsPrivate returns true if the repository is private
func (s *StarredRepo) IsPrivate() bool {
	return s.Private
}

// IsOwned returns true if the repository is owned by the specified user
func (s *StarredRepo) IsOwned(username string) bool {
	return s.Owner.User == username
}

// Empty returns true if the repository data is empty
func (s *StarredRepo) Empty() bool {
	return s.Name == ""
}

func getStarredRepos(ctx context.Context, client *GitHubAPIClient, username string, maxItems int) ([]StarredRepo, error) {
	u := fmt.Sprintf("https://api.github.com/users/%s/starred", username)

	var starredRepos []StarredRepo
	if err := client.Get(ctx, &starredRepos, u); err != nil {
		return nil, fmt.Errorf("failed to get starred repos for user %q: %w", username, err)
	}

	filtered := make([]StarredRepo, 0, len(starredRepos))
	for _, repo := range starredRepos {
		if len(filtered) >= maxItems {
			break
		}

		if !repo.IsPrivate() && !repo.IsOwned(username) {
			filtered = append(filtered, repo)
		}
	}

	return filtered, nil
}
