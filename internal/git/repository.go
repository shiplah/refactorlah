package git

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"

	"refactorlah/internal/planning"
)

type Repository struct{}

func NewRepository() *Repository {
	return &Repository{}
}

func (r *Repository) DetectRoot(ctx context.Context, cwd string) (string, bool, error) {
	command := exec.CommandContext(ctx, "git", "rev-parse", "--show-toplevel")
	command.Dir = cwd
	output, err := command.Output()
	if err != nil {
		var exitErr *exec.ExitError
		if errors.As(err, &exitErr) {
			return "", false, nil
		}
		return "", false, err
	}
	return strings.TrimSpace(string(output)), true, nil
}

func (r *Repository) IsDirty(ctx context.Context, projectRoot string) (bool, error) {
	command := exec.CommandContext(ctx, "git", "-C", projectRoot, "status", "--porcelain")
	output, err := command.Output()
	if err != nil {
		return false, err
	}
	return strings.TrimSpace(string(output)) != "", nil
}

func (r *Repository) IsTracked(ctx context.Context, projectRoot string, path string) (bool, error) {
	command := exec.CommandContext(ctx, "git", "-C", projectRoot, "ls-files", "--error-unmatch", "--", path)
	if err := command.Run(); err != nil {
		var exitErr *exec.ExitError
		if errors.As(err, &exitErr) {
			return false, nil
		}
		return false, err
	}
	return true, nil
}

func (r *Repository) MoveFiles(ctx context.Context, projectRoot string, moves []planning.FileMove) error {
	for _, move := range moves {
		if err := r.moveFile(ctx, projectRoot, move); err != nil {
			return err
		}
	}
	return r.removeEmptyDirectories(projectRoot, moves)
}

func (r *Repository) moveFile(ctx context.Context, projectRoot string, move planning.FileMove) error {
	oldAbsolute := filepath.Join(projectRoot, filepath.FromSlash(move.OldPath))
	newAbsolute := filepath.Join(projectRoot, filepath.FromSlash(move.NewPath))
	if err := os.MkdirAll(filepath.Dir(newAbsolute), 0o755); err != nil {
		return err
	}

	if move.Tracked {
		command := exec.CommandContext(ctx, "git", "-C", projectRoot, "mv", "--", move.OldPath, move.NewPath)
		if output, err := command.CombinedOutput(); err != nil {
			return fmt.Errorf("git mv %s -> %s failed: %w: %s", move.OldPath, move.NewPath, err, strings.TrimSpace(string(output)))
		}
		return nil
	}

	if err := os.Rename(oldAbsolute, newAbsolute); err != nil {
		return fmt.Errorf("rename %s -> %s failed: %w", move.OldPath, move.NewPath, err)
	}
	return nil
}

func (r *Repository) removeEmptyDirectories(projectRoot string, moves []planning.FileMove) error {
	seen := map[string]struct{}{}
	for _, move := range moves {
		sourceDir := filepath.Dir(move.OldPath)
		for sourceDir != "." && sourceDir != "/" {
			if _, ok := seen[sourceDir]; ok {
				break
			}
			seen[sourceDir] = struct{}{}
			sourceDir = filepath.ToSlash(filepath.Dir(filepath.FromSlash(sourceDir)))
		}
	}

	ordered := make([]string, 0, len(seen))
	for dir := range seen {
		ordered = append(ordered, dir)
	}

	sortDescendingByDepth(ordered)
	for _, dir := range ordered {
		absolute := filepath.Join(projectRoot, filepath.FromSlash(dir))
		entries, err := os.ReadDir(absolute)
		if err != nil {
			if os.IsNotExist(err) {
				continue
			}
			return err
		}
		if len(entries) == 0 {
			if err := os.Remove(absolute); err != nil && !os.IsNotExist(err) {
				return err
			}
		}
	}
	return nil
}

func sortDescendingByDepth(paths []string) {
	sort.Slice(paths, func(i, j int) bool {
		leftDepth := strings.Count(paths[i], "/")
		rightDepth := strings.Count(paths[j], "/")
		if leftDepth == rightDepth {
			return paths[i] > paths[j]
		}
		return leftDepth > rightDepth
	})
}
