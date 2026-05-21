package git

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
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

func TestStageFilesStagesSemanticEdits(t *testing.T) {
	root := t.TempDir()
	runGit(t, root, "init")
	runGit(t, root, "config", "user.email", "test@example.com")
	runGit(t, root, "config", "user.name", "Test User")

	trackedPath := filepath.Join(root, "app", "Tracked.php")
	mustWriteGitFile(t, trackedPath)
	runGit(t, root, "add", "app/Tracked.php")
	runGit(t, root, "commit", "-m", "initial")

	if err := os.WriteFile(trackedPath, []byte("<?php\n\ndeclare(strict_types=1);\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	repo := NewRepository()
	if err := repo.StageFiles(t.Context(), root, []string{"app/Tracked.php"}); err != nil {
		t.Fatalf("stage files failed: %v", err)
	}

	status := runGitOutput(t, root, "status", "--short")
	if strings.TrimSpace(status) != "M  app/Tracked.php" {
		t.Fatalf("expected staged semantic edit, got status %q", status)
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
