package project

import (
	"os"
	"path/filepath"
	"testing"
)

func TestFindPythonRootForPathsReturnsNestedProject(t *testing.T) {
	root := t.TempDir()
	packageDir := filepath.Join(root, "packages", "billing")
	if err := os.MkdirAll(packageDir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(packageDir, "pyproject.toml"), []byte("[tool.ruff]\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	pythonRoot, found, err := FindPythonRootForPaths(root, []string{
		"packages/billing/src/app/services/billing.py",
		"packages/billing/src/app/domain/billing.py",
	})
	if err != nil {
		t.Fatalf("find python root failed: %v", err)
	}
	if !found {
		t.Fatal("expected python root to be found")
	}
	if pythonRoot != packageDir {
		t.Fatalf("expected %s, got %s", packageDir, pythonRoot)
	}
}
