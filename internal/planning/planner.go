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
	return p.BuildMany(ctx, root, []RequestedMove{{
		OldPath: oldPath,
		NewPath: newPath,
	}}, track)
}

func (p *Planner) BuildMany(ctx context.Context, root string, requests []RequestedMove, track TrackFunc) (MovePlan, error) {
	_ = ctx

	if len(requests) == 0 {
		return MovePlan{}, errors.New("at least one move is required")
	}

	aggregate := MovePlan{
		OldPath: requests[0].OldPath,
		NewPath: requests[0].NewPath,
		IsDir:   len(requests) == 1,
	}
	seenSources := map[string]struct{}{}
	seenTargets := map[string]struct{}{}

	for _, request := range requests {
		plan, err := p.buildSingle(root, request.OldPath, request.NewPath, track)
		if err != nil {
			return MovePlan{}, err
		}
		for _, move := range plan.Moves {
			if _, exists := seenSources[move.OldPath]; exists {
				return MovePlan{}, fmt.Errorf("duplicate source path %q in move set", move.OldPath)
			}
			if _, exists := seenTargets[move.NewPath]; exists {
				return MovePlan{}, fmt.Errorf("duplicate target path %q in move set", move.NewPath)
			}
			seenSources[move.OldPath] = struct{}{}
			seenTargets[move.NewPath] = struct{}{}
			aggregate.Moves = append(aggregate.Moves, move)
		}
	}

	sort.Slice(aggregate.Moves, func(i, j int) bool {
		return aggregate.Moves[i].OldPath < aggregate.Moves[j].OldPath
	})

	return aggregate, nil
}

func (p *Planner) buildSingle(root string, oldPath string, newPath string, track TrackFunc) (MovePlan, error) {
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
