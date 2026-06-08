package cli

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"refactorlah/internal/buildinfo"
	"refactorlah/internal/selfupdate"
)

func TestUpdateCommandCheckJSON(t *testing.T) {
	command := newUpdateCommandForTest(t, "darwin", "arm64")

	var stdout bytes.Buffer
	var stderr bytes.Buffer
	exitCode := command.Run(t.Context(), []string{"--check", "--json"}, &stdout, &stderr)
	if exitCode != ExitSuccess {
		t.Fatalf("unexpected exit code: %d", exitCode)
	}

	var payload map[string]any
	if err := json.Unmarshal(stdout.Bytes(), &payload); err != nil {
		t.Fatalf("decode json output: %v", err)
	}
	if payload["update_available"] != true {
		t.Fatalf("expected update_available=true, got %#v", payload)
	}
	if payload["target_version"] != "v1.1.0" {
		t.Fatalf("unexpected target version: %#v", payload)
	}
	if stderr.Len() != 0 {
		t.Fatalf("expected empty stderr, got %q", stderr.String())
	}
}

func TestUpdateCommandCancelLeavesExecutableUntouched(t *testing.T) {
	command := newUpdateCommandForTest(t, "darwin", "arm64")
	command.stdin = strings.NewReader("n\n")

	executablePath := command.newUpdaterMust().Executable
	originalContent, err := os.ReadFile(executablePath)
	if err != nil {
		t.Fatalf("read executable fixture: %v", err)
	}

	var stdout bytes.Buffer
	var stderr bytes.Buffer
	exitCode := command.Run(t.Context(), nil, &stdout, &stderr)
	if exitCode != ExitSuccess {
		t.Fatalf("unexpected exit code: %d", exitCode)
	}
	if !strings.Contains(stdout.String(), "Update cancelled.") {
		t.Fatalf("expected cancellation message, got %q", stdout.String())
	}
	if stderr.Len() != 0 {
		t.Fatalf("expected empty stderr, got %q", stderr.String())
	}

	currentContent, err := os.ReadFile(executablePath)
	if err != nil {
		t.Fatalf("read executable after cancel: %v", err)
	}
	if !bytes.Equal(originalContent, currentContent) {
		t.Fatal("expected cancelled update to leave executable untouched")
	}
}

func newUpdateCommandForTest(t *testing.T, goos string, goarch string) *UpdateCommand {
	t.Helper()

	tempDir := t.TempDir()
	executablePath := filepath.Join(tempDir, "refactorlah")
	if err := os.WriteFile(executablePath, []byte("current-binary"), 0o755); err != nil {
		t.Fatalf("write executable fixture: %v", err)
	}

	command := NewUpdateCommand()
	command.newUpdater = func() (*selfupdate.Updater, error) {
		return &selfupdate.Updater{
			BuildInfo: buildinfo.Info{
				Version:      "v1.0.0",
				Commit:       "abc1234",
				BuildDate:    "2026-06-08T10:11:12Z",
				Distribution: "github-release",
				GOOS:         goos,
				GOARCH:       goarch,
			},
			Executable: executablePath,
			Locator: staticReleaseLocator{
				release: selfupdate.Release{
					TagName: "v1.1.0",
					HTMLURL: "https://example.test/releases/v1.1.0",
					Assets: []selfupdate.Asset{
						{Name: "refactorlah_darwin-arm64.tar.gz", BrowserDownloadURL: "https://example.test/assets/archive"},
						{Name: "refactorlah_checksums.txt", BrowserDownloadURL: "https://example.test/assets/checksums"},
					},
				},
			},
			Downloader: staticDownloader{},
			Stdout:     io.Discard,
			Stderr:     io.Discard,
		}, nil
	}

	return command
}

func (c *UpdateCommand) newUpdaterMust() *selfupdate.Updater {
	updater, err := c.newUpdater()
	if err != nil {
		panic(err)
	}
	return updater
}

type staticReleaseLocator struct {
	release selfupdate.Release
}

func (l staticReleaseLocator) Latest(_ context.Context) (selfupdate.Release, error) {
	return l.release, nil
}

func (l staticReleaseLocator) ByTag(_ context.Context, _ string) (selfupdate.Release, error) {
	return l.release, nil
}

type staticDownloader struct{}

func (staticDownloader) Download(_ context.Context, _ string) ([]byte, error) {
	return nil, nil
}
