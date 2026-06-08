package buildinfo

import (
	"runtime"
	"runtime/debug"
	"strings"
)

const (
	DistributionDev           = "dev"
	DistributionGitHubRelease = "github-release"
	DistributionGoInstall     = "go-install"
	DistributionSourceInstall = "source-install"
)

var (
	Version      = "dev"
	Commit       = "unknown"
	BuildDate    = "unknown"
	Distribution = DistributionDev
)

type Info struct {
	Version      string `json:"version"`
	Commit       string `json:"commit"`
	BuildDate    string `json:"build_date"`
	Distribution string `json:"distribution"`
	GOOS         string `json:"goos"`
	GOARCH       string `json:"goarch"`
}

func Current() Info {
	return currentWithModuleVersion(mainModuleVersion())
}

func currentWithModuleVersion(moduleVersion string) Info {
	version := fallback(strings.TrimSpace(Version), "dev")
	distribution := fallback(strings.TrimSpace(Distribution), DistributionDev)
	if version == "dev" && moduleVersion != "" {
		version = moduleVersion
		if distribution == DistributionDev {
			distribution = DistributionGoInstall
		}
	}

	return Info{
		Version:      version,
		Commit:       fallback(strings.TrimSpace(Commit), "unknown"),
		BuildDate:    fallback(strings.TrimSpace(BuildDate), "unknown"),
		Distribution: distribution,
		GOOS:         runtime.GOOS,
		GOARCH:       runtime.GOARCH,
	}
}

func (i Info) Target() string {
	return i.GOOS + "/" + i.GOARCH
}

func fallback(value string, fallbackValue string) string {
	if value == "" {
		return fallbackValue
	}

	return value
}

func mainModuleVersion() string {
	build, ok := debug.ReadBuildInfo()
	if !ok {
		return ""
	}
	if build.Main.Version == "" || build.Main.Version == "(devel)" {
		return ""
	}

	return build.Main.Version
}
