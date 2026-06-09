package selfupdate

import (
	"bytes"
	"context"
	"crypto/sha256"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/shiplah/refactorlah/internal/buildinfo"
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
	if !result.SelfUpdateSupported {
		t.Fatalf("expected GitHub release build to support self-update, got %#v", result)
	}
	if result.TargetVersion != "v1.1.0" {
		t.Fatalf("unexpected target version: %#v", result)
	}
}

func TestNewUpdaterInitialisesRuntimeDefaults(t *testing.T) {
	updater, err := NewUpdater()
	if err != nil {
		t.Fatalf("create updater: %v", err)
	}

	if updater.Executable == "" {
		t.Fatal("expected executable path")
	}
	if updater.Locator == nil {
		t.Fatal("expected release locator")
	}
	if updater.Downloader == nil {
		t.Fatal("expected release downloader")
	}
	if updater.Stdout == nil || updater.Stderr == nil {
		t.Fatal("expected output writers")
	}
	if updater.BuildInfo.GOOS == "" || updater.BuildInfo.GOARCH == "" {
		t.Fatalf("expected runtime build target, got %#v", updater.BuildInfo)
	}
}

func TestUpdaterCheckExplainsUnsupportedInstallsWithoutReleaseLookup(t *testing.T) {
	tests := []struct {
		name                string
		distribution        string
		expectedInstruction string
	}{
		{
			name:                "source install",
			distribution:        buildinfo.DistributionSourceInstall,
			expectedInstruction: "bin/install.sh",
		},
		{
			name:                "go install",
			distribution:        buildinfo.DistributionGoInstall,
			expectedInstruction: "go install github.com/shiplah/refactorlah/cmd/refactorlah@latest",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			updater := &Updater{
				BuildInfo: buildinfo.Info{
					Version:      "v1.0.0",
					Distribution: test.distribution,
					GOOS:         "linux",
					GOARCH:       "amd64",
				},
				Executable: "/tmp/refactorlah",
				Locator:    failingReleaseLocator{},
			}

			result, err := updater.Check(t.Context(), CheckOptions{})
			if err != nil {
				t.Fatalf("check unsupported install for updates: %v", err)
			}
			if result.SelfUpdateSupported {
				t.Fatalf("did not expect unsupported install to support self-update: %#v", result)
			}
			if result.UpdateAvailable || result.TargetVersion != "" {
				t.Fatalf("did not expect unsupported install to report release metadata, got %#v", result)
			}
			if !strings.Contains(result.UpdateInstructions, test.expectedInstruction) {
				t.Fatalf("expected manual update instructions, got %#v", result)
			}
		})
	}
}

func TestUpdaterPlanRequiresReleaseAssetsForSupportedInstalls(t *testing.T) {
	tests := []struct {
		name          string
		assets        []Asset
		expectedError string
	}{
		{
			name: "missing archive",
			assets: []Asset{
				{Name: checksumAssetName, BrowserDownloadURL: "https://example.test/checksums"},
			},
			expectedError: "release v1.1.0 does not contain refactorlah_darwin-arm64.tar.gz",
		},
		{
			name: "missing checksums",
			assets: []Asset{
				{Name: "refactorlah_darwin-arm64.tar.gz", BrowserDownloadURL: "https://example.test/archive"},
			},
			expectedError: "release v1.1.0 does not contain refactorlah_checksums.txt",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
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
						Assets:  test.assets,
					},
				},
			}

			_, err := updater.Plan(t.Context(), CheckOptions{})
			if err == nil {
				t.Fatal("expected missing asset error")
			}
			if !strings.Contains(err.Error(), test.expectedError) {
				t.Fatalf("unexpected error: %v", err)
			}
		})
	}
}

