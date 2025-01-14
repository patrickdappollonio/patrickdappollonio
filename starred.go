package main

import (
	"context"
	"fmt"
)

type StarredRepo struct {
	Name    string `json:"full_name"`
	Private bool   `json:"private"`
	URL     string `json:"html_url"`
	Stars   int    `json:"stargazers_count"`
	Owner   struct {
		User string `json:"login"`
	} `json:"owner"`
}

func (s *StarredRepo) IsPrivate() bool {
	return s.Private
}

func (s *StarredRepo) IsOwned(username string) bool {
	return s.Owner.User == username
}

func (s *StarredRepo) Empty() bool {
	return s.Name == ""
}

func getStarredRepos(ctx context.Context, token, username string, maxItems int) ([]StarredRepo, error) {
	u := fmt.Sprintf("https://api.github.com/users/%s/starred", username)

	var starredRepos []StarredRepo
	if err := doGet(ctx, &starredRepos, token, u); err != nil {
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
