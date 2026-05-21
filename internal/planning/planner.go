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
	if len(requests) == 1 {
		return p.buildSingle(root, requests[0].OldPath, requests[0].NewPath, track)
	}

	return p.buildVirtual(root, requests, track)
}

func (p *Planner) buildVirtual(root string, requests []RequestedMove, track TrackFunc) (MovePlan, error) {
	aggregate := MovePlan{
		OldPath: requests[0].OldPath,
		NewPath: requests[0].NewPath,
		IsDir:   false,
	}
	filesInProject, err := files.CollectFiles(root, ".")
	if err != nil {
		return MovePlan{}, err
	}

	virtualFiles := make([]virtualFile, 0, len(filesInProject))
	for _, path := range filesInProject {
		tracked, err := track(path)
		if err != nil {
			return MovePlan{}, err
		}
		virtualFiles = append(virtualFiles, virtualFile{
			SourcePath:  path,
			CurrentPath: path,
			Tracked:     tracked,
		})
	}

	for _, request := range requests {
		matches, isDir, err := matchVirtualFiles(virtualFiles, request.OldPath)
		if err != nil {
			return MovePlan{}, err
		}
		if err := ensureVirtualTargetAvailable(virtualFiles, matches, request.NewPath); err != nil {
			return MovePlan{}, err
		}

		for _, index := range matches {
			currentPath := virtualFiles[index].CurrentPath
			nextPath := request.NewPath
			if isDir {
				suffix := strings.TrimPrefix(currentPath, request.OldPath)
				suffix = strings.TrimPrefix(suffix, "/")
				nextPath = filepath.ToSlash(filepath.Join(filepath.FromSlash(request.NewPath), filepath.FromSlash(suffix)))
			}
			virtualFiles[index].CurrentPath = nextPath
		}
	}

	for _, file := range virtualFiles {
		if file.SourcePath == file.CurrentPath {
			continue
		}

		mover := "filesystem rename"
		if file.Tracked {
			mover = "git mv"
		}

		aggregate.Moves = append(aggregate.Moves, FileMove{
			OldPath: file.SourcePath,
			NewPath: file.CurrentPath,
			Tracked: file.Tracked,
			Mover:   mover,
		})
	}

	sort.Slice(aggregate.Moves, func(i, j int) bool {
		return aggregate.Moves[i].OldPath < aggregate.Moves[j].OldPath
	})

	return aggregate, nil
}

type virtualFile struct {
	SourcePath  string
	CurrentPath string
	Tracked     bool
}

func matchVirtualFiles(files []virtualFile, oldPath string) ([]int, bool, error) {
	exact := []int{}
	for index, file := range files {
		if file.CurrentPath == oldPath {
			exact = append(exact, index)
		}
	}
	if len(exact) > 0 {
		return exact, false, nil
	}

	prefix := oldPath + "/"
	matches := []int{}
	for index, file := range files {
		if strings.HasPrefix(file.CurrentPath, prefix) {
			matches = append(matches, index)
		}
	}
	if len(matches) > 0 {
		return matches, true, nil
	}

	return nil, false, fmt.Errorf("old path %q does not exist", oldPath)
}

func ensureVirtualTargetAvailable(files []virtualFile, moving []int, newPath string) error {
	movingSet := make(map[int]struct{}, len(moving))
	for _, index := range moving {
		movingSet[index] = struct{}{}
	}

	prefix := newPath + "/"
	for index, file := range files {
		if _, ok := movingSet[index]; ok {
			continue
		}
		if file.CurrentPath == newPath || strings.HasPrefix(file.CurrentPath, prefix) {
			return ErrTargetExists
		}
	}

	return nil
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
