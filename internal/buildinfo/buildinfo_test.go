package buildinfo

import "testing"

func TestCurrentWithModuleVersionDetectsGoInstall(t *testing.T) {
	withBuildInfoDefaults(t)

	info := currentWithModuleVersion("v1.2.3")

	if info.Version != "v1.2.3" {
		t.Fatalf("unexpected version: %#v", info)
	}
	if info.Distribution != DistributionGoInstall {
		t.Fatalf("unexpected distribution: %#v", info)
	}
}

func TestCurrentWithModuleVersionKeepsSourceInstallDistribution(t *testing.T) {
	withBuildInfoDefaults(t)
	Distribution = DistributionSourceInstall

	info := currentWithModuleVersion("v1.2.3")

	if info.Version != "v1.2.3" {
		t.Fatalf("unexpected version: %#v", info)
	}
	if info.Distribution != DistributionSourceInstall {
		t.Fatalf("unexpected distribution: %#v", info)
	}
}

func TestCurrentWithModuleVersionPrefersInjectedReleaseVersion(t *testing.T) {
	withBuildInfoDefaults(t)
	Version = "v2.0.0"
	Distribution = DistributionGitHubRelease

	info := currentWithModuleVersion("v1.2.3")

	if info.Version != "v2.0.0" {
		t.Fatalf("unexpected version: %#v", info)
	}
	if info.Distribution != DistributionGitHubRelease {
		t.Fatalf("unexpected distribution: %#v", info)
	}
}

func TestCurrentWithModuleVersionUsesDefaultsForBlankInjectedValues(t *testing.T) {
	withBuildInfoDefaults(t)
	Version = " "
	Commit = " "
	BuildDate = " "
	Distribution = " "

	info := currentWithModuleVersion("")

	if info.Version != "dev" {
		t.Fatalf("unexpected version: %#v", info)
	}
	if info.Commit != "unknown" {
		t.Fatalf("unexpected commit: %#v", info)
	}
	if info.BuildDate != "unknown" {
		t.Fatalf("unexpected build date: %#v", info)
	}
	if info.Distribution != DistributionDev {
		t.Fatalf("unexpected distribution: %#v", info)
	}
}

func TestInfoTargetUsesRuntimeTuple(t *testing.T) {
	info := Info{GOOS: "darwin", GOARCH: "arm64"}

	if info.Target() != "darwin/arm64" {
		t.Fatalf("unexpected target: %s", info.Target())
	}
}

func withBuildInfoDefaults(t *testing.T) {
	t.Helper()

	originalVersion := Version
	originalCommit := Commit
	originalBuildDate := BuildDate
	originalDistribution := Distribution

	Version = "dev"
	Commit = "unknown"
	BuildDate = "unknown"
	Distribution = DistributionDev

	t.Cleanup(func() {
		Version = originalVersion
		Commit = originalCommit
		BuildDate = originalBuildDate
		Distribution = originalDistribution
	})
}
