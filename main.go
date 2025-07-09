package main

import (
	"bytes"
	"context"
	"errors"
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

// Custom error types for better error handling
var (
	ErrTemplateNotFound = errors.New("template file not found")
	ErrTemplateInvalid  = errors.New("template is invalid")
	ErrDataFileInvalid  = errors.New("data file is invalid")
	ErrAPIRequest       = errors.New("API request failed")
	ErrDataFetch        = errors.New("data fetch failed")
)

// ConfigError represents configuration-related errors
type ConfigError struct {
	Field string
	Value any
	Err   error
}

func (e ConfigError) Error() string {
	return fmt.Sprintf("configuration error for field %q with value %q: %v", e.Field, e.Value, e.Err)
}

func (e ConfigError) Unwrap() error {
	return e.Err
}

func main() {
	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, "Error:", err)
		os.Exit(1)
	}
}

// Config holds all configuration values for the application
type Config struct {
	Username         string
	RSSFeed          string
	TemplateFile     string
	MaxPRs           int
	MaxOrgs          int
	MaxStarred       int
	MaxArticles      int
	DisableRSS       bool
	DisablePRs       bool
	DisableStars     bool
	DisableDataFiles bool
	GitHubToken      string
}

// Validate validates the configuration values
func (c Config) Validate() error {
	if c.Username == "" {
		return ConfigError{Field: "Username", Value: c.Username, Err: errors.New("cannot be empty")}
	}

	if c.TemplateFile == "" {
		return ConfigError{Field: "TemplateFile", Value: c.TemplateFile, Err: errors.New("cannot be empty")}
	}

	if c.MaxPRs < 0 {
		return ConfigError{Field: "MaxPRs", Value: c.MaxPRs, Err: errors.New("must be non-negative")}
	}

	if c.MaxOrgs < 0 {
		return ConfigError{Field: "MaxOrgs", Value: c.MaxOrgs, Err: errors.New("must be non-negative")}
	}

	if c.MaxStarred < 0 {
		return ConfigError{Field: "MaxStarred", Value: c.MaxStarred, Err: errors.New("must be non-negative")}
	}

	if c.MaxArticles < 0 {
		return ConfigError{Field: "MaxArticles", Value: c.MaxArticles, Err: errors.New("must be non-negative")}
	}

	if !c.DisableRSS && c.RSSFeed == "" {
		return ConfigError{Field: "RSSFeed", Value: c.RSSFeed, Err: errors.New("cannot be empty when RSS is enabled")}
	}

	return nil
}

// LoadConfig loads configuration from environment variables
func LoadConfig() Config {
	return Config{
		Username:         envdefault("GITHUB_USERNAME", "patrickdappollonio"),
		RSSFeed:          envdefault("RSS_FEED", "https://www.patrickdap.com/index.xml"),
		TemplateFile:     envdefault("TEMPLATE_FILE", "template.md.gotmpl"),
		MaxPRs:           envintdefault("MAX_PULL_REQUESTS", 10),
		MaxOrgs:          envintdefault("MAX_CONTRIBUTED_ORGS", 5),
		MaxStarred:       envintdefault("MAX_STARRED_REPOS", 20),
		MaxArticles:      envintdefault("MAX_ARTICLES", 5),
		DisableRSS:       envbooldefault("DISABLE_RSS", false),
		DisablePRs:       envbooldefault("DISABLE_PULL_REQUESTS", false),
		DisableStars:     envbooldefault("DISABLE_STARRED_REPOS", false),
		DisableDataFiles: envbooldefault("DISABLE_DATA_FILES", false),
		GitHubToken:      os.Getenv("GITHUB_TOKEN"),
	}
}

// LoadAndValidateConfig loads and validates configuration from environment variables
func LoadAndValidateConfig() (Config, error) {
	config := LoadConfig()
	if err := config.Validate(); err != nil {
		return Config{}, fmt.Errorf("configuration validation failed: %w", err)
	}
	return config, nil
}

// TemplateData holds all data passed to the template
type TemplateData struct {
	GitHubUsername  string
	PullRequests    []PullRequest
	ContributedOrgs []string
	StarredRepos    []StarredRepo
	Articles        []Article
	Data            map[string]any
}

// DataLoader handles loading data from various sources
type DataLoader struct {
	config Config
	client *GitHubAPIClient
}

// NewDataLoader creates a new data loader
func NewDataLoader(config Config, client *GitHubAPIClient) *DataLoader {
	return &DataLoader{
		config: config,
		client: client,
	}
}

// LoadAdditionalData loads YAML data files from the current directory
func (dl *DataLoader) LoadAdditionalData() (map[string]any, error) {
	additionalData := make(map[string]any)

	if dl.config.DisableDataFiles {
		return additionalData, nil
	}

	dir, err := os.ReadDir(".")
	if err != nil {
		return nil, fmt.Errorf("failed to read directory: %w", err)
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
					return nil, fmt.Errorf("%w: file %s", ErrDataFileInvalid, fname)
				}

				return nil, fmt.Errorf("failed to read data file %q: %w", fname, err)
			}

			if err := yaml.Unmarshal(dataFile, &contents); err != nil {
				return nil, fmt.Errorf("%w: failed to decode YAML in file %s: %v", ErrDataFileInvalid, fname, err)
			}

			if contents != nil {
				additionalData[name] = contents
			}
		}
	}

	return additionalData, nil
}

