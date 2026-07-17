//go:build cgo

package php

import (
	"path/filepath"
	"testing"

	"github.com/shiplah/refactorlah/internal/planning"
	"github.com/shiplah/refactorlah/internal/testfixtures"
)

func TestReadComposerAutoloadFilesReturnsProjectRelativePaths(t *testing.T) {
	root := testfixtures.CopyDir(t, "tests/fixtures/php-unqualified-symbols/composer-files")

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
	root := testfixtures.CopyDir(t, "tests/fixtures/php-unqualified-symbols/before")

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
