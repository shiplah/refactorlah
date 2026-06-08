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
