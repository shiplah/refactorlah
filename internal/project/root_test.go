package project

import (
	"context"
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
