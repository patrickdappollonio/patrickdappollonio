package main

import (
	"encoding/json"
	"fmt"
	"net/http"
)

type StarredRepos []StarredRepo

func (p StarredRepos) Take(start, limit int) StarredRepos {
	if start >= len(p) {
		return nil
	}

	end := start + limit
	if end > len(p) {
		end = len(p)
	}

	return p[start:end]
}

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

func getStarredRepos(username string, maxItems int) ([]StarredRepo, error) {
	u := fmt.Sprintf("https://api.github.com/users/%s/starred", username)

	resp, err := http.Get(u)
	if err != nil {
		return nil, fmt.Errorf("failed to get starred repos for user %q: %w", username, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to get starred repos for user %q: API returned non-200 status code: %s", username, resp.Status)
	}

	var starredRepos []StarredRepo
	if err := json.NewDecoder(resp.Body).Decode(&starredRepos); err != nil {
		return nil, fmt.Errorf("failed to decode starred repos for user %q: %w", username, err)
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
