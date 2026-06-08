package git

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"sort"
	"strings"
	"time"

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
	transientEntries := transientEntriesByDirectory(moves)
	ordered := make([]string, 0, len(transientEntries))
	for dir := range transientEntries {
		ordered = append(ordered, dir)
	}

	sortDescendingByDepth(ordered)
	return pruneEmptyDirectories(projectRoot, transientEntries, ordered)
}

func transientEntriesByDirectory(moves []planning.FileMove) map[string]map[string]struct{} {
	result := make(map[string]map[string]struct{})
	for _, move := range moves {
		current := filepath.ToSlash(move.OldPath)
		for {
			dir := path.Dir(current)
			if dir == "." || dir == "/" {
				break
			}
			entry := path.Base(current)
			if _, ok := result[dir]; !ok {
				result[dir] = make(map[string]struct{})
			}
			result[dir][entry] = struct{}{}
			current = dir
		}
	}

	return result
}

// Retry only directories whose remaining entries are still expected to
// disappear from the source side of the move chain.
func pruneEmptyDirectories(projectRoot string, transientEntries map[string]map[string]struct{}, directories []string) error {
	const (
		attempts      = 8
		retryInterval = 10 * time.Millisecond
	)

	pending := append([]string(nil), directories...)
	for attempt := 0; attempt < attempts; attempt++ {
		nextPending := make([]string, 0, len(pending))
		for _, dir := range pending {
			prunePath, err := safeProjectDirectoryPath(projectRoot, dir)
			if err != nil {
				return err
			}

			result, err := pruneDirectory(prunePath, transientEntries[dir])
			if err != nil {
				return err
			}
			if result == directoryNeedsRetry {
				nextPending = append(nextPending, dir)
			}
		}

		if len(nextPending) == 0 {
			return nil
		}

		if attempt == attempts-1 {
			return nil
		}

		pending = nextPending
		time.Sleep(retryInterval)
	}

	return nil
}

func safeProjectDirectoryPath(projectRoot string, relativeDir string) (string, error) {
	cleaned := path.Clean(filepath.ToSlash(relativeDir))
	if cleaned == "" || cleaned == "." || cleaned == "/" || path.IsAbs(cleaned) || filepath.IsAbs(relativeDir) {
		return "", fmt.Errorf("refusing to prune unsafe project directory %q", relativeDir)
	}
	if cleaned == ".." || strings.HasPrefix(cleaned, "../") {
		return "", fmt.Errorf("refusing to prune directory outside project root: %q", relativeDir)
	}

	root, err := canonicalProjectDirectory(projectRoot)
	if err != nil {
		return "", err
	}

	current := root
	for _, segment := range strings.Split(cleaned, "/") {
		if segment == "" || segment == "." || segment == ".." {
			return "", fmt.Errorf("refusing to prune unsafe project directory %q", relativeDir)
		}

		current = filepath.Join(current, filepath.FromSlash(segment))
		info, err := os.Lstat(current)
		if err != nil {
			if os.IsNotExist(err) {
				return filepath.Clean(current), nil
			}
			return "", err
		}
		if info.Mode()&os.ModeSymlink != 0 {
			return "", fmt.Errorf("refusing to prune symlinked directory path %q", relativeDir)
		}
		if !info.IsDir() {
			return "", fmt.Errorf("refusing to prune non-directory path %q", relativeDir)
		}
	}

	if current == root {
		return "", fmt.Errorf("refusing to prune project root: %q", relativeDir)
	}

	return current, nil
}

func canonicalProjectDirectory(path string) (string, error) {
	absolute, err := filepath.Abs(path)
	if err != nil {
		return "", err
	}
	resolved, err := filepath.EvalSymlinks(absolute)
	if err != nil {
		return filepath.Clean(absolute), nil
	}
	return filepath.Clean(resolved), nil
}

type directoryPruneResult int

const (
	directoryRemoved directoryPruneResult = iota
	directoryNeedsRetry
	directoryDone
)

func pruneDirectory(path string, transientEntries map[string]struct{}) (directoryPruneResult, error) {
	entries, err := os.ReadDir(path)
	if err != nil {
		if os.IsNotExist(err) {
			return directoryRemoved, nil
		}
		return directoryDone, err
	}
	if len(entries) != 0 {
		if directoryContainsOnlyTransientEntries(entries, transientEntries) {
			return directoryNeedsRetry, nil
		}
		return directoryDone, nil
	}
	if err := os.Remove(path); err != nil && !os.IsNotExist(err) {
		return directoryDone, err
	}
	return directoryRemoved, nil
}

func directoryContainsOnlyTransientEntries(entries []os.DirEntry, transientEntries map[string]struct{}) bool {
	if len(transientEntries) == 0 {
		return false
	}

	for _, entry := range entries {
		if _, ok := transientEntries[entry.Name()]; !ok {
			return false
		}
	}

	return true
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
