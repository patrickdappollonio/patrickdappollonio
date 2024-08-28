package main

import (
	"encoding/xml"
	"fmt"
	"net/http"
	"time"
)

type Articles []Article

func (p Articles) Take(start, limit int) Articles {
	if start >= len(p) {
		return nil
	}

	end := start + limit
	if end > len(p) {
		end = len(p)
	}

	return p[start:end]
}

type Article struct {
	Title string `xml:"title"`
	Link  string `xml:"link"`
	Date  string `xml:"pubDate"`
}

func (a *Article) GoDate() (time.Time, error) {
	return time.Parse("Mon, 02 Jan 2006 15:04:05 -0700", a.Date)
}

func getArticles(location string, maxItems int) ([]Article, error) {
	res, err := http.Get(location)
	if err != nil {
		return nil, fmt.Errorf("failed to get feed at %q: %w", location, err)
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to get feed at %q: API returned non-200 status code: %s", location, res.Status)
	}

	var feed struct {
		Items []Article `xml:"channel>item"`
	}
	if err := xml.NewDecoder(res.Body).Decode(&feed); err != nil {
		return nil, fmt.Errorf("failed to decode feed at %q: %w", location, err)
	}

	cleaned := make([]Article, 0, len(feed.Items))
	for _, item := range feed.Items {
		if len(cleaned) >= maxItems {
			break
		}

		if item.Title == "" || item.Link == "" || item.Date == "" {
			continue
		}

		if d, err := item.GoDate(); err != nil || d.IsZero() {
			continue
		}

		cleaned = append(cleaned, item)
	}

	return cleaned, nil
}
