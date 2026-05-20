package project

import (
	"os"
	"path/filepath"
	"strings"
)

func FindComposerRootForPaths(projectRoot string, paths []string) (string, bool, error) {
	candidateRoots := []string{}
	first := true

	for _, path := range paths {
		roots, err := composerAncestors(projectRoot, path)
		if err != nil {
			return "", false, err
		}
		if len(roots) == 0 {
			continue
		}

		if first {
			candidateRoots = roots
			first = false
			continue
		}

		candidateRoots = intersectStrings(candidateRoots, roots)
	}

	if len(candidateRoots) == 0 {
		if exists, err := composerExists(projectRoot); err != nil {
			return "", false, err
		} else if exists {
			return projectRoot, true, nil
		}
		return "", false, nil
	}

	best := candidateRoots[0]
	for _, candidate := range candidateRoots[1:] {
		if len(candidate) > len(best) {
			best = candidate
		}
	}

	if best == "." {
		return projectRoot, true, nil
	}
	return filepath.Join(projectRoot, filepath.FromSlash(best)), true, nil
}

func composerAncestors(projectRoot string, path string) ([]string, error) {
	normalized := filepath.ToSlash(path)
	if strings.Contains(filepath.Base(normalized), ".") {
		normalized = filepath.ToSlash(filepath.Dir(filepath.FromSlash(normalized)))
	}

	roots := []string{}
	current := normalized
	for current != "." && current != "/" && current != "" {
		candidate := filepath.Join(projectRoot, filepath.FromSlash(current), "composer.json")
		if info, err := os.Stat(candidate); err == nil && !info.IsDir() {
			roots = append(roots, current)
		} else if err != nil && !os.IsNotExist(err) {
			return nil, err
		}

		next := filepath.ToSlash(filepath.Dir(filepath.FromSlash(current)))
		if next == current {
			break
		}
		current = next
	}

	if exists, err := composerExists(projectRoot); err != nil {
		return nil, err
	} else if exists {
		roots = append(roots, ".")
	}

	return roots, nil
}

func composerExists(path string) (bool, error) {
	info, err := os.Stat(filepath.Join(path, "composer.json"))
	if err != nil {
		if os.IsNotExist(err) {
			return false, nil
		}
		return false, err
	}
	return !info.IsDir(), nil
}

func intersectStrings(left []string, right []string) []string {
	set := map[string]struct{}{}
	for _, item := range right {
		set[item] = struct{}{}
	}

	result := []string{}
	for _, item := range left {
		if _, ok := set[item]; ok {
			result = append(result, item)
		}
	}

	return result
}
