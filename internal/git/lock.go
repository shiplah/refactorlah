package git

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
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
	if err := waitForLockRelease(ctx, lockPath, "another refactorlah apply is running", options); err != nil {
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
		if err := waitForLockRelease(ctx, lockPath, "another refactorlah apply is running", options); err != nil {
			return nil, err
		}
	}
}

func (r *Repository) WaitForIndexLock(ctx context.Context, projectRoot string, options LockOptions) error {
	gitDir, err := r.gitDir(ctx, projectRoot)
	if err != nil {
		return err
	}

	return waitForLockRelease(ctx, filepath.Join(gitDir, "index.lock"), "git index is locked", options)
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

func waitForLockRelease(ctx context.Context, path string, reason string, options LockOptions) error {
	if _, err := os.Stat(path); errors.Is(err, os.ErrNotExist) {
		return nil
	} else if err != nil {
		return fmt.Errorf("check lock %s: %w", path, err)
	}

	waitInterval := options.WaitInterval
	if waitInterval <= 0 {
		waitInterval = defaultLockWaitInterval
	}
	statusInterval := options.StatusInterval
	if statusInterval <= 0 {
		statusInterval = defaultLockStatusInterval
	}

	started := time.Now()
	nextStatus := started
	for {
		if _, err := os.Stat(path); errors.Is(err, os.ErrNotExist) {
			return nil
		} else if err != nil {
			return fmt.Errorf("check lock %s: %w", path, err)
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
