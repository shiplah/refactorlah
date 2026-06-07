package project

import (
	"errors"
	"fmt"
	"os"
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
	return r.ResolveFromBase(projectRoot, projectRoot, input)
}

func (r *PathResolver) ResolveMove(projectRoot string, cwd string, oldInput string, newInput string) (string, string, error) {
	base, err := r.baseForOldPath(projectRoot, cwd, oldInput)
	if err != nil {
		return "", "", err
	}

	oldPath, err := r.ResolveFromBase(projectRoot, base, oldInput)
	if err != nil {
		return "", "", err
	}
	newPath, err := r.ResolveFromBase(projectRoot, base, newInput)
	if err != nil {
		return "", "", err
	}

	return oldPath, newPath, nil
}

func (r *PathResolver) ResolveFromBase(projectRoot string, base string, input string) (string, error) {
	if strings.TrimSpace(input) == "" {
		return "", errors.New("path must not be empty")
	}
	var err error
	projectRoot, err = canonicalExistingPath(projectRoot)
	if err != nil {
		return "", err
	}
	base, err = canonicalExistingPath(base)
	if err != nil {
		return "", err
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
		absolute = filepath.Join(base, filepath.FromSlash(cleanedSlash))
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

func (r *PathResolver) baseForOldPath(projectRoot string, cwd string, oldInput string) (string, error) {
	normalizedInput := filepath.ToSlash(strings.ReplaceAll(oldInput, `\`, `/`))
	if filepath.IsAbs(oldInput) || windowsDrivePattern.MatchString(normalizedInput) {
		return cwd, nil
	}

	rootRelative, err := r.ResolveFromBase(projectRoot, projectRoot, oldInput)
	rootErr := err
	cwdRelative, err := r.ResolveFromBase(projectRoot, cwd, oldInput)
	cwdErr := err
	if rootErr != nil && cwdErr != nil {
		return "", rootErr
	}
	if rootErr != nil {
		return cwd, nil
	}
	if cwdErr != nil {
		return projectRoot, nil
	}

	if rootRelative == cwdRelative {
		return projectRoot, nil
	}

	rootExists := projectRelativePathExists(projectRoot, rootRelative)
	cwdExists := projectRelativePathExists(projectRoot, cwdRelative)

	if rootExists && cwdExists {
		return "", fmt.Errorf("path %q is ambiguous; it exists relative to both current directory and project root", oldInput)
	}
	if cwdExists {
		return cwd, nil
	}

	return projectRoot, nil
}

func projectRelativePathExists(projectRoot string, relativePath string) bool {
	absolute := filepath.Join(projectRoot, filepath.FromSlash(relativePath))
	if strings.Contains(relativePath, "*") {
		matches, err := filepath.Glob(absolute)
		return err == nil && len(matches) > 0
	}

	_, err := os.Stat(absolute)
	return err == nil
}

func canonicalExistingPath(path string) (string, error) {
	absolute, err := filepath.Abs(path)
	if err != nil {
		return "", err
	}
	resolved, err := filepath.EvalSymlinks(absolute)
	if err != nil {
		return filepath.Clean(absolute), nil
	}
	return filepath.Clean(resolved), nil
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
