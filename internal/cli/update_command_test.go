package cli

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/NickSdot/refactorlah/internal/buildinfo"
	"github.com/NickSdot/refactorlah/internal/selfupdate"
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
	if payload["self_update_supported"] != true {
		t.Fatalf("expected self_update_supported=true, got %#v", payload)
	}
	if payload["target_version"] != "v1.1.0" {
		t.Fatalf("unexpected target version: %#v", payload)
	}
	if stderr.Len() != 0 {
		t.Fatalf("expected empty stderr, got %q", stderr.String())
	}
}

func TestUpdateHelpShowsUsageWithoutCreatingUpdater(t *testing.T) {
	command := NewUpdateCommand()
	command.newUpdater = func() (*selfupdate.Updater, error) {
		return nil, errors.New("updater should not be created for help")
	}

	var stdout bytes.Buffer
	var stderr bytes.Buffer
	exitCode := command.Run(t.Context(), []string{"--help"}, &stdout, &stderr)
	if exitCode != ExitSuccess {
		t.Fatalf("unexpected exit code: %d", exitCode)
	}
	if !strings.Contains(stdout.String(), "refactorlah update [--check] [--yes] [--to TAG] [--json]") {
		t.Fatalf("expected update usage, got %q", stdout.String())
	}
	if stderr.Len() != 0 {
		t.Fatalf("expected empty stderr, got %q", stderr.String())
	}
}

func TestUpdateCommandCheckHumanOutputReportsAvailableRelease(t *testing.T) {
	command := newUpdateCommandForTest(t, "darwin", "arm64")

	var stdout bytes.Buffer
	var stderr bytes.Buffer
	exitCode := command.Run(t.Context(), []string{"--check"}, &stdout, &stderr)
	if exitCode != ExitSuccess {
		t.Fatalf("unexpected exit code: %d", exitCode)
	}

	output := stdout.String()
	for _, expected := range []string{
		"Update available: v1.0.0 -> v1.1.0",
		"Asset: refactorlah_darwin-arm64.tar.gz",
		"Release: https://example.test/releases/v1.1.0",
	} {
		if !strings.Contains(output, expected) {
			t.Fatalf("expected %q in update check output, got %q", expected, output)
		}
	}
	if stderr.Len() != 0 {
		t.Fatalf("expected empty stderr, got %q", stderr.String())
	}
}

func TestUpdateCommandCheckHumanOutputReportsUpToDate(t *testing.T) {
	command := newUpdateCommandForRelease(t, "darwin", "arm64", buildinfo.DistributionGitHubRelease, "v1.0.0")

	var stdout bytes.Buffer
	var stderr bytes.Buffer
	exitCode := command.Run(t.Context(), []string{"--check"}, &stdout, &stderr)
	if exitCode != ExitSuccess {
		t.Fatalf("unexpected exit code: %d", exitCode)
	}
	if stdout.String() != "refactorlah is up to date (v1.0.0)\n" {
		t.Fatalf("unexpected up-to-date output: %q", stdout.String())
	}
	if stderr.Len() != 0 {
		t.Fatalf("expected empty stderr, got %q", stderr.String())
	}
}

func TestUpdateCommandCancelLeavesExecutableUntouched(t *testing.T) {
	command := newUpdateCommandForTest(t, "darwin", "arm64")
	command.stdin = strings.NewReader("n\n")

	executablePath := updateCommandExecutable(t, command)
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

func TestConfirmUpdateAcceptsOnlyExplicitYes(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected bool
	}{
		{name: "yes", input: "yes\n", expected: true},
		{name: "short yes", input: "Y\n", expected: true},
		{name: "no", input: "n\n", expected: false},
		{name: "empty input", input: "", expected: false},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			var stdout bytes.Buffer
			confirmed := confirmUpdate(strings.NewReader(test.input), &stdout, selfupdate.CheckResult{
				TargetVersion:  "v1.1.0",
				ExecutablePath: "/bin/refactorlah",
			})
			if confirmed != test.expected {
				t.Fatalf("unexpected confirmation result: %v", confirmed)
			}
			if !strings.Contains(stdout.String(), "Install v1.1.0 at /bin/refactorlah? [y/N]:") {
				t.Fatalf("unexpected prompt: %q", stdout.String())
			}
		})
	}
}

func TestUpdatePromptExplainsDowngradeTarget(t *testing.T) {
	prompt := updatePrompt(selfupdate.CheckResult{
		CurrentVersion: "v1.1.0",
		TargetVersion:  "v1.0.0",
		ExecutablePath: "/bin/refactorlah",
		Downgrade:      true,
	})

	expected := "Install v1.0.0 over current v1.1.0 at /bin/refactorlah?"
	if prompt != expected {
		t.Fatalf("unexpected downgrade prompt:\nwant %q\n got %q", expected, prompt)
	}
}

