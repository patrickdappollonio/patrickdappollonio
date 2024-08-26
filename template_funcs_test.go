package main

import "testing"

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
