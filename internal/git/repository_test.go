package git

import (
	"bytes"
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
	"time"

	"refactorlah/internal/planning"
)

func TestMoveFilesHandlesTrackedAndUntrackedFiles(t *testing.T) {
	root := t.TempDir()
	runGit(t, root, "init")
	runGit(t, root, "config", "user.email", "test@example.com")
	runGit(t, root, "config", "user.name", "Test User")

	trackedPath := filepath.Join(root, "app", "Tracked.php")
	untrackedPath := filepath.Join(root, "app", "Scratch.php")
	mustWriteGitFile(t, trackedPath)
	mustWriteGitFile(t, untrackedPath)

	runGit(t, root, "add", "app/Tracked.php")
	runGit(t, root, "commit", "-m", "initial")

	repo := NewRepository()
	err := repo.MoveFiles(t.Context(), root, []planning.FileMove{
		{OldPath: "app/Tracked.php", NewPath: "app/MovedTracked.php", Tracked: true},
		{OldPath: "app/Scratch.php", NewPath: "app/MovedScratch.php", Tracked: false},
	})
	if err != nil {
		t.Fatalf("move files failed: %v", err)
	}

	if _, err := os.Stat(filepath.Join(root, "app", "MovedTracked.php")); err != nil {
		t.Fatalf("tracked move missing: %v", err)
	}
	if _, err := os.Stat(filepath.Join(root, "app", "MovedScratch.php")); err != nil {
		t.Fatalf("untracked move missing: %v", err)
	}
}

func TestMoveFilesRemovesEmptySourceDirectories(t *testing.T) {
	root := t.TempDir()
	runGit(t, root, "init")
	runGit(t, root, "config", "user.email", "test@example.com")
	runGit(t, root, "config", "user.name", "Test User")

	sourcePath := filepath.Join(root, "src", "Old", "Nested", "Thing.php")
	mustWriteGitFile(t, sourcePath)
	runGit(t, root, "add", "src/Old/Nested/Thing.php")
	runGit(t, root, "commit", "-m", "initial")

	repo := NewRepository()
	err := repo.MoveFiles(t.Context(), root, []planning.FileMove{
		{
			OldPath: "src/Old/Nested/Thing.php",
			NewPath: "src/New/Thing.php",
			Tracked: true,
			Mover:   "git mv",
		},
	})
	if err != nil {
		t.Fatalf("move files failed: %v", err)
	}

	if _, err := os.Stat(filepath.Join(root, "src", "New", "Thing.php")); err != nil {
		t.Fatalf("moved file missing: %v", err)
	}
	if _, err := os.Stat(filepath.Join(root, "src", "Old")); !os.IsNotExist(err) {
		t.Fatalf("expected emptied source directory to be removed, got err: %v", err)
	}
	if _, err := os.Stat(filepath.Join(root, "src")); err != nil {
		t.Fatalf("expected non-empty ancestor to remain: %v", err)
	}
}

func TestRemoveEmptyDirectoriesWaitsForSourceTreeToBecomeEmpty(t *testing.T) {
	root := t.TempDir()
	oldFile := filepath.Join(root, "src", "Old", "Nested", "Thing.php")
	newFile := filepath.Join(root, "src", "New", "Thing.php")
	mustWriteGitFile(t, oldFile)
	mustWriteGitFile(t, newFile)

	go func() {
		time.Sleep(20 * time.Millisecond)
		_ = os.Remove(oldFile)
	}()

	repo := NewRepository()
	err := repo.removeEmptyDirectories(root, []planning.FileMove{
		{
			OldPath: "src/Old/Nested/Thing.php",
			NewPath: "src/New/Thing.php",
		},
	})
	if err != nil {
		t.Fatalf("remove empty directories failed: %v", err)
	}

	if _, err := os.Stat(filepath.Join(root, "src", "Old")); !os.IsNotExist(err) {
		t.Fatalf("expected source directory to be removed after transient file disappeared, got err: %v", err)
	}
	if _, err := os.Stat(filepath.Join(root, "src")); err != nil {
		t.Fatalf("expected destination ancestor to remain: %v", err)
	}
}

