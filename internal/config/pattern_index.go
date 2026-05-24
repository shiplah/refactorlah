package config

import (
	"fmt"
	"path/filepath"
	"strings"
)

type patternIndex struct {
	projectRoot string
	include     []string
	exclude     []string
	seenInclude map[string]struct{}
	seenExclude map[string]struct{}
}

func newPatternIndex(projectRoot string) *patternIndex {
	return &patternIndex{
		projectRoot: projectRoot,
		seenInclude: map[string]struct{}{},
		seenExclude: map[string]struct{}{},
	}
}

func (i *patternIndex) addIncludes(configDir string, patterns []string) error {
	return i.addPatterns(configDir, patterns, &i.include, i.seenInclude)
}

func (i *patternIndex) addExcludes(configDir string, patterns []string) error {
	return i.addPatterns(configDir, patterns, &i.exclude, i.seenExclude)
}

func (i *patternIndex) addPatterns(configDir string, patterns []string, target *[]string, seen map[string]struct{}) error {
	for _, pattern := range patterns {
		absolute, relative, err := i.normalizePattern(configDir, pattern)
		if err != nil {
			return err
		}
		if absolute == "" {
			continue
		}
		if _, ok := seen[absolute]; ok {
			continue
		}
		seen[absolute] = struct{}{}
		*target = append(*target, relative)
	}

	return nil
}

func (i *patternIndex) normalizePattern(configDir string, pattern string) (string, string, error) {
	pattern = strings.TrimSpace(filepath.ToSlash(pattern))
	if pattern == "" {
		return "", "", nil
	}

	var absolute string
	if strings.HasPrefix(pattern, "/") {
		absolute = filepath.Join(i.projectRoot, filepath.FromSlash(strings.TrimPrefix(pattern, "/")))
	} else {
		absolute = filepath.Join(configDir, filepath.FromSlash(pattern))
	}
	absolute = filepath.Clean(absolute)

	inside, err := isWithin(i.projectRoot, absolute)
	if err != nil {
		return "", "", err
	}
	if !inside {
		return "", "", fmt.Errorf("config pattern %q resolves outside project root", pattern)
	}

	relative, err := filepath.Rel(i.projectRoot, absolute)
	if err != nil {
		return "", "", err
	}

	absolute = filepath.ToSlash(absolute)
	relative = filepath.ToSlash(relative)
	if relative == "." {
		relative = ""
	}

	return absolute, relative, nil
}

func (i *patternIndex) config() Config {
	return Config{
		Include: append([]string(nil), i.include...),
		Exclude: append([]string(nil), i.exclude...),
	}
}
