package selfupdate

import (
	"archive/tar"
	"archive/zip"
	"bytes"
	"compress/gzip"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

func releaseArchiveName(goos string, goarch string) (string, error) {
	slug, err := targetSlug(goos, goarch)
	if err != nil {
		return "", err
	}

	extension := "tar.gz"
	if goos == "windows" {
		extension = "zip"
	}

	return fmt.Sprintf("refactorlah_%s.%s", slug, extension), nil
}

func targetSlug(goos string, goarch string) (string, error) {
	switch goos + "/" + goarch {
	case "darwin/arm64", "linux/arm64", "windows/amd64":
		return goos + "-" + goarch, nil
	default:
		return "", fmt.Errorf("self-update is not available for %s/%s", goos, goarch)
	}
}

func binaryName(goos string) string {
	if goos == "windows" {
		return "refactorlah.exe"
	}

	return "refactorlah"
}

func expectedChecksum(checksumFile []byte, assetName string) (string, error) {
	lines := strings.Split(string(checksumFile), "\n")
	for _, line := range lines {
		fields := strings.Fields(line)
		if len(fields) < 2 {
			continue
		}

		name := strings.TrimPrefix(fields[1], "*")
		if name == assetName {
			return strings.ToLower(fields[0]), nil
		}
	}

	return "", fmt.Errorf("checksum for %s not found", assetName)
}

func verifyChecksum(content []byte, expected string) error {
	digest := sha256.Sum256(content)
	actual := hex.EncodeToString(digest[:])
	if strings.ToLower(expected) != actual {
		return fmt.Errorf("checksum mismatch: expected %s, got %s", expected, actual)
	}

	return nil
}

func extractBinaryFromArchive(archiveName string, archiveContent []byte, destinationPath string, expectedBinaryName string) error {
	switch {
	case strings.HasSuffix(archiveName, ".zip"):
		return extractBinaryFromZip(archiveContent, destinationPath, expectedBinaryName)
	case strings.HasSuffix(archiveName, ".tar.gz"):
		return extractBinaryFromTarGz(archiveContent, destinationPath, expectedBinaryName)
	default:
		return fmt.Errorf("unsupported archive format for %s", archiveName)
	}
}

func extractBinaryFromZip(archiveContent []byte, destinationPath string, expectedBinaryName string) error {
	reader, err := zip.NewReader(bytes.NewReader(archiveContent), int64(len(archiveContent)))
	if err != nil {
		return fmt.Errorf("open zip archive: %w", err)
	}

	for _, file := range reader.File {
		if filepath.Base(file.Name) != expectedBinaryName {
			continue
		}

		stream, err := file.Open()
		if err != nil {
			return fmt.Errorf("open %s in zip archive: %w", file.Name, err)
		}
		defer stream.Close()

		return writeExecutableFile(destinationPath, stream, 0o755)
	}

	return fmt.Errorf("binary %s not found in zip archive", expectedBinaryName)
}

func extractBinaryFromTarGz(archiveContent []byte, destinationPath string, expectedBinaryName string) error {
	gzipReader, err := gzip.NewReader(bytes.NewReader(archiveContent))
	if err != nil {
		return fmt.Errorf("open gzip archive: %w", err)
	}
	defer gzipReader.Close()

	reader := tar.NewReader(gzipReader)
	for {
		header, err := reader.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return fmt.Errorf("read tar archive: %w", err)
		}
		if header.Typeflag != tar.TypeReg {
			continue
		}
		if filepath.Base(header.Name) != expectedBinaryName {
			continue
		}

		mode := os.FileMode(0o755)
		if header.FileInfo().Mode().Perm() != 0 {
			mode = header.FileInfo().Mode().Perm()
		}

		return writeExecutableFile(destinationPath, reader, mode)
	}

	return fmt.Errorf("binary %s not found in tar archive", expectedBinaryName)
}

func writeExecutableFile(destinationPath string, source io.Reader, mode os.FileMode) error {
	file, err := os.OpenFile(destinationPath, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, mode)
	if err != nil {
		return fmt.Errorf("create %s: %w", destinationPath, err)
	}

	if _, err := io.Copy(file, source); err != nil {
		_ = file.Close()
		return fmt.Errorf("write %s: %w", destinationPath, err)
	}
	if err := file.Close(); err != nil {
		return fmt.Errorf("close %s: %w", destinationPath, err)
	}
	if err := os.Chmod(destinationPath, mode); err != nil && os.PathSeparator == '/' {
		return fmt.Errorf("chmod %s: %w", destinationPath, err)
	}

	return nil
}
