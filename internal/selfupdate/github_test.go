package selfupdate

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestNewGitHubClientUsesDefaultsAndConfiguredToken(t *testing.T) {
	t.Setenv("REFACTORLAH_GITHUB_TOKEN", "primary-token")
	t.Setenv("GH_TOKEN", "secondary-token")
	t.Setenv("GITHUB_TOKEN", "fallback-token")

	client := NewGitHubClient()
	if client.BaseURL != defaultGitHubAPIBaseURL {
		t.Fatalf("unexpected base URL: %s", client.BaseURL)
	}
	if client.Owner != defaultReleaseOwner || client.Repo != defaultReleaseRepo {
		t.Fatalf("unexpected repository: %s/%s", client.Owner, client.Repo)
	}
	if client.HTTPClient == nil {
		t.Fatal("expected HTTP client")
	}
	if client.Token != "primary-token" {
		t.Fatalf("unexpected token: %q", client.Token)
	}
}

func TestGitHubClientFetchesLatestRelease(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		if request.URL.Path != "/api/repos/acme/tool/releases/latest" {
			t.Fatalf("unexpected release path: %s", request.URL.Path)
		}
		if request.Header.Get("Authorization") != "Bearer secret-token" {
			t.Fatalf("unexpected authorization header: %q", request.Header.Get("Authorization"))
		}
		if request.Header.Get("Accept") != "application/vnd.github+json" {
			t.Fatalf("unexpected accept header: %q", request.Header.Get("Accept"))
		}
		if request.Header.Get("User-Agent") != "refactorlah-self-update" {
			t.Fatalf("unexpected user-agent header: %q", request.Header.Get("User-Agent"))
		}

		writer.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(writer).Encode(Release{
			TagName: "v1.2.3",
			HTMLURL: "https://example.test/releases/v1.2.3",
			Assets: []Asset{
				{Name: "refactorlah_darwin-arm64.tar.gz", BrowserDownloadURL: "https://example.test/archive"},
			},
		})
	}))
	defer server.Close()

	client := &GitHubClient{
		BaseURL:    server.URL + "/api/",
		Owner:      "acme",
		Repo:       "tool",
		HTTPClient: server.Client(),
		Token:      "secret-token",
	}

	release, err := client.Latest(t.Context())
	if err != nil {
		t.Fatalf("fetch latest release: %v", err)
	}
	if release.TagName != "v1.2.3" {
		t.Fatalf("unexpected release: %#v", release)
	}
	if len(release.Assets) != 1 || release.Assets[0].Name != "refactorlah_darwin-arm64.tar.gz" {
		t.Fatalf("unexpected release assets: %#v", release.Assets)
	}
}

func TestGitHubClientByTagNormalisesNumericTags(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		if request.URL.Path != "/repos/NickSdot/refactorlah/releases/tags/v1.2.3" {
			t.Fatalf("unexpected release tag path: %s", request.URL.Path)
		}

		writer.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(writer).Encode(Release{TagName: "v1.2.3"})
	}))
	defer server.Close()

	client := &GitHubClient{
		BaseURL:    server.URL + "/",
		HTTPClient: server.Client(),
	}

	release, err := client.ByTag(t.Context(), " 1.2.3 ")
	if err != nil {
		t.Fatalf("fetch release by tag: %v", err)
	}
	if release.TagName != "v1.2.3" {
		t.Fatalf("unexpected release: %#v", release)
	}
}

func TestGitHubClientByTagRejectsMissingTag(t *testing.T) {
	client := &GitHubClient{}

	_, err := client.ByTag(t.Context(), "  ")
	if err == nil {
		t.Fatal("expected missing tag error")
	}
	if !strings.Contains(err.Error(), "missing release tag") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestGitHubClientFetchReleaseReportsHTTPAndDecodeErrors(t *testing.T) {
	tests := []struct {
		name          string
		status        int
		body          string
		expectedError string
	}{
		{
			name:          "http error",
			status:        http.StatusForbidden,
			body:          "rate limited",
			expectedError: "fetch release metadata: unexpected status 403 Forbidden: rate limited",
		},
		{
			name:          "invalid json",
			status:        http.StatusOK,
			body:          "{not-json",
			expectedError: "decode release metadata",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, _ *http.Request) {
				writer.WriteHeader(test.status)
				_, _ = writer.Write([]byte(test.body))
			}))
			defer server.Close()

			client := &GitHubClient{
				BaseURL:    server.URL,
				HTTPClient: server.Client(),
			}

			_, err := client.Latest(t.Context())
			if err == nil {
				t.Fatal("expected release metadata error")
			}
			if !strings.Contains(err.Error(), test.expectedError) {
				t.Fatalf("unexpected error: %v", err)
			}
		})
	}
}

func TestGitHubClientDownloadReportsHTTPStatus(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, _ *http.Request) {
		writer.WriteHeader(http.StatusNotFound)
		_, _ = writer.Write([]byte("missing asset"))
	}))
	defer server.Close()

	client := &GitHubClient{HTTPClient: server.Client()}

	_, err := client.Download(t.Context(), server.URL)
	if err == nil {
		t.Fatal("expected download status error")
	}
	if !strings.Contains(err.Error(), "download asset: unexpected status 404 Not Found: missing asset") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestGitHubClientDownloadRejectsOversizedAsset(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		_, _ = writer.Write([]byte("too large"))
	}))
	defer server.Close()

	client := &GitHubClient{
		HTTPClient:       server.Client(),
		MaxDownloadBytes: 3,
	}

	_, err := client.Download(t.Context(), server.URL)
	if err == nil {
		t.Fatal("expected oversized download error")
	}
	if !strings.Contains(err.Error(), "response exceeds 3 bytes") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestGitHubClientDownloadAcceptsExactSizeLimit(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		_, _ = writer.Write([]byte("exact"))
	}))
	defer server.Close()

	client := &GitHubClient{
		HTTPClient:       server.Client(),
		MaxDownloadBytes: 5,
	}

	content, err := client.Download(t.Context(), server.URL)
	if err != nil {
		t.Fatalf("download exact-size asset: %v", err)
	}
	if string(content) != "exact" {
		t.Fatalf("unexpected content: %q", content)
	}
}

func TestGitHubClientReleaseURLHonoursConfiguredRepository(t *testing.T) {
	client := &GitHubClient{
		BaseURL: "https://github.example.test/api/v3/",
		Owner:   "acme",
		Repo:    "tool",
	}

	got := client.releaseURL("/releases/latest")
	want := "https://github.example.test/api/v3/repos/acme/tool/releases/latest"
	if got != want {
		t.Fatalf("unexpected release URL:\nwant %s\n got %s", want, got)
	}
}
