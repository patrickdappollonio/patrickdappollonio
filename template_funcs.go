package main

import (
	"strconv"
	"strings"
	"text/template"
	"time"
)

var fncs = template.FuncMap{
	"formatDate": func(date time.Time) string {
		return date.Format("January 02, 2006")
	},
	"formatNumber": formatNumber,
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
