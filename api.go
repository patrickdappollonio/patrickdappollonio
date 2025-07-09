package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
)

// GitHubAPIClient handles GitHub API requests
type GitHubAPIClient struct {
	httpClient *http.Client
	token      string
}

// NewGitHubAPIClient creates a new GitHub API client
func NewGitHubAPIClient(token string) *GitHubAPIClient {
	return &GitHubAPIClient{
		httpClient: &http.Client{},
		token:      token,
	}
}

// Get makes a GET request to the GitHub API
func (c *GitHubAPIClient) Get(ctx context.Context, v interface{}, url string) error {
	// Validate inputs
	if v == nil {
		return fmt.Errorf("response destination cannot be nil")
	}

	if url == "" {
		return fmt.Errorf("URL cannot be empty")
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	if c.token != "" {
		req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", c.token))
	}

	req.Header.Set("Accept", "application/json")
	req.Header.Set("X-Github-Api-Version", "2022-11-28")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to make request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		var response bytes.Buffer
		response.ReadFrom(resp.Body)

		if response.Len() > 0 {
			return fmt.Errorf("%w: HTTP %s - %s", ErrAPIRequest, resp.Status, response.String())
		}

		return fmt.Errorf("%w: HTTP %s", ErrAPIRequest, resp.Status)
	}

	// Validate content type
	contentType := resp.Header.Get("Content-Type")
	if contentType != "" && !strings.Contains(contentType, "application/json") {
		return fmt.Errorf("unexpected content type: %s", contentType)
	}

	if err := json.NewDecoder(resp.Body).Decode(v); err != nil {
		return fmt.Errorf("failed to decode response: %w", err)
	}

	return nil
}

// doGet is a backward compatibility function for existing code
func doGet[T any](ctx context.Context, v T, token string, url string) error {
	client := NewGitHubAPIClient(token)
	return client.Get(ctx, v, url)
}
