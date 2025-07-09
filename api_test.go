// File: api_test.go
package main

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"
)

type mockResponse struct {
	Message string `json:"message"`
}

func TestDoGet(t *testing.T) {
	tests := []struct {
		name           string
		token          string
		url            string
		responseStatus int
		responseBody   interface{}
		wantError      bool
	}{
		{
			name:           "Successful GET request",
			token:          "valid-token",
			url:            "/success",
			responseStatus: http.StatusOK,
			responseBody:   mockResponse{Message: "success"},
		},
		{
			name:           "Failed GET request due to HTTP error",
			token:          "valid-token",
			url:            "/error",
			responseStatus: http.StatusInternalServerError,
			responseBody:   mockResponse{Message: "error"},
			wantError:      true,
		},
		{
			name:           "Failed GET request due to JSON decoding error",
			token:          "valid-token",
			url:            "/invalid-json",
			responseStatus: http.StatusOK,
			responseBody:   "invalid-json",
			wantError:      true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a mock server
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if r.URL.Path != tt.url {
					t.Fatalf("wanted URL to be %q, got %q", tt.url, r.URL.Path)
				}

				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(tt.responseStatus)
				if err := json.NewEncoder(w).Encode(tt.responseBody); err != nil {
					http.Error(w, err.Error(), http.StatusInternalServerError)
				}
			}))
			defer server.Close()

			// Prepare the context and response variable
			ctx := context.Background()

			// Call the DoGet function
			var response mockResponse
			err := doGet(ctx, &response, tt.token, server.URL+tt.url)

			// Check for expected error
			if (err != nil) != tt.wantError {
				t.Fatalf("DoGet() error = %v, wantErr %v", err, tt.wantError)
			}

			// Check for expected response
			if !tt.wantError && !reflect.DeepEqual(response, tt.responseBody) {
				t.Fatalf("DoGet() response = %#v, want %#v", response, tt.responseBody)
			}
		})
	}
}
