package project

import (
	"os"
	"path/filepath"
	"testing"
)

func TestFindGoRootForPaths(t *testing.T) {
	root := t.TempDir()
	if err := os.WriteFile(filepath.Join(root, "go.mod"), []byte("module example.com/app\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	goRoot, found, err := FindGoRootForPaths(root, []string{"internal/old/file.go", "internal/new/file.go"})
	if err != nil {
		t.Fatalf("find go root: %v", err)
	}
	if !found {
		t.Fatal("expected go root")
	}
	if goRoot != root {
		t.Fatalf("expected %s, got %s", root, goRoot)
	}
}

func TestReadGoModulePath(t *testing.T) {
	root := t.TempDir()
	if err := os.WriteFile(filepath.Join(root, "go.mod"), []byte("// comment\nmodule example.com/app\n\ngo 1.26\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	modulePath, err := ReadGoModulePath(root)
	if err != nil {
		t.Fatalf("read go module path: %v", err)
	}
	if modulePath != "example.com/app" {
		t.Fatalf("expected module path, got %q", modulePath)
	}
}
