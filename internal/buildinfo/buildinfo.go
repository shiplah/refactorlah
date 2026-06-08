package buildinfo

import (
	"runtime"
	"strings"
)

const (
	DistributionDev           = "dev"
	DistributionGitHubRelease = "github-release"
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
	return Info{
		Version:      fallback(strings.TrimSpace(Version), "dev"),
		Commit:       fallback(strings.TrimSpace(Commit), "unknown"),
		BuildDate:    fallback(strings.TrimSpace(BuildDate), "unknown"),
		Distribution: fallback(strings.TrimSpace(Distribution), DistributionDev),
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
