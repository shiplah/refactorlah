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

func TestRootDetectorFallsBackToCurrentDirectoryWithoutGitOrMarkers(t *testing.T) {
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

func TestRootDetectorPrefersGitRootOverProjectMarkers(t *testing.T) {
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

func TestRootDetectorUsesConfiguredRootBeforeNestedPackageRoot(t *testing.T) {
	root := t.TempDir()
	cwd := filepath.Join(root, "platform", "src")
	if err := os.MkdirAll(cwd, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(root, ".refactorlah.json"), []byte("{}"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(root, "platform", "composer.json"), []byte("{}"), 0o644); err != nil {
		t.Fatal(err)
	}

	detector := NewRootDetector(stubGitRootDetector{})
	info, err := detector.Detect(t.Context(), cwd)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if info.ProjectRoot != root {
		t.Fatalf("expected configured root %q, got %q", root, info.ProjectRoot)
	}
}

func TestRootDetectorSupportsPackageJsonRootWithoutComposer(t *testing.T) {
	root := t.TempDir()
	cwd := filepath.Join(root, "src")
	if err := os.MkdirAll(cwd, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(root, "package.json"), []byte("{}"), 0o644); err != nil {
		t.Fatal(err)
	}

	detector := NewRootDetector(stubGitRootDetector{})
	info, err := detector.Detect(t.Context(), cwd)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if info.ProjectRoot != root {
		t.Fatalf("expected package root %q, got %q", root, info.ProjectRoot)
	}
}

func TestRootDetectorSupportsPyprojectRootWithoutComposer(t *testing.T) {
	root := t.TempDir()
	cwd := filepath.Join(root, "src")
	if err := os.MkdirAll(cwd, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(root, "pyproject.toml"), []byte("[project]\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	detector := NewRootDetector(stubGitRootDetector{})
	info, err := detector.Detect(t.Context(), cwd)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if info.ProjectRoot != root {
		t.Fatalf("expected pyproject root %q, got %q", root, info.ProjectRoot)
	}
}
