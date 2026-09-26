package git

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

const (
	defaultLockWaitInterval   = 500 * time.Millisecond
	defaultLockStatusInterval = 5 * time.Second
)

type LockOptions struct {
	Writer         io.Writer
	WaitInterval   time.Duration
	StatusInterval time.Duration
}

type WorktreeLock struct {
	path  string
	token string
}

func (r *Repository) AcquireApplyLock(ctx context.Context, projectRoot string, options LockOptions) (*WorktreeLock, error) {
	gitDir, err := r.gitDir(ctx, projectRoot)
	if err != nil {
		return nil, err
	}

	lockPath := filepath.Join(gitDir, "refactorlah.lock")
	token := fmt.Sprintf("pid=%d\ncreated=%s\n", os.Getpid(), time.Now().UTC().Format(time.RFC3339Nano))
	if err := waitForLockRelease(ctx, lockPath, "another refactorlah apply is running", options, true); err != nil {
		return nil, err
	}

	for {
		file, err := os.OpenFile(lockPath, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0o600)
		if err == nil {
			if _, writeErr := file.WriteString(token); writeErr != nil {
				_ = file.Close()
				_ = os.Remove(lockPath)
				return nil, writeErr
			}
			if closeErr := file.Close(); closeErr != nil {
				_ = os.Remove(lockPath)
				return nil, closeErr
			}

			return &WorktreeLock{path: lockPath, token: token}, nil
		}

		if !errors.Is(err, os.ErrExist) {
			return nil, fmt.Errorf("create refactorlah lock %s: %w", lockPath, err)
		}
		if err := waitForLockRelease(ctx, lockPath, "another refactorlah apply is running", options, true); err != nil {
			return nil, err
		}
	}
}

func (r *Repository) WaitForIndexLock(ctx context.Context, projectRoot string, options LockOptions) error {
	gitDir, err := r.gitDir(ctx, projectRoot)
	if err != nil {
		return err
	}

	return waitForLockRelease(ctx, filepath.Join(gitDir, "index.lock"), "git index is locked", options, false)
}

func (l *WorktreeLock) Release() error {
	if l == nil || l.path == "" {
		return nil
	}

	content, err := os.ReadFile(l.path)
	if errors.Is(err, os.ErrNotExist) {
		return nil
	}
	if err != nil {
		return err
	}
	if string(content) != l.token {
		return fmt.Errorf("refactorlah lock changed while running: %s", l.path)
	}

	return os.Remove(l.path)
}

func waitForLockRelease(ctx context.Context, path string, reason string, options LockOptions, reclaimStale bool) error {
	waitInterval := options.WaitInterval
	if waitInterval <= 0 {
		waitInterval = defaultLockWaitInterval
	}
	statusInterval := options.StatusInterval
	if statusInterval <= 0 {
		statusInterval = defaultLockStatusInterval
	}

	var started time.Time
	var nextStatus time.Time
	for {
		if _, err := os.Stat(path); errors.Is(err, os.ErrNotExist) {
			return nil
		} else if err != nil {
			return fmt.Errorf("check lock %s: %w", path, err)
		}
		if reclaimStale {
			removed, err := removeStaleApplyLock(path)
			if err != nil {
				return err
			}
			if removed {
				if options.Writer != nil {
					_, _ = fmt.Fprintf(options.Writer, "removed stale refactorlah lock at %s\n", path)
				}
				return nil
			}
		}
		if started.IsZero() {
			started = time.Now()
			nextStatus = started.Add(statusInterval)
			if options.Writer != nil {
				_, _ = fmt.Fprintf(options.Writer, "waiting for %s at %s (%s)\n", reason, path, 0*time.Second)
			}
		}

		now := time.Now()
		if options.Writer != nil && !now.Before(nextStatus) {
			_, _ = fmt.Fprintf(options.Writer, "waiting for %s at %s (%s)\n", reason, path, now.Sub(started).Round(time.Second))
			nextStatus = now.Add(statusInterval)
		}

		timer := time.NewTimer(waitInterval)
		select {
		case <-ctx.Done():
			timer.Stop()
			return fmt.Errorf("waiting for lock %s: %w", path, ctx.Err())
		case <-timer.C:
		}
	}
}

func removeStaleApplyLock(path string) (bool, error) {
	content, err := os.ReadFile(path)
	if errors.Is(err, os.ErrNotExist) {
		return false, nil
	}
	if err != nil {
		return false, fmt.Errorf("read refactorlah lock %s: %w", path, err)
	}

	firstLine, _, _ := strings.Cut(string(content), "\n")
	pidText, ok := strings.CutPrefix(firstLine, "pid=")
	pid, parseErr := strconv.Atoi(pidText)
	if !ok || parseErr != nil || pid <= 0 {
		return false, fmt.Errorf("refactorlah lock %s has no valid owner PID; confirm no apply is running before removing it", path)
	}

	alive, err := processAlive(pid)
	if err != nil {
		return false, fmt.Errorf("check owner of refactorlah lock %s: %w", path, err)
	}
	if alive {
		return false, nil
	}

	current, err := os.ReadFile(path)
	if errors.Is(err, os.ErrNotExist) {
		return false, nil
	}
	if err != nil {
		return false, fmt.Errorf("recheck refactorlah lock %s: %w", path, err)
	}
	if !bytes.Equal(current, content) {
		return false, nil
	}
	if err := os.Remove(path); errors.Is(err, os.ErrNotExist) {
		return false, nil
	} else if err != nil {
		return false, fmt.Errorf("remove stale refactorlah lock %s: %w", path, err)
	}
	return true, nil
}
