package files

import (
	"os"
	"path/filepath"
	"testing"
)

func TestIsIgnoredPathIncludesPythonGeneratedDirectories(t *testing.T) {
	for _, path := range []string{
		".venv/lib/site-packages/example.py",
		"src/app/__pycache__/module.py",
		"nested/.venv/bin/python",
		"nested/__pycache__/module.pyc",
	} {
		if !IsIgnoredPath(path) {
			t.Fatalf("expected %s to be ignored", path)
		}
	}
}

func TestCollectFilesPrunesIgnoredPythonGeneratedDirectories(t *testing.T) {
	root := t.TempDir()
	writeCollectedFile(t, root, "src/app/service.py")
	writeCollectedFile(t, root, "node_modules/package/ignored.py")
	writeCollectedFile(t, root, "src/app/__pycache__/ignored.py")
	writeCollectedFile(t, root, ".venv/lib/ignored.py")

	collected, err := CollectFiles(root, ".")
	if err != nil {
		t.Fatal(err)
	}

	if len(collected) != 1 || collected[0] != "src/app/service.py" {
		t.Fatalf("expected only service.py, got %#v", collected)
	}
}

func writeCollectedFile(t *testing.T, root string, relativePath string) {
	t.Helper()

	absolutePath := filepath.Join(root, filepath.FromSlash(relativePath))
	if err := os.MkdirAll(filepath.Dir(absolutePath), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(absolutePath, []byte(""), 0o644); err != nil {
		t.Fatal(err)
	}
}
