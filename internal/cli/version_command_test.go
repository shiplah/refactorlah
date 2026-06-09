package cli

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"

	"github.com/shiplah/refactorlah/internal/buildinfo"
)

func TestVersionCommandShortAndJSON(t *testing.T) {
	restore := setBuildInfoForTest()
	defer restore()

	command := NewVersionCommand()

	var shortStdout bytes.Buffer
	var shortStderr bytes.Buffer
	exitCode := command.Run([]string{"--short"}, &shortStdout, &shortStderr)
	if exitCode != ExitSuccess {
		t.Fatalf("unexpected exit code: %d", exitCode)
	}
	if shortStdout.String() != "v1.2.3\n" {
		t.Fatalf("unexpected short version output: %q", shortStdout.String())
	}
	if shortStderr.Len() != 0 {
		t.Fatalf("expected empty stderr, got %q", shortStderr.String())
	}

	var jsonStdout bytes.Buffer
	var jsonStderr bytes.Buffer
	exitCode = command.Run([]string{"--json"}, &jsonStdout, &jsonStderr)
	if exitCode != ExitSuccess {
		t.Fatalf("unexpected exit code: %d", exitCode)
	}

	var payload map[string]string
	if err := json.Unmarshal(jsonStdout.Bytes(), &payload); err != nil {
		t.Fatalf("decode json output: %v", err)
	}
	if payload["version"] != "v1.2.3" {
		t.Fatalf("unexpected version in json output: %#v", payload)
	}
	if payload["distribution"] != buildinfo.DistributionGitHubRelease {
		t.Fatalf("unexpected distribution in json output: %#v", payload)
	}
	if jsonStderr.Len() != 0 {
		t.Fatalf("expected empty stderr, got %q", jsonStderr.String())
	}
}

func TestRootVersionFlagUsesShortOutput(t *testing.T) {
	restore := setBuildInfoForTest()
	defer restore()

	command := NewRootCommand()

	var stdout bytes.Buffer
	var stderr bytes.Buffer
	exitCode := command.Run(t.Context(), []string{"--version"}, &stdout, &stderr)
	if exitCode != ExitSuccess {
		t.Fatalf("unexpected exit code: %d", exitCode)
	}
	if stdout.String() != "v1.2.3\n" {
		t.Fatalf("unexpected --version output: %q", stdout.String())
	}
	if stderr.Len() != 0 {
		t.Fatalf("expected empty stderr, got %q", stderr.String())
	}
}

func TestVersionHelpShowsUsageWithoutError(t *testing.T) {
	command := NewVersionCommand()

	var stdout bytes.Buffer
	var stderr bytes.Buffer
	exitCode := command.Run([]string{"--help"}, &stdout, &stderr)
	if exitCode != ExitSuccess {
		t.Fatalf("unexpected exit code: %d", exitCode)
	}
	if !strings.Contains(stdout.String(), "refactorlah version [--short|--json]") {
		t.Fatalf("expected version usage, got %q", stdout.String())
	}
	if stderr.Len() != 0 {
		t.Fatalf("expected empty stderr, got %q", stderr.String())
	}
}

func setBuildInfoForTest() func() {
	originalVersion := buildinfo.Version
	originalCommit := buildinfo.Commit
	originalBuildDate := buildinfo.BuildDate
	originalDistribution := buildinfo.Distribution

	buildinfo.Version = "v1.2.3"
	buildinfo.Commit = "abc1234"
	buildinfo.BuildDate = "2026-06-08T10:11:12Z"
	buildinfo.Distribution = buildinfo.DistributionGitHubRelease

	return func() {
		buildinfo.Version = originalVersion
		buildinfo.Commit = originalCommit
		buildinfo.BuildDate = originalBuildDate
		buildinfo.Distribution = originalDistribution
	}
}
