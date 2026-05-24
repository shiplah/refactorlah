package config

import (
	"path/filepath"
	"strings"
)

func resolveExistingPath(path string) (string, error) {
	absolute, err := filepath.Abs(path)
	if err != nil {
		return "", err
	}
	resolved, err := filepath.EvalSymlinks(absolute)
	if err != nil {
		return "", err
	}
	return resolved, nil
}

func depth(path string) int {
	path = strings.Trim(filepath.ToSlash(path), "/")
	if path == "" || path == "." {
		return 0
	}
	return strings.Count(path, "/") + 1
}

func isWithin(root string, candidate string) (bool, error) {
	relative, err := filepath.Rel(root, candidate)
	if err != nil {
		return false, err
	}
	relative = filepath.ToSlash(relative)
	return relative == "." || (!strings.HasPrefix(relative, "../") && relative != ".."), nil
}
