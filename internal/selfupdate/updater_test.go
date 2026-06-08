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
			Distribution: "github-release",
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
			Distribution: "github-release",
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

type fakeReleaseLocator struct {
	release Release
}

func (l fakeReleaseLocator) Latest(_ context.Context) (Release, error) {
	return l.release, nil
}

func (l fakeReleaseLocator) ByTag(_ context.Context, _ string) (Release, error) {
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

	switch {
	case strings.HasSuffix(archiveName, ".zip"):
		buffer := new(bytes.Buffer)
		writer := zip.NewWriter(buffer)
		file, err := writer.Create("package/" + binaryFileName)
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
	default:
		buffer := new(bytes.Buffer)
		gzipWriter := gzip.NewWriter(buffer)
		tarWriter := tar.NewWriter(gzipWriter)
		header := &tar.Header{
			Name: "package/" + binaryFileName,
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
	}
}

func sha256Bytes(content []byte) []byte {
	sum := sha256.Sum256(content)
	return sum[:]
}
