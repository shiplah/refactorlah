package planning

import (
	"context"
	"errors"
	"fmt"
	"path/filepath"
	"sort"
	"strings"

	"refactorlah/internal/files"
)

var ErrTargetExists = errors.New("target path already exists")

type TrackFunc func(path string) (bool, error)

type Planner struct{}

func NewPlanner() *Planner {
	return &Planner{}
}

func (p *Planner) Build(ctx context.Context, root string, oldPath string, newPath string, track TrackFunc) (MovePlan, error) {
	_ = ctx

	exists, oldInfo, err := files.Exists(root, oldPath)
	if err != nil {
		return MovePlan{}, err
	}
	if !exists {
		return MovePlan{}, fmt.Errorf("old path %q does not exist", oldPath)
	}

	if exists, _, err := files.Exists(root, newPath); err != nil {
		return MovePlan{}, err
	} else if exists {
		return MovePlan{}, ErrTargetExists
	}

	plan := MovePlan{
		OldPath: oldPath,
		NewPath: newPath,
		IsDir:   oldInfo.IsDir(),
	}

	if !oldInfo.IsDir() {
		move, err := buildFileMove(oldPath, newPath, track)
		if err != nil {
			return MovePlan{}, err
		}
		plan.Moves = []FileMove{move}
		return plan, nil
	}

	filesToMove, err := files.CollectFiles(root, oldPath)
	if err != nil {
		return MovePlan{}, err
	}

	sort.Strings(filesToMove)
	for _, source := range filesToMove {
		suffix := strings.TrimPrefix(source, oldPath)
		suffix = strings.TrimPrefix(suffix, "/")
		target := filepath.ToSlash(filepath.Join(filepath.FromSlash(newPath), filepath.FromSlash(suffix)))
		move, err := buildFileMove(source, target, track)
		if err != nil {
			return MovePlan{}, err
		}
		plan.Moves = append(plan.Moves, move)
	}

	return plan, nil
}

func buildFileMove(oldPath string, newPath string, track TrackFunc) (FileMove, error) {
	tracked, err := track(oldPath)
	if err != nil {
		return FileMove{}, err
	}

	mover := "filesystem rename"
	if tracked {
		mover = "git mv"
	}

	return FileMove{
		OldPath: oldPath,
		NewPath: newPath,
		Tracked: tracked,
		Mover:   mover,
	}, nil
}
