package project

import (
	"context"
	"errors"
	"os"
	"path/filepath"
)

type GitRootDetector interface {
	DetectRoot(ctx context.Context, cwd string) (string, bool, error)
}

type RootInfo struct {
	ProjectRoot string
	InGitRepo   bool
}

type RootDetector struct {
	git GitRootDetector
}

func NewRootDetector(git GitRootDetector) *RootDetector {
	return &RootDetector{git: git}
}

func (d *RootDetector) Detect(ctx context.Context, cwd string) (RootInfo, error) {
	if root, ok, err := d.git.DetectRoot(ctx, cwd); err != nil {
		return RootInfo{}, err
	} else if ok {
		return RootInfo{ProjectRoot: root, InGitRepo: true}, nil
	}

	root, found, err := findConfiguredRoot(cwd)
	if err != nil {
		return RootInfo{}, err
	}
	if found {
		return RootInfo{ProjectRoot: root, InGitRepo: false}, nil
	}

	root, found, err = findProjectMarkerRoot(cwd)
	if err != nil {
		return RootInfo{}, err
	}
	if found {
		return RootInfo{ProjectRoot: root, InGitRepo: false}, nil
	}

	abs, err := filepath.Abs(cwd)
	if err != nil {
		return RootInfo{}, err
	}
	return RootInfo{ProjectRoot: abs, InGitRepo: false}, nil
}

func findConfiguredRoot(start string) (string, bool, error) {
	current, err := filepath.Abs(start)
	if err != nil {
		return "", false, err
	}

	var root string
	for {
		configPath := filepath.Join(current, ".refactorlah.json")
		info, statErr := os.Stat(configPath)
		if statErr == nil && !info.IsDir() {
			root = current
		}
		if !errors.Is(statErr, os.ErrNotExist) && statErr != nil {
			return "", false, statErr
		}

		parent := filepath.Dir(current)
		if parent == current {
			break
		}
		current = parent
	}

	return root, root != "", nil
}

func findProjectMarkerRoot(start string) (string, bool, error) {
	return findNearestRootWithAnyMarker(start, []string{
		"composer.json",
		"package.json",
		"pnpm-workspace.yaml",
		"yarn.lock",
		"pyproject.toml",
		"setup.py",
		"go.mod",
		"Cargo.toml",
	})
}

func findComposerRoot(start string) (string, bool, error) {
	return findNearestRootWithAnyMarker(start, []string{"composer.json"})
}

func findNearestRootWithAnyMarker(start string, markers []string) (string, bool, error) {
	current, err := filepath.Abs(start)
	if err != nil {
		return "", false, err
	}

	for {
		for _, marker := range markers {
			markerPath := filepath.Join(current, marker)
			info, statErr := os.Stat(markerPath)
			if statErr == nil && !info.IsDir() {
				return current, true, nil
			}
			if !errors.Is(statErr, os.ErrNotExist) && statErr != nil {
				return "", false, statErr
			}
		}

		parent := filepath.Dir(current)
		if parent == current {
			return "", false, nil
		}
		current = parent
	}
}
