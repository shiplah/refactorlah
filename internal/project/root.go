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

func (d *RootDetector) Detect(ctx context.Context, cwd string, requireGit bool) (RootInfo, error) {
	if root, ok, err := d.git.DetectRoot(ctx, cwd); err != nil {
		return RootInfo{}, err
	} else if ok {
		return RootInfo{ProjectRoot: root, InGitRepo: true}, nil
	}

	root, found, err := findComposerRoot(cwd)
	if err != nil {
		return RootInfo{}, err
	}
	if found {
		return RootInfo{ProjectRoot: root, InGitRepo: false}, nil
	}

	if requireGit {
		return RootInfo{}, errors.New("could not determine project root inside a git repository; initialize git or remove --require-git")
	}

	abs, err := filepath.Abs(cwd)
	if err != nil {
		return RootInfo{}, err
	}
	return RootInfo{ProjectRoot: abs, InGitRepo: false}, nil
}

func findComposerRoot(start string) (string, bool, error) {
	current, err := filepath.Abs(start)
	if err != nil {
		return "", false, err
	}

	for {
		composerPath := filepath.Join(current, "composer.json")
		info, statErr := os.Stat(composerPath)
		if statErr == nil && !info.IsDir() {
			return current, true, nil
		}
		if !errors.Is(statErr, os.ErrNotExist) && statErr != nil {
			return "", false, statErr
		}

		parent := filepath.Dir(current)
		if parent == current {
			return "", false, nil
		}
		current = parent
	}
}
