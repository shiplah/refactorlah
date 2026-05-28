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

func (r *Repository) gitDir(ctx context.Context, projectRoot string) (string, error) {
	command := exec.CommandContext(ctx, "git", "-C", projectRoot, "rev-parse", "--git-dir")
	output, err := command.Output()
	if err != nil {
		return "", err
	}

	gitDir := strings.TrimSpace(string(output))
	if filepath.IsAbs(gitDir) {
		return gitDir, nil
	}

	return filepath.Clean(filepath.Join(projectRoot, gitDir)), nil
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

func (r *Repository) MoveFiles(ctx context.Context, projectRoot string, moves []planning.FileMove, options ...LockOptions) error {
	lockOptions := LockOptions{}
	if len(options) > 0 {
		lockOptions = options[0]
	}

	for _, move := range moves {
		if err := r.moveFile(ctx, projectRoot, move, lockOptions); err != nil {
			return err
		}
	}
	return r.removeEmptyDirectories(projectRoot, moves)
}

func (r *Repository) moveFile(ctx context.Context, projectRoot string, move planning.FileMove, lockOptions LockOptions) error {
	oldAbsolute := filepath.Join(projectRoot, filepath.FromSlash(move.OldPath))
	newAbsolute := filepath.Join(projectRoot, filepath.FromSlash(move.NewPath))
	if err := os.MkdirAll(filepath.Dir(newAbsolute), 0o755); err != nil {
		return err
	}

	if move.Tracked {
		for {
			if err := r.WaitForIndexLock(ctx, projectRoot, lockOptions); err != nil {
				return err
			}

			command := exec.CommandContext(ctx, "git", "-C", projectRoot, "mv", "--", move.OldPath, move.NewPath)
			if output, err := command.CombinedOutput(); err != nil {
				trimmedOutput := strings.TrimSpace(string(output))
				if isIndexLockFailure(trimmedOutput) {
					continue
				}

				return fmt.Errorf("git mv %s -> %s failed: %w: %s", move.OldPath, move.NewPath, err, trimmedOutput)
			}

			return nil
		}
	}

	if err := os.Rename(oldAbsolute, newAbsolute); err != nil {
		return fmt.Errorf("rename %s -> %s failed: %w", move.OldPath, move.NewPath, err)
	}
	return nil
}

func isIndexLockFailure(output string) bool {
	return strings.Contains(output, "index.lock") || strings.Contains(output, "Unable to create")
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
