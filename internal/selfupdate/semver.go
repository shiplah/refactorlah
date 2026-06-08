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

	if index := strings.IndexByte(trimmed, '+'); index >= 0 {
		trimmed = trimmed[:index]
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

	return comparePrerelease(a.Prerelease, b.Prerelease), true
}

func comparePrerelease(left string, right string) int {
	leftIdentifiers := strings.Split(left, ".")
	rightIdentifiers := strings.Split(right, ".")

	limit := len(leftIdentifiers)
	if len(rightIdentifiers) < limit {
		limit = len(rightIdentifiers)
	}

	for index := 0; index < limit; index++ {
		if compare := comparePrereleaseIdentifier(leftIdentifiers[index], rightIdentifiers[index]); compare != 0 {
			return compare
		}
	}

	switch {
	case len(leftIdentifiers) < len(rightIdentifiers):
		return -1
	case len(leftIdentifiers) > len(rightIdentifiers):
		return 1
	default:
		return 0
	}
}

func comparePrereleaseIdentifier(left string, right string) int {
	leftNumeric := isNumericIdentifier(left)
	rightNumeric := isNumericIdentifier(right)

	switch {
	case leftNumeric && rightNumeric:
		return compareNumericIdentifier(left, right)
	case leftNumeric:
		return -1
	case rightNumeric:
		return 1
	case left < right:
		return -1
	case left > right:
		return 1
	default:
		return 0
	}
}

func isNumericIdentifier(value string) bool {
	if value == "" {
		return false
	}

	for _, char := range value {
		if char < '0' || char > '9' {
			return false
		}
	}

	return true
}

func compareNumericIdentifier(left string, right string) int {
	left = strings.TrimLeft(left, "0")
	right = strings.TrimLeft(right, "0")
	if left == "" {
		left = "0"
	}
	if right == "" {
		right = "0"
	}

	switch {
	case len(left) < len(right):
		return -1
	case len(left) > len(right):
		return 1
	case left < right:
		return -1
	case left > right:
		return 1
	default:
		return 0
	}
}
