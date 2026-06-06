package config

import (
	"regexp"
	"strings"
)

func (c Config) Allows(path string) bool {
	normalized := normalizePolicyPath(path)
	for _, pattern := range c.Include {
		if globMatches(normalizePolicyPath(pattern), normalized) {
			return true
		}
	}

	for _, pattern := range c.Exclude {
		if globMatches(normalizePolicyPath(pattern), normalized) {
			return false
		}
	}

	return true
}

func normalizePolicyPath(path string) string {
	return strings.Trim(strings.ReplaceAll(path, "\\", "/"), "/")
}

func globMatches(pattern string, path string) bool {
	if pattern == "" {
		return path == ""
	}

	regex := "^" + globToRegexp(pattern) + "$"
	matched, err := regexp.MatchString(regex, path)
	return err == nil && matched
}

func globToRegexp(pattern string) string {
	var builder strings.Builder
	for index := 0; index < len(pattern); index++ {
		character := pattern[index]
		switch character {
		case '*':
			if index+1 < len(pattern) && pattern[index+1] == '*' {
				builder.WriteString(".*")
				index++
				continue
			}
			builder.WriteString("[^/]*")
		case '?':
			builder.WriteString("[^/]")
		default:
			builder.WriteString(regexp.QuoteMeta(string(character)))
		}
	}
	return builder.String()
}
