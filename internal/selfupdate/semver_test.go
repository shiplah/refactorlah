package selfupdate

import "testing"

func TestCompareSemanticVersionsUsesSemverPrereleasePrecedence(t *testing.T) {
	tests := []struct {
		name     string
		left     string
		right    string
		expected int
	}{
		{
			name:     "numeric prerelease identifiers compare numerically",
			left:     "v1.0.0-rc.10",
			right:    "v1.0.0-rc.2",
			expected: 1,
		},
		{
			name:     "longer prerelease has higher precedence after matching identifiers",
			left:     "v1.0.0-alpha.1",
			right:    "v1.0.0-alpha",
			expected: 1,
		},
		{
			name:     "numeric prerelease identifiers sort before non numeric identifiers",
			left:     "v1.0.0-alpha.1",
			right:    "v1.0.0-alpha.beta",
			expected: -1,
		},
		{
			name:     "build metadata does not affect stable version precedence",
			left:     "v1.0.0+build.2",
			right:    "v1.0.0+build.1",
			expected: 0,
		},
		{
			name:     "build metadata does not affect prerelease precedence",
			left:     "v1.0.0-rc.1+build.2",
			right:    "v1.0.0-rc.1+build.1",
			expected: 0,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			actual, ok := compareSemanticVersions(test.left, test.right)
			if !ok {
				t.Fatalf("expected semantic versions to parse: %s %s", test.left, test.right)
			}
			if comparisonSign(actual) != test.expected {
				t.Fatalf("expected comparison %d, got %d", test.expected, actual)
			}
		})
	}
}

func comparisonSign(value int) int {
	switch {
	case value < 0:
		return -1
	case value > 0:
		return 1
	default:
		return 0
	}
}
