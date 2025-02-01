package main

import (
	"testing"
)

func Test_humanizeBigNumber(t *testing.T) {
	tests := []struct {
		name string
		n    int
		want string
	}{
		{"1", 1, "1"},
		{"1000", 1000, "1K"},
		{"4049", 4049, "4K"},
		{"4050", 4050, "4.1K"},
		{"4099", 4099, "4.1K"},
		{"12600", 12600, "12.6K"},
		{"3500000", 3500000, "3.5M"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := humanizeBigNumber(tt.n); got != tt.want {
				t.Errorf("humanizeBigNumber() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_contributedOrgsMarkdown(t *testing.T) {
	tests := []struct {
		name string
		orgs []string
		want string
	}{
		{
			name: "single",
			orgs: []string{"org1"},
			want: "[@org1](https://github.com/org1)",
		},
		{
			name: "double",
			orgs: []string{"org1", "org2"},
			want: "[@org1](https://github.com/org1) and [@org2](https://github.com/org2)",
		},
		{
			name: "triple",
			orgs: []string{"org1", "org2", "org3"},
			want: "[@org1](https://github.com/org1), [@org2](https://github.com/org2) and [@org3](https://github.com/org3)",
		},
		{
			name: "multiple",
			orgs: []string{"org1", "org2", "org3", "org4"},
			want: "[@org1](https://github.com/org1), [@org2](https://github.com/org2), [@org3](https://github.com/org3) and [@org4](https://github.com/org4)",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := contributedOrgsMarkdown(tt.orgs); string(got) != tt.want {
				t.Errorf("contributedOrgsMarkdown() = %q, want %q", got, tt.want)
			}
		})
	}
}
