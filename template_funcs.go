package main

import (
	"fmt"
	"math"
	"strconv"
	"strings"
	"text/template"
	"time"
)

var fncs = template.FuncMap{
	"formatDate": func(date time.Time) string {
		return date.Format("January 02, 2006")
	},
	"formatDateTime": func(date time.Time) string {
		return date.Format("January 02, 2006 at 15:04:05 MST")
	},
	"formatNumber": formatNumber,
	"now": func() time.Time {
		return time.Now()
	},
	// humanizeBigNumber takes a number over 1000 and returns
	// its short, human-readable form. For example, 1000 becomes 1K.
	"humanizeBigNumber": humanizeBigNumber,
	"sub": func(a, b int) int {
		return a - b
	},
	"ellipsize": func(n int, s string) string {
		// cut the string on `n` where `n` is the number of words
		// to keep in the string
		words := strings.Fields(s)
		if len(words) <= n {
			return s
		}

		return strings.Join(words[:n], " ") + "..."
	},
}

func humanizeBigNumber(n int) string {
	if n < 1000 {
		return strconv.Itoa(n)
	}

	if n < 1000000 {
		remainder := n % 1000
		if remainder == 0 {
			return fmt.Sprintf("%dK", n/1000)
		}
		decimal := int(math.Round(float64(remainder) / 100))
		if decimal == 0 {
			return fmt.Sprintf("%dK", n/1000)
		}
		return fmt.Sprintf("%d.%dK", n/1000, decimal)
	}

	remainder := n % 1000000
	if remainder == 0 {
		return fmt.Sprintf("%dM", n/1000000)
	}
	decimal := int(math.Round(float64(remainder) / 100000))
	if decimal == 0 {
		return fmt.Sprintf("%dM", n/1000000)
	}
	return fmt.Sprintf("%d.%dM", n/1000000, decimal)
}

func formatNumber(n int) string {
	in := strconv.Itoa(n)
	var out strings.Builder
	digitCount := 0

	for i := len(in) - 1; i >= 0; i-- {
		if digitCount > 0 && digitCount%3 == 0 {
			out.WriteString(",")
		}
		out.WriteByte(in[i])
		digitCount++
	}

	// Reverse the string to get the correct order
	return reverse(out.String())
}

func reverse(s string) string {
	r := []rune(s)
	for i, j := 0, len(r)-1; i < j; i, j = i+1, j-1 {
		r[i], r[j] = r[j], r[i]
	}

	return string(r)
}
