package project

import (
	"context"
	"strings"
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

	info, err := detector.Detect(t.Context(), cwd, false)
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

func TestRootDetectorCanRequireGit(t *testing.T) {
	cwd := t.TempDir()
	detector := NewRootDetector(stubGitRootDetector{})

	_, err := detector.Detect(t.Context(), cwd, true)
	if err == nil {
		t.Fatal("expected error")
	}
	if !strings.Contains(err.Error(), "--require-git") {
		t.Fatalf("expected require-git guidance, got %v", err)
	}
}
