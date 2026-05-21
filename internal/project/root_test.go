package project

import (
	"context"
	"os"
	"path/filepath"
	"testing"
)

type stubGitRootDetector struct {
	root string
	ok   bool
	err  error
}

func (s stubGitRootDetector) DetectRoot(context.Context, string) (string, bool, error) {
	return s.root, s.ok, s.err
}

func TestRootDetectorFallsBackToCurrentDirectoryWithoutGitOrComposer(t *testing.T) {
	cwd := t.TempDir()
	detector := NewRootDetector(stubGitRootDetector{})

	info, err := detector.Detect(t.Context(), cwd)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if info.ProjectRoot != cwd {
		t.Fatalf("expected project root %q, got %q", cwd, info.ProjectRoot)
	}
	if info.InGitRepo {
		t.Fatal("expected non-git root info")
	}
}

func TestRootDetectorPrefersGitRootOverComposerRoot(t *testing.T) {
	cwd := t.TempDir()
	if err := os.WriteFile(filepath.Join(cwd, "composer.json"), []byte("{}"), 0o644); err != nil {
		t.Fatal(err)
	}

	detector := NewRootDetector(stubGitRootDetector{
		root: filepath.Join(cwd, "git-root"),
		ok:   true,
	})

	info, err := detector.Detect(t.Context(), cwd)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if info.ProjectRoot != filepath.Join(cwd, "git-root") {
		t.Fatalf("expected git root, got %q", info.ProjectRoot)
	}
	if !info.InGitRepo {
		t.Fatal("expected git repo root info")
	}
}
