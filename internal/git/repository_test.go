package git

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"

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
