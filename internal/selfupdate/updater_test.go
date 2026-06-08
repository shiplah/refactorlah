package selfupdate

import (
	"archive/tar"
	"archive/zip"
	"bytes"
	"compress/gzip"
	"context"
	"crypto/sha256"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"refactorlah/internal/buildinfo"
)

func TestUpdaterCheckDetectsAvailableRelease(t *testing.T) {
	updater := &Updater{
		BuildInfo: buildinfo.Info{
			Version:      "v1.0.0",
			Distribution: buildinfo.DistributionGitHubRelease,
			GOOS:         "darwin",
			GOARCH:       "arm64",
		},
		Executable: "/tmp/refactorlah",
		Locator: fakeReleaseLocator{
			release: Release{
				TagName: "v1.1.0",
				HTMLURL: "https://example.test/releases/v1.1.0",
				Assets: []Asset{
					{Name: "refactorlah_darwin-arm64.tar.gz", BrowserDownloadURL: "https://example.test/archive"},
					{Name: checksumAssetName, BrowserDownloadURL: "https://example.test/checksums"},
				},
			},
		},
	}

	result, err := updater.Check(t.Context(), CheckOptions{})
	if err != nil {
		t.Fatalf("check for updates: %v", err)
	}
	if !result.UpdateAvailable {
		t.Fatalf("expected update to be available, got %#v", result)
	}
	if result.TargetVersion != "v1.1.0" {
		t.Fatalf("unexpected target version: %#v", result)
	}
}

func TestUpdaterApplyReplacesExecutableOnNonWindows(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("non-Windows replacement test")
	}

	archiveName, err := releaseArchiveName(runtime.GOOS, runtime.GOARCH)
	if err != nil {
		t.Skipf("unsupported host target for updater test: %v", err)
	}

	newBinary := []byte("new-binary-content")
	archiveContent := mustCreateArchive(t, archiveName, binaryName(runtime.GOOS), newBinary)
	checksumContent := []byte(fmt.Sprintf("%x  %s\n", sha256Bytes(archiveContent), archiveName))

	tempDir := t.TempDir()
	executablePath := filepath.Join(tempDir, binaryName(runtime.GOOS))
	if err := os.WriteFile(executablePath, []byte("old-binary-content"), 0o755); err != nil {
		t.Fatalf("write current executable: %v", err)
	}

	release := Release{
		TagName: "v1.1.0",
		HTMLURL: "https://example.test/releases/v1.1.0",
		Assets: []Asset{
			{Name: archiveName, BrowserDownloadURL: "https://example.test/archive"},
			{Name: checksumAssetName, BrowserDownloadURL: "https://example.test/checksums"},
		},
	}

	updater := &Updater{
		BuildInfo: buildinfo.Info{
			Version:      "v1.0.0",
			Distribution: buildinfo.DistributionGitHubRelease,
			GOOS:         runtime.GOOS,
			GOARCH:       runtime.GOARCH,
		},
		Executable: executablePath,
		Locator:    fakeReleaseLocator{release: release},
		Downloader: fakeDownloader{
			assets: map[string][]byte{
				"https://example.test/archive":   archiveContent,
				"https://example.test/checksums": checksumContent,
			},
		},
		Stdout: io.Discard,
		Stderr: io.Discard,
	}

	result, err := updater.Apply(t.Context(), CheckOptions{})
	if err != nil {
		t.Fatalf("apply update: %v", err)
	}
	if !result.Updated || result.Staged {
		t.Fatalf("unexpected apply result: %#v", result)
	}

	content, err := os.ReadFile(executablePath)
	if err != nil {
		t.Fatalf("read replaced executable: %v", err)
	}
	if !bytes.Equal(content, newBinary) {
		t.Fatalf("unexpected executable content after update: %q", string(content))
	}
}

func TestUpdaterApplyUsesOneReleaseLookup(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("non-Windows replacement test")
	}

	archiveName, err := releaseArchiveName(runtime.GOOS, runtime.GOARCH)
	if err != nil {
		t.Skipf("unsupported host target for updater test: %v", err)
	}

	archiveContent := mustCreateArchive(t, archiveName, binaryName(runtime.GOOS), []byte("new-binary-content"))
	checksumContent := []byte(fmt.Sprintf("%x  %s\n", sha256Bytes(archiveContent), archiveName))

	tempDir := t.TempDir()
	executablePath := filepath.Join(tempDir, binaryName(runtime.GOOS))
	if err := os.WriteFile(executablePath, []byte("old-binary-content"), 0o755); err != nil {
		t.Fatalf("write current executable: %v", err)
	}

	locator := &countingReleaseLocator{
		release: Release{
			TagName: "v1.1.0",
			HTMLURL: "https://example.test/releases/v1.1.0",
			Assets: []Asset{
				{Name: archiveName, BrowserDownloadURL: "https://example.test/archive"},
				{Name: checksumAssetName, BrowserDownloadURL: "https://example.test/checksums"},
			},
		},
	}
	updater := &Updater{
		BuildInfo: buildinfo.Info{
			Version:      "v1.0.0",
			Distribution: buildinfo.DistributionGitHubRelease,
			GOOS:         runtime.GOOS,
			GOARCH:       runtime.GOARCH,
		},
		Executable: executablePath,
		Locator:    locator,
		Downloader: fakeDownloader{
			assets: map[string][]byte{
				"https://example.test/archive":   archiveContent,
				"https://example.test/checksums": checksumContent,
			},
		},
		Stdout: io.Discard,
		Stderr: io.Discard,
	}

	if _, err := updater.Apply(t.Context(), CheckOptions{}); err != nil {
		t.Fatalf("apply update: %v", err)
	}
	if locator.latestCalls != 1 {
		t.Fatalf("expected one latest release lookup, got %d", locator.latestCalls)
	}
}

