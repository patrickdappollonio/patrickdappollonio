# How to use this project

- [How to use this project](#how-to-use-this-project)
    - [Customizing your GitHub profile](#customizing-your-github-profile)
      - [How does customizing work?](#how-does-customizing-work)
      - [Adding images](#adding-images)
      - [Preventing images from linking to themselves](#preventing-images-from-linking-to-themselves)
    - [Using this project](#using-this-project)
      - [Configuring the application](#configuring-the-application)
      - [Testing the configuration](#testing-the-configuration)
      - [Contextual details](#contextual-details)
      - [Data files](#data-files)
      - [Template functions](#template-functions)
      - [Scheduling updates](#scheduling-updates)
      - [Updating the app](#updating-the-app)

If you've stumbled on this project, you're probably wondering how to use it to improve your GitHub profile with some dynamism. This document will guide you through the process of setting up your profile and use it.

### Customizing your GitHub profile

To start, you need a repository with the same name as your GitHub handle. If your GitHub username is `octocat`, then you need a repo called `octocat` too, yielding a URL like `github.com/octocat/octocat`.

Or in my case, the repo would be `github.com/patrickdappollonio/patrickdappollonio` (this repository where you're reading this doc).

You can create a new repository by [clicking here](https://github.com/new).

#### How does customizing work?

In short, GitHub allows you to customize your profile by creating a `README.md` file in the repository with your name. Anything in that `README` will be loaded whenever someone opens your GitHub profile.

In Layman's terms, if I create a file in `github.com/patrickdappollonio/patrickdappollonio/tree/main/README.md`, the contents of that file will be displayed in my profile, at `github.com/patrickdappollonio`.

#### Adding images

You can add images to GitHub readme files by simply linking them either in Markdown format:

```md
![Alt text](https://example.com/image.jpg)
```

Or in HTML format:

```html
<img src="https://example.com/image.jpg" alt="Alt text" width="200"/>
```

> [!WARNING]
> Markdown images by default will always link to themselves. Essentially, if someone clicks on the image, they will be taken to the image URL. If you want to avoid this, use the trick below.

#### Preventing images from linking to themselves

This is a cheat, and it's up to GitHub to ensure it keeps working. Providing a `<picture>` html element with a `<source>` tag for both dark and light modes will prevent the image from linking to itself:

```html
<picture>
  <source media="(prefers-color-scheme: dark)" srcset="images/icons-dark.png">
  <source media="(prefers-color-scheme: light)" srcset="images/icons-light.png">
  <img src="images/icons-dark.png" alt="technologies I use">
</picture>
```

### Using this project

> [!IMPORTANT]
> Experience with Go templates is required to use this project. If you're not familiar with Go templates, you can learn more about them [here](https://pkg.go.dev/text/template).

Now that you know how to update your GitHub profile, let's make it more dynamic by using the `patrickdappollonio` tool. You need 2 files minimum in your GitHub repository to make it work:

* A readme "template", which will be used to generate-then-overwrite the real `README.md` (in my case, [it's the `template.md.gotmpl` file](template.md.gotmpl)). The file name does not matter.
* A GitHub action file to run the tool and update the `README.md` file on a cadence (in my case, [it's the `.github/workflows/schedule.yaml` file](.github/workflows/schedule.yaml)).

You can copy these two files to the same locations in your own repository. My readme also includes an image with all the technologies I use, but you don't have to include those images if you don't want to.

For easy finding: copy these two files to your GitHub repository:

```
https://github.com/patrickdappollonio/patrickdappollonio/blob/main/template.md.gotmpl
https://github.com/patrickdappollonio/patrickdappollonio/blob/main/.github/workflows/schedule.yaml
```

Finally, you need the application itself, which the workflow will automatically download for you, but you can also download it to your local machine to try out before committing it to your repository. You can download the latest release from the [releases page](https://github.com/patrickdappollonio/patrickdappollonio/releases).

#### Configuring the application

By default, the application receives all its configurations using environment variables. You can change these values by setting the following environment variables:

| Variable name           | Description                                                                                                                                                                                                                                                                   | Default value                          |
| ----------------------- | ----------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- | -------------------------------------- |
| `GITHUB_USERNAME`       | Your GitHub username                                                                                                                                                                                                                                                          | `patrickdappollonio`                   |
| `RSS_FEED`              | Your blog RSS feed                                                                                                                                                                                                                                                            | `https://www.patrickdap.com/index.xml` |
| `TEMPLATE_FILE`         | The template file to use                                                                                                                                                                                                                                                      | `template.md.gotmpl`                   |
| `MAX_PULL_REQUESTS`     | The maximum number of pull requests to show                                                                                                                                                                                                                                   | `10`                                   |
| `MAX_CONTRIBUTED_ORGS`  | The maximum number of unique (non-repeated) organizations to show where you've contributed code by sending Pull Requests to it. This is a _best effort_ since the maximum list of organizations might not be as big as what's provided in up to 100 records of pull requests. | `10`                                   |
| `MAX_STARRED_REPOS`     | The maximum number of starred repositories to show                                                                                                                                                                                                                            | `20`                                   |
| `MAX_ARTICLES`          | The maximum number of articles from the RSS feed to show                                                                                                                                                                                                                      | `5`                                    |
| `DISABLE_RSS`           | If true, disables the RSS feed from being shown                                                                                                                                                                                                                               | `false`                                |
| `DISABLE_PULL_REQUESTS` | If true, disables the pull requests from being shown                                                                                                                                                                                                                          | `false`                                |
| `DISABLE_STARRED_REPOS` | If true, disables the starred repositories from being shown, as well as the contributed organizations                                                                                                                                                                         | `false`                                |
| `DISABLE_DATA_FILES`    | If true, disables the data files from being read                                                                                                                                                                                                                              | `false`                                |

You can set these environment variables in the GitHub workflow. For example, to use Octocat's GitHub profile, note the `export GITHUB_USERNAME=octocat` line:

```yaml
- name: Update README with latest information
  run: |
    git config user.name "GitHub Actions"
    git config user.email "github-actions[bot]@users.noreply.github.com"
    export GITHUB_USERNAME=octocat
    patrickdappollonio > README.md
    git add README.md || echo "No changes to add"
    git commit -m "[ci skip] Updating README with latest information" || echo "No changes to commit"
    git push || echo "No changes to push"
```

#### Testing the configuration

You can run the application locally and see if it would generate an appropriate `README.md` file, simply download a release from the releases page then run it with the required parameters:

```bash
GITHUB_USERNAME=octocat ./patrickdappollonio
```

The contents will be outputted to the console. If you want to save them to a file, you can simply redirect the output to a file:

```bash
GITHUB_USERNAME=octocat ./patrickdappollonio > README.md
```

#### Contextual details

Like any Go template, all the information available to the template is stored under `.`. The following keys are available:

| Key name           | Type                     | Description                                                                                                                                         |
| ------------------ | ------------------------ |
| `.GitHubUsername`  | `string`                 | The GitHub username being used to generate the README file                                                                                          |
| `.PullRequests`    | `[]PullRequest`          | A list of pull requests made by the user, up to the maximum number specified in the configuration                                                   |
| `.ContributedOrgs` | `[]string`               | A list of organizations where the user has contributed code by sending Pull Requests to it, up to the maximum number specified in the configuration |
| `.StarredRepos`    | `[]StarredRepo`          | A list of starred repositories by the user, up to the maximum number specified in the configuration                                                 |
| `.Articles`        | `[]Article`              | A list of articles from the RSS feed, up to the maximum number specified in the configuration                                                       |
| `.Data`            | `map[string]interface{}` | A map of string to any, containing all the data files loaded into the application                                                                   |


The `PullRequest` struct has the following fields:

```go
type PullRequest struct {
    URL              string    `json:"html_url"`
    RepositoryAPIURL string    `json:"repository_url"`
    ID               int64     `json:"number"`
    Title            string    `json:"title"`
    State            string    `json:"state"`
    Locked           bool      `json:"locked"`
    Comments         int       `json:"comments"`
    CreatedAt        time.Time `json:"created_at"`
    UpdatedAt        time.Time `json:"updated_at"`
    ClosedAt         time.Time `json:"closed_at"`
    Draft            bool      `json:"draft"`
    Body             string    `json:"body"`
    PullRequest      struct {
        MergedAt time.Time `json:"merged_at"`
    } `json:"pull_request"`
}
func (p *PullRequest) Closed() bool
func (p *PullRequest) ContributedToOrg() string
func (p *PullRequest) Merged() bool
func (p *PullRequest) ProjectOrg() string
func (p *PullRequest) RepositoryName() string
func (p *PullRequest) RepositoryURL() string
```

The `StarredRepo` struct has the following fields:

```go
type StarredRepo struct {
    Name    string `json:"full_name"`
    Private bool   `json:"private"`
    URL     string `json:"html_url"`
    Stars   int    `json:"stargazers_count"`
    Owner   struct {
        User string `json:"login"`
    } `json:"owner"`
}
func (s *StarredRepo) IsOwned(username string) bool
func (s *StarredRepo) IsPrivate() bool
```

The `Article` struct has the following fields:

```go
type Article struct {
    Title string `xml:"title"`
    Link  string `xml:"link"`
    Date  string `xml:"pubDate"`
}
func (a *Article) GoDate() (time.Time, error)
```

Any of these fields can be accessed by using the dot-notation as common in Go templates.

#### Data files

Data files are YAML files in the current working directory (excluding subdirectories). The data is loaded into the template as a map of string to any. Data values are available under `.Data` and the file name without extension is used as the key.

For example, if you have a collection of links you want to show in your README, you can create a file called `links.yaml` with the following content:

```yaml
- name: GitHub
  url: https://github.com/patrickdappollonio
- name: LinkedIn
  url: https://www.linkedin.com/in/patrickdappollonio
- name: Twitter
  url: https://twitter.com/marlex
```

Then you can read this information by using the `.Data.links` key in your template:

```md
## Links

{{ range .Data.links }}
- [{{ .name }}]({{ .url }})
{{ end }}
```

#### Template functions

While limited, there are a few template functions available for you to use when writing your template. The following functions are available:

| Function name       | Description                                                                                                                                                          | Example                                                   |
| ------------------- | -------------------------------------------------------------------------------------------------------------------------------------------------------------------- | --------------------------------------------------------- |
| `formatDate`        | Formats a date string into a human-readable format of `January 02, 2006` (replacing each part with the corresponding date part)                                      | `formatDate now`                                          |
| `formatDateTime`    | Formats a date string into a human-readable format of `January 02, 2006 at 15:04:05 MST` (replacing each part with the corresponding date part)                      | `formatDateTime now`                                      |
| `formatNumber`      | Formats a number into a human-readable format with commas separating the thousands (e.g. `1,000`)                                                                    | `formatNumber 1000`                                       |
| `now`               | Returns the current time Go object (akin to `time.Now()`)                                                                                                            | `now`                                                     |
| `humanizeBigNumber` | Formats a number into a human-readable format with a suffix (e.g. `1K`, `1M`, `1B`). If the number is bigger than 1,000, commas are added to separate the thousands. | `humanizeBigNumber 1000`                                  |
| `sub`               | Subtracts two numbers and returns the result.                                                                                                                        | `sub 10 5`                                                |
| `ellipsize`         | Cuts a string to a maximum number of words, appending an ellipsis at the end.                                                                                        | `ellipsize (printf "%s" "This is a long string") 3`       |
| `ellipsizechars`    | Cuts a string to a maximum number of characters, appending an ellipsis at the end.                                                                                   | `ellipsizechars (printf "%s" "This is a long string") 10` |
| `take`              | Takes a slice of a list up to a maximum number of elements.                                                                                                          | `take 3 (seq 5)`                                          |
| `skip`              | Skips a number of elements from a list and returns the rest.                                                                                                         | `skip 2 (seq 5)`                                          |
| `seq`               | Creates a sequence of numbers from 1 to the desired maximum.                                                                                                         | `seq 5`                                                   |
| `max`               | Returns the maximum of two numbers.                                                                                                                                  | `max 5 10`                                                |

#### Scheduling updates

Since the code uses publicly available information, normal GitHub rate limits are at play. It is currently not possible to supply a custom token to the application, so it will use the default rate limits for unauthenticated requests.

On GitHub workflows, if you use the following settings, it'll update every 4 hours or whenever you trigger the workflow in the "Actions" tab:

```yaml
on:
  schedule:
    - cron: "0 */4 * * *" # every 4 hours
  workflow_dispatch:
```

#### Updating the app

Every now and then I might update the application to include new features or fix bugs. By default, if you're using my Workflow, the version is pinned:

```yaml
env:
  VERSION: "0.1.10"
```

You can change that value to any value from the [releases page](https://github.com/patrickdappollonio/patrickdappollonio/releases).

---

If you have any questions, feel free to open an issue in this repository. I'll be happy to help you out!
