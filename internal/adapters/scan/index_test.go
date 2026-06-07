package scan

import (
	"errors"
	"path/filepath"
	"reflect"
	"testing"

	"refactorlah/internal/config"
)

func TestIndexFiltersFilesByRootExtensionAndConfig(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	collector := func(root string, relativePath string) ([]string, error) {
		if relativePath != "." {
			t.Fatalf("expected relative path '.', got %q", relativePath)
		}
		return []string{
			"src/App.php",
			"src/Generated.php",
			"src/Controller.go",
			"README.md",
		}, nil
	}

	index := newIndex(root, config.Config{
		Exclude: []string{"platform/src/Generated.php"},
	}, collector)

	files, err := index.Files(filepath.Join(root, "platform"), ".php")
	if err != nil {
		t.Fatal(err)
	}

	expected := []string{"platform/src/App.php"}
	if !reflect.DeepEqual(files, expected) {
		t.Fatalf("unexpected files: %#v", files)
	}
}

func TestIndexCachesRootWalks(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	calls := 0
	index := newIndex(root, config.Config{}, func(root string, relativePath string) ([]string, error) {
		calls++
		return []string{"src/App.php", "src/app.py"}, nil
	})

	if _, err := index.Files(root, ".php"); err != nil {
		t.Fatal(err)
	}
	if _, err := index.Files(root, ".py"); err != nil {
		t.Fatal(err)
	}
	if calls != 1 {
		t.Fatalf("expected one walk for cached root, got %d", calls)
	}
}

func TestIndexRejectsRootsOutsideProject(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	index := newIndex(root, config.Config{}, func(root string, relativePath string) ([]string, error) {
		return nil, errors.New("collector should not be called")
	})

	if _, err := index.Files(filepath.Dir(root)); err == nil {
		t.Fatal("expected outside root to fail")
	}
}