func TestRemoveEmptyDirectoriesRejectsPathsOutsideProjectRoot(t *testing.T) {
	parent := t.TempDir()
	root := filepath.Join(parent, "project")
	outside := filepath.Join(parent, "outside")
	if err := os.MkdirAll(root, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(outside, 0o755); err != nil {
		t.Fatal(err)
	}

	repo := NewRepository()
	err := repo.removeEmptyDirectories(root, []planning.FileMove{
		{
			OldPath: "../outside/Thing.php",
			NewPath: "src/New/Thing.php",
		},
	})
	if err == nil || !strings.Contains(err.Error(), "outside project root") {
		t.Fatalf("expected outside-project-root error, got %v", err)
	}
	if _, statErr := os.Stat(outside); statErr != nil {
		t.Fatalf("expected outside directory to remain, got %v", statErr)
	}
}

func TestRemoveEmptyDirectoriesRejectsSymlinkedDirectories(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("symlink permissions vary on Windows")
	}

	parent := t.TempDir()
	root := filepath.Join(parent, "project")
	outside := filepath.Join(parent, "outside")
	if err := os.MkdirAll(root, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(outside, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink(outside, filepath.Join(root, "linked")); err != nil {
		t.Fatal(err)
	}

	repo := NewRepository()
	err := repo.removeEmptyDirectories(root, []planning.FileMove{
		{
			OldPath: "linked/Thing.php",
			NewPath: "src/New/Thing.php",
		},
	})
	if err == nil || !strings.Contains(err.Error(), "symlinked directory path") {
		t.Fatalf("expected symlink rejection error, got %v", err)
	}
}

func TestAcquireApplyLockWaitsForExistingRefactorlahLock(t *testing.T) {
	root := t.TempDir()
	runGit(t, root, "init")

	repo := NewRepository()
	gitDir, err := repo.gitDir(t.Context(), root)
	if err != nil {
		t.Fatal(err)
	}
	lockPath := filepath.Join(gitDir, "refactorlah.lock")
	if err := os.WriteFile(lockPath, []byte("other process\n"), 0o600); err != nil {
		t.Fatal(err)
	}

	go func() {
		time.Sleep(20 * time.Millisecond)
		_ = os.Remove(lockPath)
	}()

	var stderr bytes.Buffer
	ctx, cancel := context.WithTimeout(t.Context(), time.Second)
	defer cancel()

	lock, err := repo.AcquireApplyLock(ctx, root, LockOptions{
		Writer:         &stderr,
		WaitInterval:   time.Millisecond,
		StatusInterval: time.Millisecond,
	})
	if err != nil {
		t.Fatalf("acquire lock failed: %v", err)
	}
	defer func() {
		if err := lock.Release(); err != nil {
			t.Fatalf("release lock failed: %v", err)
		}
	}()

	if !strings.Contains(stderr.String(), "another refactorlah apply is running") {
		t.Fatalf("expected waiting output, got %q", stderr.String())
	}
}

func TestMoveFilesWaitsForGitIndexLock(t *testing.T) {
	root := t.TempDir()
	runGit(t, root, "init")
	runGit(t, root, "config", "user.email", "test@example.com")
	runGit(t, root, "config", "user.name", "Test User")

	sourcePath := filepath.Join(root, "src", "Thing.php")
	mustWriteGitFile(t, sourcePath)
	runGit(t, root, "add", "src/Thing.php")
	runGit(t, root, "commit", "-m", "initial")

	repo := NewRepository()
	gitDir, err := repo.gitDir(t.Context(), root)
	if err != nil {
		t.Fatal(err)
	}
	indexLockPath := filepath.Join(gitDir, "index.lock")
	if err := os.WriteFile(indexLockPath, []byte("locked\n"), 0o600); err != nil {
		t.Fatal(err)
	}

	go func() {
		time.Sleep(20 * time.Millisecond)
		_ = os.Remove(indexLockPath)
	}()

	var stderr bytes.Buffer
	ctx, cancel := context.WithTimeout(t.Context(), time.Second)
	defer cancel()

	err = repo.MoveFiles(ctx, root, []planning.FileMove{
		{
			OldPath: "src/Thing.php",
			NewPath: "src/MovedThing.php",
			Tracked: true,
			Mover:   "git mv",
		},
	}, LockOptions{
		Writer:         &stderr,
		WaitInterval:   time.Millisecond,
		StatusInterval: time.Millisecond,
	})
	if err != nil {
		t.Fatalf("move files failed: %v", err)
	}
	if !strings.Contains(stderr.String(), "git index is locked") {
		t.Fatalf("expected waiting output, got %q", stderr.String())
	}
	if _, err := os.Stat(filepath.Join(root, "src", "MovedThing.php")); err != nil {
		t.Fatalf("moved file missing: %v", err)
	}
}

func runGit(t *testing.T, dir string, args ...string) {
	t.Helper()
	if output, err := runGitCommand(dir, args...); err != nil {
		t.Fatalf("git %v failed: %v: %s", args, err, output)
	}
}

func runGitOutput(t *testing.T, dir string, args ...string) string {
	t.Helper()
	output, err := runGitCommand(dir, args...)
	if err != nil {
		t.Fatalf("git %v failed: %v: %s", args, err, output)
	}
	return output
}

func runGitCommand(dir string, args ...string) (string, error) {
	command := exec.Command("git", args...)
	command.Dir = dir
	output, err := command.CombinedOutput()
	return string(output), err
}

func mustWriteGitFile(t *testing.T, path string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte("<?php\n"), 0o644); err != nil {
		t.Fatal(err)
	}
}
