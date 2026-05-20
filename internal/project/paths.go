package project

import (
	"errors"
	"fmt"
	"path/filepath"
	"regexp"
	"strings"
)

var windowsDrivePattern = regexp.MustCompile(`^[A-Za-z]:/`)

type PathResolver struct{}

func NewPathResolver() *PathResolver {
	return &PathResolver{}
}

func (r *PathResolver) Resolve(projectRoot string, input string) (string, error) {
	if strings.TrimSpace(input) == "" {
		return "", errors.New("path must not be empty")
	}

	normalizedInput := filepath.ToSlash(strings.ReplaceAll(input, `\`, `/`))
	cleanedSlash := filepath.ToSlash(filepath.Clean(filepath.FromSlash(normalizedInput)))

	var absolute string
	switch {
	case filepath.IsAbs(input):
		absolute = filepath.Clean(input)
	case windowsDrivePattern.MatchString(normalizedInput):
		absolute = filepath.Clean(filepath.FromSlash(normalizedInput))
	default:
		absolute = filepath.Join(projectRoot, filepath.FromSlash(cleanedSlash))
	}

	absolute = filepath.Clean(absolute)
	inside, err := isWithin(projectRoot, absolute)
	if err != nil {
		return "", err
	}
	if !inside {
		return "", fmt.Errorf("path %q resolves outside project root", input)
	}

	relative, err := filepath.Rel(projectRoot, absolute)
	if err != nil {
		return "", err
	}

	relative = filepath.ToSlash(relative)
	if relative == "." || strings.HasPrefix(relative, "../") {
		return "", fmt.Errorf("path %q resolves outside project root", input)
	}

	return relative, nil
}

func isWithin(root string, candidate string) (bool, error) {
	absRoot, err := filepath.Abs(root)
	if err != nil {
		return false, err
	}
	absCandidate, err := filepath.Abs(candidate)
	if err != nil {
		return false, err
	}

	rel, err := filepath.Rel(absRoot, absCandidate)
	if err != nil {
		return false, err
	}
	rel = filepath.ToSlash(rel)
	return rel == "." || (!strings.HasPrefix(rel, "../") && rel != ".."), nil
}
