package selfupdate

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"path"
	"strings"
	"time"
)

const (
	defaultGitHubAPIBaseURL = "https://api.github.com/"
	defaultReleaseOwner     = "NickSdot"
	defaultReleaseRepo      = "refactorlah"
	checksumAssetName       = "refactorlah_checksums.txt"
	defaultMaxDownloadBytes = 128 * 1024 * 1024
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
	BaseURL          string
	Owner            string
	Repo             string
	HTTPClient       *http.Client
	MaxDownloadBytes int64
}

func NewGitHubClient() *GitHubClient {
	return &GitHubClient{
		BaseURL: defaultGitHubAPIBaseURL,
		Owner:   defaultReleaseOwner,
		Repo:    defaultReleaseRepo,
		HTTPClient: &http.Client{
			Timeout: 30 * time.Second,
		},
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

	limit := c.maxDownloadBytes()
	content, err := io.ReadAll(io.LimitReader(response.Body, limit+1))
	if err != nil {
		return nil, fmt.Errorf("download asset: %w", err)
	}
	if int64(len(content)) > limit {
		return nil, fmt.Errorf("download asset: response exceeds %d bytes", limit)
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

	response, err := client.Do(request)
	if err != nil {
		return nil, fmt.Errorf("request %s: %w", requestURL, err)
	}

	return response, nil
}

func (c *GitHubClient) releaseURL(relativePath string) string {
	base := firstNonEmpty(c.BaseURL, defaultGitHubAPIBaseURL)
	owner := firstNonEmpty(c.Owner, defaultReleaseOwner)
	repo := firstNonEmpty(c.Repo, defaultReleaseRepo)
	parsedBase, err := url.Parse(base)
	if err != nil {
		return defaultGitHubAPIBaseURL + "repos/" + owner + "/" + repo + "/" + strings.TrimLeft(relativePath, "/")
	}

	basePath := path.Join(parsedBase.Path, "repos", owner, repo)
	parsedBase.Path = strings.TrimRight(basePath, "/") + "/" + strings.TrimLeft(relativePath, "/")
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

func (c *GitHubClient) maxDownloadBytes() int64 {
	if c.MaxDownloadBytes > 0 {
		return c.MaxDownloadBytes
	}

	return defaultMaxDownloadBytes
}