func TestUpdaterClassifiesVersionStates(t *testing.T) {
	tests := []struct {
		name            string
		currentVersion  string
		targetVersion   string
		explicitTarget  bool
		updateAvailable bool
		upToDate        bool
		downgrade       bool
	}{
		{
			name:            "latest newer semantic release",
			currentVersion:  "v1.0.0",
			targetVersion:   "v1.1.0",
			updateAvailable: true,
		},
		{
			name:           "latest same semantic release",
			currentVersion: "v1.0.0",
			targetVersion:  "v1.0.0",
			upToDate:       true,
		},
		{
			name:           "latest older semantic release",
			currentVersion: "v1.1.0",
			targetVersion:  "v1.0.0",
			upToDate:       true,
			downgrade:      true,
		},
		{
			name:            "latest non semantic release falls back to exact comparison",
			currentVersion:  "snapshot-a",
			targetVersion:   "snapshot-b",
			updateAvailable: true,
		},
		{
			name:            "explicit newer release",
			currentVersion:  "v1.0.0",
			targetVersion:   "v1.1.0",
			explicitTarget:  true,
			updateAvailable: true,
		},
		{
			name:           "explicit same release",
			currentVersion: "v1.0.0",
			targetVersion:  "v1.0.0",
			explicitTarget: true,
			upToDate:       true,
		},
		{
			name:            "explicit older release is allowed but marked as downgrade",
			currentVersion:  "v1.1.0",
			targetVersion:   "v1.0.0",
			explicitTarget:  true,
			updateAvailable: true,
			downgrade:       true,
		},
		{
			name:            "stable release is newer than prerelease",
			currentVersion:  "v1.0.0-rc.1",
			targetVersion:   "v1.0.0",
			explicitTarget:  true,
			updateAvailable: true,
		},
		{
			name:            "explicit prerelease upgrade compares numeric identifiers",
			currentVersion:  "v1.0.0-rc.2",
			targetVersion:   "v1.0.0-rc.10",
			explicitTarget:  true,
			updateAvailable: true,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			result := classifyVersionState(CheckResult{
				CurrentVersion: test.currentVersion,
				TargetVersion:  test.targetVersion,
			}, test.explicitTarget)

			if result.UpdateAvailable != test.updateAvailable {
				t.Fatalf("unexpected update_available for %#v", result)
			}
			if result.UpToDate != test.upToDate {
				t.Fatalf("unexpected up_to_date for %#v", result)
			}
			if result.Downgrade != test.downgrade {
				t.Fatalf("unexpected downgrade for %#v", result)
			}
		})
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

func TestUpdaterApplyPlanStagesReplacementOnWindows(t *testing.T) {
	archiveName := "refactorlah_windows-amd64.zip"
	newBinary := []byte("new-windows-binary-content")
	archiveContent := mustCreateArchive(t, archiveName, binaryName("windows"), newBinary)
	checksumContent := []byte(fmt.Sprintf("%x  %s\n", sha256Bytes(archiveContent), archiveName))

	tempDir := t.TempDir()
	executablePath := filepath.Join(tempDir, binaryName("windows"))
	if err := os.WriteFile(executablePath, []byte("old-windows-binary-content"), 0o755); err != nil {
		t.Fatalf("write current executable: %v", err)
	}

	var stagedTempDir string
	var stagedSourcePath string
	updater := &Updater{
		BuildInfo: buildinfo.Info{
			Version:      "v1.0.0",
			Distribution: buildinfo.DistributionGitHubRelease,
			GOOS:         "windows",
			GOARCH:       "amd64",
		},
		Executable: executablePath,
		Downloader: fakeDownloader{
			assets: map[string][]byte{
				"https://example.test/archive":   archiveContent,
				"https://example.test/checksums": checksumContent,
			},
		},
		Stdout:      io.Discard,
		Stderr:      io.Discard,
		runtimeGOOS: "windows",
		windowsReplacementLauncher: func(tempDir string, sourcePath string) error {
			stagedTempDir = tempDir
			stagedSourcePath = sourcePath

			if _, err := os.Stat(tempDir); err != nil {
				return fmt.Errorf("stat staged temp dir: %w", err)
			}

			content, err := os.ReadFile(sourcePath)
			if err != nil {
				return fmt.Errorf("read staged replacement: %w", err)
			}
			if !bytes.Equal(content, newBinary) {
				return fmt.Errorf("unexpected staged replacement content: %q", string(content))
			}

			return nil
		},
	}

	result, err := updater.ApplyPlan(t.Context(), UpdatePlan{
		CheckResult: CheckResult{
			CurrentVersion:      "v1.0.0",
			CurrentDistribution: buildinfo.DistributionGitHubRelease,
			TargetVersion:       "v1.1.0",
			UpdateAvailable:     true,
			SelfUpdateSupported: true,
			AssetName:           archiveName,
		},
		archiveAsset:  Asset{Name: archiveName, BrowserDownloadURL: "https://example.test/archive"},
		checksumAsset: Asset{Name: checksumAssetName, BrowserDownloadURL: "https://example.test/checksums"},
	})
	if err != nil {
		t.Fatalf("apply windows update plan: %v", err)
	}
	if !result.Updated || !result.Staged || !result.RestartRequired {
		t.Fatalf("expected staged Windows update result, got %#v", result)
	}
	if stagedTempDir == "" || stagedSourcePath == "" {
		t.Fatalf("expected Windows replacement launcher to be called, temp=%q source=%q", stagedTempDir, stagedSourcePath)
	}
	t.Cleanup(func() { _ = os.RemoveAll(stagedTempDir) })

	if relativePath, err := filepath.Rel(stagedTempDir, stagedSourcePath); err != nil || relativePath != binaryName("windows") {
		t.Fatalf("expected staged source in update workspace, got relative path %q with error %v", relativePath, err)
	}
	if _, err := os.Stat(stagedTempDir); err != nil {
		t.Fatalf("expected staged temp dir to remain for helper cleanup: %v", err)
	}
}

func TestUpdaterApplyPlanRejectsUnsafeReleaseContent(t *testing.T) {
	archiveName := "refactorlah_darwin-arm64.tar.gz"
	validArchive := mustCreateArchive(t, archiveName, "refactorlah", []byte("new-binary-content"))
	missingBinaryArchive := mustCreateArchiveWithPath(t, archiveName, "refactorlah_darwin-arm64/other", []byte("unexpected"))
	missingBinaryChecksum := []byte(fmt.Sprintf("%x  %s\n", sha256Bytes(missingBinaryArchive), archiveName))

	tests := []struct {
		name          string
		archive       []byte
		checksum      []byte
		expectedError string
	}{
		{
			name:          "missing checksum entry",
			archive:       validArchive,
			checksum:      []byte(fmt.Sprintf("%x  other-asset.tar.gz\n", sha256Bytes(validArchive))),
			expectedError: "checksum for refactorlah_darwin-arm64.tar.gz not found",
		},
		{
			name:          "checksum mismatch",
			archive:       validArchive,
			checksum:      []byte(strings.Repeat("0", sha256.Size*2) + "  refactorlah_darwin-arm64.tar.gz\n"),
			expectedError: "checksum mismatch",
		},
		{
			name:          "archive missing expected binary",
			archive:       missingBinaryArchive,
			checksum:      missingBinaryChecksum,
			expectedError: "binary refactorlah_darwin-arm64/refactorlah not found",
		},
		{
			name:          "corrupt archive",
			archive:       []byte("not an archive"),
			checksum:      []byte(fmt.Sprintf("%x  %s\n", sha256Bytes([]byte("not an archive")), archiveName)),
			expectedError: "open gzip archive",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			updater := &Updater{
				BuildInfo: buildinfo.Info{
					Distribution: buildinfo.DistributionGitHubRelease,
					GOOS:         "darwin",
					GOARCH:       "arm64",
				},
				Downloader: fakeDownloader{
					assets: map[string][]byte{
						"https://example.test/archive":   test.archive,
						"https://example.test/checksums": test.checksum,
					},
				},
			}

			_, err := updater.ApplyPlan(t.Context(), UpdatePlan{
				CheckResult: CheckResult{
					UpdateAvailable:     true,
					SelfUpdateSupported: true,
					TargetVersion:       "v1.1.0",
				},
				archiveAsset:  Asset{Name: archiveName, BrowserDownloadURL: "https://example.test/archive"},
				checksumAsset: Asset{Name: checksumAssetName, BrowserDownloadURL: "https://example.test/checksums"},
			})
			if err == nil {
				t.Fatal("expected unsafe release content error")
			}
			if !strings.Contains(err.Error(), test.expectedError) {
				t.Fatalf("unexpected error: %v", err)
			}
		})
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

func TestUpdaterApplyPlanDoesNotRequireDownloaderForUnsupportedInstalls(t *testing.T) {
	updater := &Updater{}
	result, err := updater.ApplyPlan(t.Context(), UpdatePlan{
		CheckResult: CheckResult{
			CurrentVersion:      "v1.0.0",
			CurrentDistribution: buildinfo.DistributionGoInstall,
			TargetVersion:       "v1.1.0",
			UpdateAvailable:     true,
			SelfUpdateSupported: false,
		},
	})
	if err != nil {
		t.Fatalf("apply unsupported update plan: %v", err)
	}
	if result.Updated {
		t.Fatalf("did not expect unsupported install to replace executable: %#v", result)
	}
}

func TestUpdaterReleaseDownloaderFallsBackToGitHubClientLocator(t *testing.T) {
	client := &GitHubClient{}
	updater := &Updater{Locator: client}

	downloader, err := updater.releaseDownloader()
	if err != nil {
		t.Fatalf("resolve downloader: %v", err)
	}
	if downloader != client {
		t.Fatalf("expected GitHub client downloader, got %#v", downloader)
	}
}

func TestUpdaterReleaseDownloaderRequiresDownloadCapableLocator(t *testing.T) {
	updater := &Updater{Locator: fakeReleaseLocator{}}

	_, err := updater.releaseDownloader()
	if err == nil {
		t.Fatal("expected missing downloader error")
	}
	if !strings.Contains(err.Error(), "missing release downloader") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestUpdaterLookupReleaseUsesLatestOrExplicitTag(t *testing.T) {
	locator := &countingReleaseLocator{
		release: Release{TagName: "v1.1.0"},
	}
	updater := &Updater{Locator: locator}

	if _, err := updater.lookupRelease(t.Context(), ""); err != nil {
		t.Fatalf("lookup latest release: %v", err)
	}
	if locator.latestCalls != 1 || locator.tagCalls != 0 {
		t.Fatalf("expected latest lookup only, got latest=%d tag=%d", locator.latestCalls, locator.tagCalls)
	}

	if _, err := updater.lookupRelease(t.Context(), "v1.0.0"); err != nil {
		t.Fatalf("lookup tagged release: %v", err)
	}
	if locator.latestCalls != 1 || locator.tagCalls != 1 || locator.lastTag != "v1.0.0" {
		t.Fatalf("expected explicit tag lookup, got latest=%d tag=%d last=%q", locator.latestCalls, locator.tagCalls, locator.lastTag)
	}
}

func TestUpdaterLookupReleaseRequiresLocator(t *testing.T) {
	updater := &Updater{}

	_, err := updater.lookupRelease(t.Context(), "")
	if err == nil {
		t.Fatal("expected missing locator error")
	}
	if !strings.Contains(err.Error(), "missing release locator") {
		t.Fatalf("unexpected error: %v", err)
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
	if _, err := os.Stat(sourcePath); !os.IsNotExist(err) {
		t.Fatalf("expected replacement source to be removed, got %v", err)
	}
}

func TestReplacementHelperRejectsMissingRequiredFlags(t *testing.T) {
	var stderr bytes.Buffer
	helper := &ReplacementHelper{}

	exitCode := helper.Run(context.Background(), []string{"--target", "/tmp/refactorlah"}, &stderr)
	if exitCode == 0 {
		t.Fatal("expected replacement helper to reject missing source")
	}
	if !strings.Contains(stderr.String(), "requires --target and --source") {
		t.Fatalf("unexpected stderr: %q", stderr.String())
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

type failingReleaseLocator struct{}

func (failingReleaseLocator) Latest(_ context.Context) (Release, error) {
	return Release{}, errors.New("release lookup should not be called")
}

func (failingReleaseLocator) ByTag(_ context.Context, _ string) (Release, error) {
	return Release{}, errors.New("release lookup should not be called")
}

type countingReleaseLocator struct {
	release     Release
	latestCalls int
	tagCalls    int
	lastTag     string
}

func (l *countingReleaseLocator) Latest(_ context.Context) (Release, error) {
	l.latestCalls++
	return l.release, nil
}

func (l *countingReleaseLocator) ByTag(_ context.Context, tag string) (Release, error) {
	l.tagCalls++
	l.lastTag = tag
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
