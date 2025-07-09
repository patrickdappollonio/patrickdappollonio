package main

import (
	"fmt"
	"html/template"
	"math"
	"reflect"
	"strconv"
	"strings"
	"time"
)

// TemplateFunctions returns a template.FuncMap with all custom template functions
func TemplateFunctions() template.FuncMap {
	return template.FuncMap{
		// Date and time functions
		"formatDate":     formatDate,
		"formatDateTime": formatDateTime,
		"now":            now,

		// Number formatting functions
		"formatNumber":      formatNumber,
		"humanizeBigNumber": humanizeBigNumber,

		// String manipulation functions
		"ellipsize":      ellipsize,
		"ellipsizechars": ellipsizechars,

		// Array/slice manipulation functions
		"take": take,
		"skip": skip,
		"seq":  seq,
		"sseq": sseq,

		// Math functions
		"add":     add,
		"sub":     sub,
		"div":     div,
		"mod":     mod,
		"divCeil": divCeil,
		"lt":      lt,
		"gt":      gt,

		// HTML rendering functions
		"dualimage": renderImageUnclickable,

		// GitHub-specific functions
		"contributedOrgsMarkdown": contributedOrgsMarkdown,
	}
}

// Date and time functions
func formatDate(date time.Time) string {
	return date.Format("January 02, 2006")
}

func formatDateTime(date time.Time) string {
	return date.Format("January 02, 2006 at 15:04:05 MST")
}

func now() time.Time {
	return time.Now()
}

// String manipulation functions
func ellipsize(n int, s string) string {
	// Cut the string on `n` where `n` is the number of words to keep
	words := strings.Fields(s)
	if len(words) <= n {
		return s
	}
	return strings.Join(words[:n], " ") + "..."
}

func ellipsizechars(n int, s string) string {
	// Cut the string on `n` where `n` is the number of characters to keep
	if len(s) <= n {
		return s
	}
	return s[:n] + "..."
}

// Array/slice manipulation functions
func take(n int, slice interface{}) interface{} {
	v := reflect.ValueOf(slice)
	if v.Kind() != reflect.Slice {
		return slice // Return as-is if not a slice
	}
	if n > v.Len() {
		n = v.Len() // Adjust n if it exceeds slice length
	}
	return v.Slice(0, n).Interface()
}

func skip(n int, slice interface{}) interface{} {
	v := reflect.ValueOf(slice)
	if v.Kind() != reflect.Slice {
		return slice // Return as-is if not a slice
	}
	if n > v.Len() {
		n = v.Len() // Adjust n if it exceeds slice length
	}
	return v.Slice(n, v.Len()).Interface()
}

func seq(n int) []int {
	s := make([]int, n)
	for i := 0; i < n; i++ {
		s[i] = i
	}
	return s
}

func sseq(start, end int) []int {
	a := make([]int, end-start+1)
	for i := range a {
		a[i] = start + i
	}
	return a
}

// Math functions
func add(a, b int) int { return a + b }
func sub(a, b int) int { return a - b }
func div(a, b int) int { return a / b }
func mod(a, b int) int { return a % b }
func lt(a, b int) bool { return a < b }
func gt(a, b int) bool { return a > b }

func divCeil(a, b int) int {
	return (a + b - 1) / b
}

// Number formatting functions
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

	return reverse(out.String())
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

// HTML rendering functions
const dualImageTemplate = `<picture><source media="(prefers-color-scheme: dark)" srcset="{{IMAGE_DARK}}"><source media="(prefers-color-scheme: light)" srcset="{{IMAGE_LIGHT}}"><img src="{{IMAGE_LIGHT}}" alt="{{ALT_TEXT}}"></picture>`

func renderImageUnclickable(images ...string) template.HTML {
	darkImage, lightImage, alt := "", "", ""

	switch len(images) {
	case 1:
		darkImage = images[0]
		lightImage = images[0]
	case 2:
		darkImage = images[0]
		lightImage = images[1]
	case 3:
		darkImage = images[0]
		lightImage = images[1]
		alt = images[2]
	default:
		return ""
	}

	return template.HTML(strings.NewReplacer(
		"{{IMAGE_LIGHT}}", lightImage,
		"{{IMAGE_DARK}}", darkImage,
		"{{ALT_TEXT}}", alt,
	).Replace(dualImageTemplate))
}

// GitHub-specific functions
func contributedOrgsMarkdown(orgs []string) template.HTML {
	var sb strings.Builder

	switch len(orgs) {
	case 0:
		return ""
	case 1:
		fmt.Fprintf(&sb, "[@%s](https://github.com/%s)", orgs[0], orgs[0])
	default:
		for i, org := range orgs {
			if i == len(orgs)-1 {
				fmt.Fprintf(&sb, " and [@%s](https://github.com/%s)", org, org)
			} else if i == 0 {
				fmt.Fprintf(&sb, "[@%s](https://github.com/%s)", org, org)
			} else {
				fmt.Fprintf(&sb, ", [@%s](https://github.com/%s)", org, org)
			}
		}
	}

	return template.HTML(sb.String())
}

// Helper functions
func reverse(s string) string {
	r := []rune(s)
	for i, j := 0, len(r)-1; i < j; i, j = i+1, j-1 {
		r[i], r[j] = r[j], r[i]
	}
	return string(r)
}
