package selfupdate

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path"
	"strings"
	"time"
)

const (
	defaultGitHubAPIBaseURL = "https://api.github.com/"
	defaultReleaseOwner     = "NickSdot"
	defaultReleaseRepo      = "refactorlah"
	checksumAssetName       = "refactorlah_checksums.txt"
)

type Release struct {
	TagName string  `json:"tag_name"`
	HTMLURL string  `json:"html_url"`
	Assets  []Asset `json:"assets"`
}

type Asset struct {
	Name               string `json:"name"`
	BrowserDownloadURL string `json:"browser_download_url"`
}

type ReleaseLocator interface {
	Latest(ctx context.Context) (Release, error)
	ByTag(ctx context.Context, tag string) (Release, error)
}

type GitHubClient struct {
	BaseURL    string
	Owner      string
	Repo       string
	HTTPClient *http.Client
	Token      string
}

func NewGitHubClient() *GitHubClient {
	return &GitHubClient{
		BaseURL: defaultGitHubAPIBaseURL,
		Owner:   defaultReleaseOwner,
		Repo:    defaultReleaseRepo,
		HTTPClient: &http.Client{
			Timeout: 30 * time.Second,
		},
		Token: firstNonEmpty(
			os.Getenv("REFACTORLAH_GITHUB_TOKEN"),
			os.Getenv("GH_TOKEN"),
			os.Getenv("GITHUB_TOKEN"),
		),
	}
}

func (c *GitHubClient) Latest(ctx context.Context) (Release, error) {
	return c.fetchRelease(ctx, "releases/latest")
}

func (c *GitHubClient) ByTag(ctx context.Context, tag string) (Release, error) {
	normalized := strings.TrimSpace(tag)
	if normalized == "" {
		return Release{}, fmt.Errorf("missing release tag")
	}
	if !strings.HasPrefix(normalized, "v") && normalized[0] >= '0' && normalized[0] <= '9' {
		normalized = "v" + normalized
	}

	return c.fetchRelease(ctx, "releases/tags/"+url.PathEscape(normalized))
}

func (c *GitHubClient) Download(ctx context.Context, assetURL string) ([]byte, error) {
	response, err := c.doRequest(ctx, assetURL)
	if err != nil {
		return nil, err
	}
	defer response.Body.Close()

	if response.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(io.LimitReader(response.Body, 4<<10))
		return nil, fmt.Errorf("download asset: unexpected status %s: %s", response.Status, strings.TrimSpace(string(body)))
	}

	content, err := io.ReadAll(response.Body)
	if err != nil {
		return nil, fmt.Errorf("download asset: %w", err)
	}

	return content, nil
}

func (c *GitHubClient) fetchRelease(ctx context.Context, relativePath string) (Release, error) {
	response, err := c.doRequest(ctx, c.releaseURL(relativePath))
	if err != nil {
		return Release{}, err
	}
	defer response.Body.Close()

	if response.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(io.LimitReader(response.Body, 4<<10))
		return Release{}, fmt.Errorf("fetch release metadata: unexpected status %s: %s", response.Status, strings.TrimSpace(string(body)))
	}

	var release Release
	if err := json.NewDecoder(response.Body).Decode(&release); err != nil {
		return Release{}, fmt.Errorf("decode release metadata: %w", err)
	}

	return release, nil
}

func (c *GitHubClient) doRequest(ctx context.Context, requestURL string) (*http.Response, error) {
	client := c.HTTPClient
	if client == nil {
		client = NewGitHubClient().HTTPClient
	}

	request, err := http.NewRequestWithContext(ctx, http.MethodGet, requestURL, nil)
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}

	request.Header.Set("Accept", "application/vnd.github+json")
	request.Header.Set("User-Agent", "refactorlah-self-update")
	if c.Token != "" {
		request.Header.Set("Authorization", "Bearer "+c.Token)
	}

	response, err := client.Do(request)
	if err != nil {
		return nil, fmt.Errorf("request %s: %w", requestURL, err)
	}

	return response, nil
}

func (c *GitHubClient) releaseURL(relativePath string) string {
	base := firstNonEmpty(c.BaseURL, defaultGitHubAPIBaseURL)
	parsedBase, err := url.Parse(base)
	if err != nil {
		return defaultGitHubAPIBaseURL + "repos/" + c.Owner + "/" + c.Repo + "/" + relativePath
	}

	parsedBase.Path = path.Join(parsedBase.Path, "repos", firstNonEmpty(c.Owner, defaultReleaseOwner), firstNonEmpty(c.Repo, defaultReleaseRepo), relativePath)
	if !strings.HasSuffix(parsedBase.Path, relativePath) && !strings.HasSuffix(parsedBase.Path, relativePath+"/") {
		parsedBase.Path += "/" + relativePath
	}
	return parsedBase.String()
}

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		if strings.TrimSpace(value) != "" {
			return value
		}
	}

	return ""
}
