package project

import (
	"os"
	"path/filepath"
	"sort"
	"strings"
)

func FindMarkerRootForPaths(projectRoot string, paths []string, markers []string) (string, bool, error) {
	candidateRoots := []string{}
	first := true

	for _, path := range paths {
		roots, err := markerAncestors(projectRoot, path, markers)
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
		if exists, err := hasAnyMarker(projectRoot, markers); err != nil {
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

func FindMarkerRootsForPaths(projectRoot string, paths []string, markers []string) ([]string, bool, error) {
	rootsByPath := map[string]bool{}
	for _, path := range paths {
		roots, err := markerAncestors(projectRoot, path, markers)
		if err != nil {
			return nil, false, err
		}
		if len(roots) == 0 {
			continue
		}

		rootsByPath[deepestMarkerRoot(roots)] = true
	}

	if len(rootsByPath) == 0 {
		if exists, err := hasAnyMarker(projectRoot, markers); err != nil {
			return nil, false, err
		} else if exists {
			return []string{projectRoot}, true, nil
		}
		return nil, false, nil
	}

	roots := make([]string, 0, len(rootsByPath))
	for root := range rootsByPath {
		if root == "." {
			roots = append(roots, projectRoot)
			continue
		}
		roots = append(roots, filepath.Join(projectRoot, filepath.FromSlash(root)))
	}
	sort.Strings(roots)
	return roots, true, nil
}

func markerAncestors(projectRoot string, path string, markers []string) ([]string, error) {
	normalized := filepath.ToSlash(path)
	if strings.Contains(filepath.Base(normalized), ".") {
		normalized = filepath.ToSlash(filepath.Dir(filepath.FromSlash(normalized)))
	}

	roots := []string{}
	current := normalized
	for current != "." && current != "/" && current != "" {
		candidateRoot := filepath.Join(projectRoot, filepath.FromSlash(current))
		if exists, err := hasAnyMarker(candidateRoot, markers); err != nil {
			return nil, err
		} else if exists {
			roots = append(roots, current)
		}

		next := filepath.ToSlash(filepath.Dir(filepath.FromSlash(current)))
		if next == current {
			break
		}
		current = next
	}

	if exists, err := hasAnyMarker(projectRoot, markers); err != nil {
		return nil, err
	} else if exists {
		roots = append(roots, ".")
	}

	return roots, nil
}

func deepestMarkerRoot(roots []string) string {
	best := roots[0]
	for _, root := range roots[1:] {
		if len(root) > len(best) {
			best = root
		}
	}
	return best
}

func hasAnyMarker(root string, markers []string) (bool, error) {
	for _, marker := range markers {
		info, err := os.Stat(filepath.Join(root, marker))
		if err == nil {
			if !info.IsDir() {
				return true, nil
			}
			continue
		}
		if !os.IsNotExist(err) {
			return false, err
		}
	}

	return false, nil
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