// LoadAllData loads all data from GitHub API, RSS feeds, and local files
func (dl *DataLoader) LoadAllData(ctx context.Context) (*TemplateData, error) {
	var (
		eg           = errgroup.Group{}
		prs          []PullRequest
		contributed  []string
		starredRepos []StarredRepo
		articles     []Article
	)

	// Load additional data files
	additionalData, err := dl.LoadAdditionalData()
	if err != nil {
		return nil, err
	}

	// Fetch pull requests
	eg.Go(func() error {
		if dl.config.DisablePRs {
			return nil
		}

		var err error
		prs, contributed, err = getPullRequests(ctx, dl.client, dl.config.Username, dl.config.MaxPRs, dl.config.MaxOrgs)
		if err != nil {
			return fmt.Errorf("%w: failed to get pull requests for user %s: %v", ErrDataFetch, dl.config.Username, err)
		}
		return nil
	})

	// Fetch starred repos
	eg.Go(func() error {
		if dl.config.DisableStars {
			return nil
		}

		var err error
		starredRepos, err = getStarredRepos(ctx, dl.client, dl.config.Username, dl.config.MaxStarred)
		if err != nil {
			return fmt.Errorf("%w: failed to get starred repos for user %s: %v", ErrDataFetch, dl.config.Username, err)
		}
		return nil
	})

	// Fetch RSS feed
	eg.Go(func() error {
		if dl.config.DisableRSS {
			return nil
		}

		var err error
		articles, err = getArticles(dl.config.RSSFeed, dl.config.MaxArticles)
		if err != nil {
			return fmt.Errorf("%w: failed to read RSS feed %s: %v", ErrDataFetch, dl.config.RSSFeed, err)
		}
		return nil
	})

	// Wait for all goroutines to finish
	if err := eg.Wait(); err != nil {
		return nil, err
	}

	return &TemplateData{
		GitHubUsername:  dl.config.Username,
		PullRequests:    prs,
		ContributedOrgs: contributed,
		StarredRepos:    starredRepos,
		Articles:        articles,
		Data:            additionalData,
	}, nil
}

func run() error {
	// Load and validate configuration
	config, err := LoadAndValidateConfig()
	if err != nil {
		return err
	}

	// Get the start time
	start := time.Now()

	// Read the template file
	tplFile, err := os.ReadFile(config.TemplateFile)
	if err != nil {
		if os.IsNotExist(err) {
			return fmt.Errorf("%w: %s", ErrTemplateNotFound, config.TemplateFile)
		}

		return fmt.Errorf("failed to read template file %q: %w", config.TemplateFile, err)
	}

	// Create GitHub API client
	githubClient := NewGitHubAPIClient(config.GitHubToken)

	// Create data loader
	dataLoader := NewDataLoader(config, githubClient)

	// Load all data
	templateData, err := dataLoader.LoadAllData(context.Background())
	if err != nil {
		return err
	}

	// Create template with functions
	funcs := sprig.FuncMap()
	for k, v := range TemplateFunctions() {
		funcs[k] = v
	}

	tmpl, err := template.New("template").Funcs(funcs).Parse(string(tplFile))
	if err != nil {
		return fmt.Errorf("%w: %s", ErrTemplateInvalid, err)
	}

	// Execute template and generate output
	output, err := executeTemplate(tmpl, templateData)
	if err != nil {
		return err
	}

	// Output the result
	bytesWritten, err := writeOutput(output)
	if err != nil {
		return err
	}

	// Print statistics
	printStatistics(bytesWritten, start)
	return nil
}

// executeTemplate executes the template with the given data
func executeTemplate(tmpl *template.Template, data *TemplateData) ([]byte, error) {
	var buf bytes.Buffer

	if err := tmpl.Execute(&buf, data); err != nil {
		return nil, fmt.Errorf("failed to execute template: %w", err)
	}

	// Add a newline at the end
	buf.WriteString("\n")

	return buf.Bytes(), nil
}

// writeOutput writes the output to stdout
func writeOutput(output []byte) (int, error) {
	n, err := os.Stdout.Write(output)
	if err != nil {
		return 0, fmt.Errorf("failed to write output: %w", err)
	}
	return n, nil
}

// printStatistics prints generation statistics to stderr
func printStatistics(bytesWritten int, start time.Time) {
	format := "January 2, 2006 @ 15:04:05 MST"

	fmt.Fprintf(
		os.Stderr,
		"Generated %s bytes of content on %s. Took %s.\n",
		formatNumber(bytesWritten),
		time.Now().Format(format),
		time.Since(start).Round(time.Millisecond),
	)
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
