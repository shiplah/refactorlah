package project

import (
	"context"
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

	abs, err := filepath.Abs(cwd)
	if err != nil {
		return RootInfo{}, err
	}
	return RootInfo{ProjectRoot: abs, InGitRepo: false}, nil
}
