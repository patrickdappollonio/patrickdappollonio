package main

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"text/template"
	"time"

	"github.com/Masterminds/sprig/v3"
	"golang.org/x/sync/errgroup"
	"gopkg.in/yaml.v3"
)

func main() {
	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, "Error:", err)
		os.Exit(1)
	}
}

var (
	username         = envdefault("GITHUB_USERNAME", "patrickdappollonio")
	rssfeed          = envdefault("RSS_FEED", "https://www.patrickdap.com/index.xml")
	templateFile     = envdefault("TEMPLATE_FILE", "template.md.gotmpl")
	maxPRs           = envintdefault("MAX_PULL_REQUESTS", 10)
	maxOrgs          = envintdefault("MAX_CONTRIBUTED_ORGS", 5)
	maxStarred       = envintdefault("MAX_STARRED_REPOS", 20)
	maxArticles      = envintdefault("MAX_ARTICLES", 5)
	disableRSS       = envbooldefault("DISABLE_RSS", false)
	disablePRs       = envbooldefault("DISABLE_PULL_REQUESTS", false)
	disableStars     = envbooldefault("DISABLE_STARRED_REPOS", false)
	disableDataFiles = envbooldefault("DISABLE_DATA_FILES", false)
	githubToken      = os.Getenv("GITHUB_TOKEN")
)

func run() error {
	// Get the start time
	start := time.Now()

	// Read the template file
	tplFile, err := os.ReadFile(templateFile)
	if err != nil {
		if os.IsNotExist(err) {
			return fmt.Errorf("template file %q does not exist", templateFile)
		}

		return fmt.Errorf("failed to read template file %q: %w", templateFile, err)
	}

	// Data files are YAML files in the current working directory (excluding
	// subdirectories). The data is loaded into the template as a map of string
	// to any. Data values are available under ".Data" and the file name without
	// extension is used as the key.
	additionalData := make(map[string]any)
	if !disableDataFiles {
		dir, err := os.ReadDir(".")
		if err != nil {
			return fmt.Errorf("failed to read directory: %w", err)
		}

		for _, entry := range dir {
			name := strings.TrimSuffix(entry.Name(), filepath.Ext(entry.Name()))
			fname := entry.Name()
			ext := filepath.Ext(entry.Name())

			if (ext == ".yaml" || ext == ".yml") && !strings.HasPrefix(name, ".") {
				var contents any
				dataFile, err := os.ReadFile(fname)
				if err != nil {
					if os.IsNotExist(err) {
						return fmt.Errorf("data file %q does not exist", fname)
					}

					return fmt.Errorf("failed to read data file %q: %w", fname, err)
				}

				if err := yaml.Unmarshal(dataFile, &contents); err != nil {
					return fmt.Errorf("failed to decode YAML data file %q: %w", fname, err)
				}

				if contents != nil {
					additionalData[name] = contents
				}
			}
		}
	}

	funcs := sprig.FuncMap()
	for k, v := range fncs {
		funcs[k] = v
	}

	tmpl, err := template.New("template").Funcs(fncs).Parse(string(tplFile))
	if err != nil {
		return fmt.Errorf("failed to parse template: %w", err)
	}

	var (
		eg           = errgroup.Group{}
		prs          []PullRequest
		contributed  []string
		starredRepos []StarredRepo
		articles     []Article
	)

	// Fetch pull requests
	eg.Go(func() error {
		if disablePRs {
			return nil
		}

		var err error
		prs, contributed, err = getPullRequests(context.Background(), githubToken, username, maxPRs, maxOrgs)
		if err != nil {
			return fmt.Errorf("failed to get pull requests: %w", err)
		}
		return nil
	})

	// Fetch starred repos
	eg.Go(func() error {
		if disableStars {
			return nil
		}

		var err error
		starredRepos, err = getStarredRepos(context.Background(), githubToken, username, maxStarred)
		if err != nil {
			return fmt.Errorf("failed to get starred repos: %w", err)
		}
		return nil
	})

	// Fetch RSS feed
	eg.Go(func() error {
		if disableRSS {
			return nil
		}

		var err error
		articles, err = getArticles(rssfeed, maxArticles)
		if err != nil {
			return fmt.Errorf("failed to read feed %q %w", rssfeed, err)
		}
		return nil
	})

	// Wait for all goroutines to finish
	if err := eg.Wait(); err != nil {
		return err
	}

	// Hold the contents of the template in a buffer
	var buf bytes.Buffer

	// Execute the template
	if err := tmpl.Execute(&buf, struct {
		GitHubUsername  string
		PullRequests    []PullRequest
		ContributedOrgs []string
		StarredRepos    []StarredRepo
		Articles        []Article
		Data            map[string]any
	}{
		GitHubUsername:  username,
		PullRequests:    prs,
		ContributedOrgs: contributed,
		StarredRepos:    starredRepos,
		Articles:        articles,
		Data:            additionalData,
	}); err != nil {
		return fmt.Errorf("failed to execute template: %w", err)
	}

	// Add a new line at the end of the buffer
	buf.WriteString("\n")

	// Write the output to stdout but get the size before doing so
	bytesWritten := buf.Len()
	if _, err := buf.WriteTo(os.Stdout); err != nil {
		return fmt.Errorf("failed to write output: %w", err)
	}

	// Print how much content was generated
	format := "January 2, 2006 @ 15:04:05 MST"

	fmt.Fprintf(
		os.Stderr,
		"Generated %s bytes of content on %s. Took %s.\n",
		formatNumber(bytesWritten),
		time.Now().Format(format),
		time.Since(start).Round(time.Millisecond),
	)
	return nil
}

func envdefault(key, def string) string {
	if val := strings.TrimSpace(os.Getenv(key)); val != "" {
		return val
	}

	return def
}

func envintdefault(key string, defval int) int {
	if val := strings.TrimSpace(os.Getenv(key)); val != "" {
		if i, err := strconv.Atoi(val); err == nil {
			return i
		}
	}

	return defval
}

func envbooldefault(key string, defval bool) bool {
	if val := strings.TrimSpace(os.Getenv(key)); val != "" {
		if b, err := strconv.ParseBool(val); err == nil {
			return b
		}
	}

	return defval
}