func TestUpdateCommandCheckUnsupportedInstallsExplainsManualRefresh(t *testing.T) {
	tests := []struct {
		name                 string
		distribution         string
		expectedInstruction  string
		expectedDistribution string
	}{
		{
			name:                 "source install",
			distribution:         buildinfo.DistributionSourceInstall,
			expectedInstruction:  "bin/install.sh",
			expectedDistribution: "source-install",
		},
		{
			name:                 "go install",
			distribution:         buildinfo.DistributionGoInstall,
			expectedInstruction:  "go install github.com/NickSdot/refactorlah/cmd/refactorlah@latest",
			expectedDistribution: "go-install",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			command := newUpdateCommandForDistribution(t, "linux", "amd64", test.distribution)

			var stdout bytes.Buffer
			var stderr bytes.Buffer
			exitCode := command.Run(t.Context(), []string{"--check"}, &stdout, &stderr)
			if exitCode != ExitSuccess {
				t.Fatalf("unexpected exit code: %d", exitCode)
			}

			output := stdout.String()
			if !strings.Contains(output, "Self-update is only available for GitHub release binaries.") {
				t.Fatalf("expected self-update explanation, got %q", output)
			}
			if !strings.Contains(output, test.expectedInstruction) {
				t.Fatalf("expected refresh instruction, got %q", output)
			}
			if !strings.Contains(output, test.expectedDistribution) {
				t.Fatalf("expected current distribution, got %q", output)
			}
			if strings.Contains(output, "Published release") {
				t.Fatalf("did not expect unsupported install to query release metadata, got %q", output)
			}
			if stderr.Len() != 0 {
				t.Fatalf("expected empty stderr, got %q", stderr.String())
			}
		})
	}
}

func TestUpdateCommandUnsupportedInstallsDoNotSelfUpdate(t *testing.T) {
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
			expectedInstruction: "go install github.com/NickSdot/refactorlah/cmd/refactorlah@latest",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			command := newUpdateCommandForDistribution(t, "linux", "amd64", test.distribution)

			var stdout bytes.Buffer
			var stderr bytes.Buffer
			exitCode := command.Run(t.Context(), nil, &stdout, &stderr)
			if exitCode != ExitGeneralFailure {
				t.Fatalf("unexpected exit code: %d", exitCode)
			}

			output := stdout.String()
			if !strings.Contains(output, "Self-update is only available for GitHub release binaries.") {
				t.Fatalf("expected self-update explanation, got %q", output)
			}
			if !strings.Contains(output, test.expectedInstruction) {
				t.Fatalf("expected refresh instruction, got %q", output)
			}
			if stderr.Len() != 0 {
				t.Fatalf("expected empty stderr, got %q", stderr.String())
			}
		})
	}
}

func TestUpdateCommandCheckExplicitOlderReleaseReportsDowngrade(t *testing.T) {
	command := newUpdateCommandForRelease(t, "darwin", "arm64", buildinfo.DistributionGitHubRelease, "v0.9.0")

	var stdout bytes.Buffer
	var stderr bytes.Buffer
	exitCode := command.Run(t.Context(), []string{"--check", "--to", "v0.9.0"}, &stdout, &stderr)
	if exitCode != ExitSuccess {
		t.Fatalf("unexpected exit code: %d", exitCode)
	}

	output := stdout.String()
	if !strings.Contains(output, "Current version v1.0.0 is newer than published release v0.9.0") {
		t.Fatalf("expected downgrade check output, got %q", output)
	}
	if strings.Contains(output, "Update available") {
		t.Fatalf("did not expect downgrade check to look like a newer update: %q", output)
	}
	if stderr.Len() != 0 {
		t.Fatalf("expected empty stderr, got %q", stderr.String())
	}
}

