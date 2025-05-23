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
	"ellipsize": func(n int, s string) string {
		// cut the string on `n` where `n` is the number of words
		// to keep in the string
		words := strings.Fields(s)
		if len(words) <= n {
			return s
		}

		return strings.Join(words[:n], " ") + "..."
	},
	"ellipsizechars": func(n int, s string) string {
		// cut the string on `n` where `n` is the number of characters
		// to keep in the string
		if len(s) <= n {
			return s
		}

		return s[:n] + "..."
	},

	"take": take,
	"skip": skip,

	"seq": func(n int) []int {
		s := make([]int, n)
		for i := 0; i < n; i++ {
			s[i] = i
		}
		return s
	},

	"add": func(a, b int) int { return a + b },
	"sub": func(a, b int) int { return a - b },
	"div": func(a, b int) int { return a / b },
	"lt":  func(a, b int) bool { return a < b },
	"gt":  func(a, b int) bool { return a > b },
	"mod": func(a, b int) int { return a % b },
	"sseq": func(start, end int) []int {
		a := make([]int, end-start+1)
		for i := range a {
			a[i] = start + i
		}
		return a
	},
	"divCeil": func(a, b int) int {
		return (a + b - 1) / b
	},

	"dualimage": renderImageUnclickable,

	"contributedOrgsMarkdown": contributedOrgsMarkdown,
}

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
