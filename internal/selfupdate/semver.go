package selfupdate

import (
	"strconv"
	"strings"
)

type semanticVersion struct {
	Major      int
	Minor      int
	Patch      int
	Prerelease string
	Valid      bool
}

func parseSemanticVersion(value string) semanticVersion {
	trimmed := strings.TrimSpace(value)
	trimmed = strings.TrimPrefix(trimmed, "v")
	if trimmed == "" {
		return semanticVersion{}
	}

	prerelease := ""
	if index := strings.IndexByte(trimmed, '-'); index >= 0 {
		prerelease = trimmed[index+1:]
		trimmed = trimmed[:index]
	}

	parts := strings.Split(trimmed, ".")
	if len(parts) != 3 {
		return semanticVersion{}
	}

	major, err := strconv.Atoi(parts[0])
	if err != nil {
		return semanticVersion{}
	}
	minor, err := strconv.Atoi(parts[1])
	if err != nil {
		return semanticVersion{}
	}
	patch, err := strconv.Atoi(parts[2])
	if err != nil {
		return semanticVersion{}
	}

	return semanticVersion{
		Major:      major,
		Minor:      minor,
		Patch:      patch,
		Prerelease: prerelease,
		Valid:      true,
	}
}

func compareSemanticVersions(left string, right string) (int, bool) {
	a := parseSemanticVersion(left)
	b := parseSemanticVersion(right)
	if !a.Valid || !b.Valid {
		return 0, false
	}

	if a.Major != b.Major {
		if a.Major < b.Major {
			return -1, true
		}
		return 1, true
	}
	if a.Minor != b.Minor {
		if a.Minor < b.Minor {
			return -1, true
		}
		return 1, true
	}
	if a.Patch != b.Patch {
		if a.Patch < b.Patch {
			return -1, true
		}
		return 1, true
	}

	if a.Prerelease == b.Prerelease {
		return 0, true
	}
	if a.Prerelease == "" {
		return 1, true
	}
	if b.Prerelease == "" {
		return -1, true
	}
	if a.Prerelease < b.Prerelease {
		return -1, true
	}
	if a.Prerelease > b.Prerelease {
		return 1, true
	}

	return 0, true
}