func TestRenderUpdateApplyResultHumanStates(t *testing.T) {
	tests := []struct {
		name     string
		result   selfupdate.ApplyResult
		expected []string
	}{
		{
			name: "staged",
			result: selfupdate.ApplyResult{
				CheckResult:     selfupdate.CheckResult{TargetVersion: "v1.1.0"},
				Staged:          true,
				RestartRequired: true,
			},
			expected: []string{
				"Update to v1.1.0 staged.",
				"Restart refactorlah to use the new version.",
			},
		},
		{
			name: "updated",
			result: selfupdate.ApplyResult{
				CheckResult: selfupdate.CheckResult{TargetVersion: "v1.1.0"},
				Updated:     true,
			},
			expected: []string{"Updated refactorlah to v1.1.0."},
		},
		{
			name: "downgrade",
			result: selfupdate.ApplyResult{
				CheckResult: selfupdate.CheckResult{
					CurrentVersion: "v1.1.0",
					TargetVersion:  "v1.0.0",
					Downgrade:      true,
				},
			},
			expected: []string{"Current version v1.1.0 is newer than published release v1.0.0"},
		},
		{
			name: "up to date",
			result: selfupdate.ApplyResult{
				CheckResult: selfupdate.CheckResult{
					CurrentVersion: "v1.0.0",
					TargetVersion:  "v1.0.0",
					UpToDate:       true,
				},
			},
			expected: []string{"refactorlah is up to date (v1.0.0)"},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			var stdout bytes.Buffer
			var stderr bytes.Buffer
			exitCode := renderUpdateApplyResult(&stdout, &stderr, test.result, false)
			if exitCode != ExitSuccess {
				t.Fatalf("unexpected exit code: %d", exitCode)
			}
			for _, expected := range test.expected {
				if !strings.Contains(stdout.String(), expected) {
					t.Fatalf("expected %q in apply output, got %q", expected, stdout.String())
				}
			}
			if stderr.Len() != 0 {
				t.Fatalf("expected empty stderr, got %q", stderr.String())
			}
		})
	}
}

func TestUpdateCommandRejectsInteractiveJSONApply(t *testing.T) {
	command := newUpdateCommandForTest(t, "darwin", "arm64")

	var stdout bytes.Buffer
	var stderr bytes.Buffer
	exitCode := command.Run(t.Context(), []string{"--json"}, &stdout, &stderr)
	if exitCode != ExitInvalidArguments {
		t.Fatalf("unexpected exit code: %d", exitCode)
	}
	if !strings.Contains(stderr.String(), "--json requires --check or --yes") {
		t.Fatalf("expected json usage error, got %q", stderr.String())
	}
	if stdout.Len() != 0 {
		t.Fatalf("expected empty stdout, got %q", stdout.String())
	}
}

func TestUpdateCommandRejectsPositionalArguments(t *testing.T) {
	command := newUpdateCommandForTest(t, "darwin", "arm64")

	var stdout bytes.Buffer
	var stderr bytes.Buffer
	exitCode := command.Run(t.Context(), []string{"unexpected"}, &stdout, &stderr)
	if exitCode != ExitInvalidArguments {
		t.Fatalf("unexpected exit code: %d", exitCode)
	}
	if !strings.Contains(stderr.String(), "update does not accept positional arguments") {
		t.Fatalf("expected positional argument error, got %q", stderr.String())
	}
	if stdout.Len() != 0 {
		t.Fatalf("expected empty stdout, got %q", stdout.String())
	}
}

func newUpdateCommandForTest(t *testing.T, goos string, goarch string) *UpdateCommand {
	t.Helper()
	return newUpdateCommandForDistribution(t, goos, goarch, buildinfo.DistributionGitHubRelease)
}

func newUpdateCommandForDistribution(t *testing.T, goos string, goarch string, distribution string) *UpdateCommand {
	t.Helper()
	return newUpdateCommandForRelease(t, goos, goarch, distribution, "v1.1.0")
}

func newUpdateCommandForRelease(t *testing.T, goos string, goarch string, distribution string, releaseTag string) *UpdateCommand {
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
				Distribution: distribution,
				GOOS:         goos,
				GOARCH:       goarch,
			},
			Executable: executablePath,
			Locator: staticReleaseLocator{
				release: selfupdate.Release{
					TagName: releaseTag,
					HTMLURL: "https://example.test/releases/" + releaseTag,
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

func updateCommandExecutable(t *testing.T, command *UpdateCommand) string {
	t.Helper()

	updater, err := command.newUpdater()
	if err != nil {
		t.Fatalf("create updater fixture: %v", err)
	}

	return updater.Executable
}

type staticReleaseLocator struct {
	release selfupdate.Release
}

func (l staticReleaseLocator) Latest(_ context.Context) (selfupdate.Release, error) {
	return l.release, nil
}

func (l staticReleaseLocator) ByTag(_ context.Context, tag string) (selfupdate.Release, error) {
	if tag != l.release.TagName {
		return selfupdate.Release{}, fmt.Errorf("unexpected release tag %q", tag)
	}

	return l.release, nil
}

type staticDownloader struct{}

func (staticDownloader) Download(_ context.Context, _ string) ([]byte, error) {
	return nil, errors.New("unexpected download in update command test")
}
