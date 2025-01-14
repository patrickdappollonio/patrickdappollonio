package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
)

var client = &http.Client{}

func doGet[T any](ctx context.Context, v T, token string, url string) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	if token != "" {
		req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", token))
	}

	req.Header.Set("Accept", "application/json")
	req.Header.Set("X-Github-Api-Version", "2022-11-28")

	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to make request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		var response bytes.Buffer
		response.ReadFrom(resp.Body)

		if response.Len() > 0 {
			return fmt.Errorf("failed to get response: %s: %s", resp.Status, response.String())
		}

		return fmt.Errorf("failed to get response: %s", resp.Status)
	}

	if err := json.NewDecoder(resp.Body).Decode(v); err != nil {
		return fmt.Errorf("failed to decode response: %w", err)
	}

	return nil
}
