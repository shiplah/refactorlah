package selfupdate

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

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
