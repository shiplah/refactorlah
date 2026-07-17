//go:build cgo

package php

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/shiplah/refactorlah/internal/planning"
)

func TestReadComposerAutoloadFilesReturnsProjectRelativePaths(t *testing.T) {
	root := t.TempDir()
	writeComposerFile(t, root, "platform/composer.json", `{
  "autoload": {
    "files": ["src/Config/symbols.php"]
  },
  "autoload-dev": {
    "files": ["tests/bootstrap.php"]
  }
}`)

	files, err := ReadComposerAutoloadFiles(root, filepath.Join(root, "platform"))
	if err != nil {
		t.Fatalf("read composer autoload files: %v", err)
	}

	for _, expected := range []string{
		"platform/src/Config/symbols.php",
		"platform/tests/bootstrap.php",
	} {
		if !files[expected] {
			t.Fatalf("expected autoload file %q, got %#v", expected, files)
		}
	}
}

func TestCollectComposerAutoloadFileReplacementsMovesFileEntries(t *testing.T) {
	root := t.TempDir()
	writeComposerFile(t, root, "composer.json", `{
  "autoload": {
    "psr-4": {
      "App\\": "src/"
    },
    "files": [
      "src/Config/symbols.php"
    ]
  }
}`)

	replacements, err := CollectComposerAutoloadFileReplacements(root, root, planning.MovePlan{
		Moves: []planning.FileMove{{
			OldPath: "src/Config/symbols.php",
			NewPath: "src/Shared/symbols.php",
		}},
	})
	if err != nil {
		t.Fatalf("collect composer autoload file replacements: %v", err)
	}
	if len(replacements) != 1 {
		t.Fatalf("expected one replacement, got %#v", replacements)
	}
	if replacements[0].File != "composer.json" || replacements[0].Replacement != "src/Shared/symbols.php" {
		t.Fatalf("unexpected replacement: %#v", replacements[0])
	}
}

func writeComposerFile(t *testing.T, root string, relativePath string, content string) {
	t.Helper()

	absolutePath := filepath.Join(root, filepath.FromSlash(relativePath))
	if err := os.MkdirAll(filepath.Dir(absolutePath), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(absolutePath, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
}