func TestUpdaterApplyPlanDoesNotRequireDownloaderWhenUpToDate(t *testing.T) {
	updater := &Updater{}
	result, err := updater.ApplyPlan(t.Context(), UpdatePlan{
		CheckResult: CheckResult{
			CurrentVersion: "v1.0.0",
			TargetVersion:  "v1.0.0",
			UpToDate:       true,
		},
	})
	if err != nil {
		t.Fatalf("apply no-op update plan: %v", err)
	}
	if result.Updated {
		t.Fatalf("did not expect no-op update plan to replace executable: %#v", result)
	}
}

func TestReplacementHelperReplacesTargetFile(t *testing.T) {
	tempDir := t.TempDir()
	targetPath := filepath.Join(tempDir, "refactorlah")
	sourcePath := filepath.Join(tempDir, "refactorlah.new")

	if err := os.WriteFile(targetPath, []byte("old"), 0o755); err != nil {
		t.Fatalf("write target fixture: %v", err)
	}
	if err := os.WriteFile(sourcePath, []byte("new"), 0o755); err != nil {
		t.Fatalf("write source fixture: %v", err)
	}

	helper := &ReplacementHelper{}
	exitCode := helper.Run(context.Background(), []string{"--target", targetPath, "--source", sourcePath}, io.Discard)
	if exitCode != 0 {
		t.Fatalf("unexpected helper exit code: %d", exitCode)
	}

	content, err := os.ReadFile(targetPath)
	if err != nil {
		t.Fatalf("read replaced target: %v", err)
	}
	if string(content) != "new" {
		t.Fatalf("unexpected replacement content: %q", string(content))
	}
}

func TestExtractBinaryFromArchiveRequiresReleasePackageLayout(t *testing.T) {
	archiveName := "refactorlah_darwin-arm64.tar.gz"
	archiveContent := mustCreateArchiveWithPath(t, archiveName, "other/refactorlah", []byte("unexpected"))
	destinationPath := filepath.Join(t.TempDir(), "refactorlah")

	err := extractBinaryFromArchive(archiveName, archiveContent, destinationPath, "refactorlah")
	if err == nil {
		t.Fatal("expected archive layout error")
	}
	if !strings.Contains(err.Error(), "refactorlah_darwin-arm64/refactorlah") {
		t.Fatalf("expected exact package path in error, got %v", err)
	}
}

type fakeReleaseLocator struct {
	release Release
}

func (l fakeReleaseLocator) Latest(_ context.Context) (Release, error) {
	return l.release, nil
}

func (l fakeReleaseLocator) ByTag(_ context.Context, _ string) (Release, error) {
	return l.release, nil
}

type countingReleaseLocator struct {
	release     Release
	latestCalls int
	tagCalls    int
}

func (l *countingReleaseLocator) Latest(_ context.Context) (Release, error) {
	l.latestCalls++
	return l.release, nil
}

func (l *countingReleaseLocator) ByTag(_ context.Context, _ string) (Release, error) {
	l.tagCalls++
	return l.release, nil
}

type fakeDownloader struct {
	assets map[string][]byte
}

func (d fakeDownloader) Download(_ context.Context, assetURL string) ([]byte, error) {
	content, ok := d.assets[assetURL]
	if !ok {
		return nil, fmt.Errorf("unexpected asset url %s", assetURL)
	}
	return content, nil
}

func mustCreateArchive(t *testing.T, archiveName string, binaryFileName string, binaryContent []byte) []byte {
	t.Helper()

	expectedPath, err := expectedBinaryArchivePath(archiveName, binaryFileName)
	if err != nil {
		t.Fatalf("build expected archive path: %v", err)
	}

	return mustCreateArchiveWithPath(t, archiveName, expectedPath, binaryContent)
}

func mustCreateArchiveWithPath(t *testing.T, archiveName string, archivePath string, binaryContent []byte) []byte {
	t.Helper()

	switch {
	case strings.HasSuffix(archiveName, ".zip"):
		buffer := new(bytes.Buffer)
		writer := zip.NewWriter(buffer)
		file, err := writer.Create(archivePath)
		if err != nil {
			t.Fatalf("create zip member: %v", err)
		}
		if _, err := file.Write(binaryContent); err != nil {
			t.Fatalf("write zip member: %v", err)
		}
		if err := writer.Close(); err != nil {
			t.Fatalf("close zip archive: %v", err)
		}
		return buffer.Bytes()
	case strings.HasSuffix(archiveName, ".tar.gz"):
		buffer := new(bytes.Buffer)
		gzipWriter := gzip.NewWriter(buffer)
		tarWriter := tar.NewWriter(gzipWriter)
		header := &tar.Header{
			Name: archivePath,
			Mode: 0o755,
			Size: int64(len(binaryContent)),
		}
		if err := tarWriter.WriteHeader(header); err != nil {
			t.Fatalf("write tar header: %v", err)
		}
		if _, err := tarWriter.Write(binaryContent); err != nil {
			t.Fatalf("write tar body: %v", err)
		}
		if err := tarWriter.Close(); err != nil {
			t.Fatalf("close tar writer: %v", err)
		}
		if err := gzipWriter.Close(); err != nil {
			t.Fatalf("close gzip writer: %v", err)
		}
		return buffer.Bytes()
	default:
		t.Fatalf("unsupported archive fixture %s", archiveName)
		return nil
	}
}

func sha256Bytes(content []byte) []byte {
	sum := sha256.Sum256(content)
	return sum[:]
}
