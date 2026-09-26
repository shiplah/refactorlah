package scan

import (
	"errors"
	"os"
	"path/filepath"
	"reflect"
	"testing"

	"github.com/shiplah/refactorlah/internal/config"
	"github.com/shiplah/refactorlah/internal/testfixtures"
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

func TestIndexReportsUniqueCandidateSizesWithoutExcludedFiles(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	files := map[string]string{
		"src/Moved.php": "<?php class Old {}",
		"storage/quality/phpstan/resultCache.php":  "<?php /* Old */ with a longer result",
		"storage/quality/phpstan/ignoredCache.php": "<?php /* Old */ with the longest ignored result",
	}
	for file, content := range files {
		absolute := filepath.Join(root, filepath.FromSlash(file))
		if err := os.MkdirAll(filepath.Dir(absolute), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(absolute, []byte(content), 0o644); err != nil {
			t.Fatal(err)
		}
	}

	index := NewIndex(root, config.Config{Exclude: []string{"storage/quality/phpstan/ignoredCache.php"}})
	query := CandidateQuery{
		Extensions:   []string{".php"},
		Needles:      []string{"Old"},
		IncludePaths: []string{"src/Moved.php"},
	}
	for range 2 {
		if _, err := index.CandidateFiles(root, query); err != nil {
			t.Fatal(err)
		}
	}

	stats := index.CandidateStats()
	if stats.Files != 2 || stats.LargestFile != "storage/quality/phpstan/resultCache.php" {
		t.Fatalf("unexpected candidate stats: %#v", stats)
	}
	if stats.LargestBytes != int64(len(files[stats.LargestFile])) {
		t.Fatalf("unexpected largest candidate size: %#v", stats)
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

func TestIndexSelectsCandidateFilesByNeedlesAndIncludes(t *testing.T) {
	t.Parallel()

	root := testfixtures.CopyDir(t, "tests/fixtures/scan-candidate-files")

	index := NewIndex(root, config.Config{})
	files, err := index.CandidateFiles(root, CandidateQuery{
		Extensions:   []string{".php"},
		Needles:      []string{"App\\Old\\Moved"},
		IncludePaths: []string{"src/Moved.php"},
	})
	if err != nil {
		t.Fatal(err)
	}

	expected := []string{"src/Moved.php", "src/UsesMoved.php"}
	if !reflect.DeepEqual(files, expected) {
		t.Fatalf("expected %#v, got %#v", expected, files)
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
