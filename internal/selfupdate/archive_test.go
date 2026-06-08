package selfupdate

import (
	"archive/tar"
	"archive/zip"
	"bytes"
	"compress/gzip"
	"crypto/sha256"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestReleaseArchiveNameMatchesReleaseTargetMatrix(t *testing.T) {
	tests := []struct {
		name          string
		goos          string
		goarch        string
		expectedName  string
		expectedError string
	}{
		{
			name:         "macOS arm64",
			goos:         "darwin",
			goarch:       "arm64",
			expectedName: "refactorlah_darwin-arm64.tar.gz",
		},
		{
			name:         "linux arm64",
			goos:         "linux",
			goarch:       "arm64",
			expectedName: "refactorlah_linux-arm64.tar.gz",
		},
		{
			name:         "windows amd64",
			goos:         "windows",
			goarch:       "amd64",
			expectedName: "refactorlah_windows-amd64.zip",
		},
		{
			name:          "unsupported target",
			goos:          "linux",
			goarch:        "amd64",
			expectedError: "self-update is not available for linux/amd64",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			name, err := releaseArchiveName(test.goos, test.goarch)
			if test.expectedError != "" {
				if err == nil {
					t.Fatal("expected unsupported target error")
				}
				if !strings.Contains(err.Error(), test.expectedError) {
					t.Fatalf("unexpected error: %v", err)
				}
				return
			}

			if err != nil {
				t.Fatalf("release archive name: %v", err)
			}
			if name != test.expectedName {
				t.Fatalf("unexpected archive name: %s", name)
			}
		})
	}
}

func TestArchiveHelpersHandleReleasePackageFormats(t *testing.T) {
	zipArchive := mustCreateArchiveWithPath(t, "refactorlah_windows-amd64.zip", `refactorlah_windows-amd64\refactorlah.exe`, []byte("windows-binary"))
	zipDestination := filepath.Join(t.TempDir(), "refactorlah.exe")
	if err := extractBinaryFromArchive("refactorlah_windows-amd64.zip", zipArchive, zipDestination, "refactorlah.exe"); err != nil {
		t.Fatalf("extract windows zip: %v", err)
	}
	content, err := os.ReadFile(zipDestination)
	if err != nil {
		t.Fatalf("read extracted windows binary: %v", err)
	}
	if string(content) != "windows-binary" {
		t.Fatalf("unexpected extracted zip content: %q", content)
	}

	checksum, err := expectedChecksum([]byte("ABCDEF  *refactorlah_windows-amd64.zip\n"), "refactorlah_windows-amd64.zip")
	if err != nil {
		t.Fatalf("read checksum entry: %v", err)
	}
	if checksum != "abcdef" {
		t.Fatalf("expected lowercase checksum, got %q", checksum)
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
