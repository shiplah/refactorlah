package golang

import (
	"os"
	"path/filepath"
	"testing"

	"refactorlah/internal/planning"
)

func TestAnalyzerUpdatesGoImportPathsForMovedPackageDirectory(t *testing.T) {
	root := t.TempDir()
	writeFile(t, root, "go.mod", "module example.com/project\n")
	writeFile(t, root, "internal/languages/php/parser.go", `package php

import "example.com/project/internal/languages/treesitter"
`)
	writeFile(t, root, "internal/languages/treesitter/document.go", "package treesitter\n")

	response, relevant, err := NewAnalyzer().Analyze(root, planning.MovePlan{
		Moves: []planning.FileMove{{
			OldPath: "internal/languages/treesitter/document.go",
			NewPath: "internal/parsing/treesitter/document.go",
		}},
	})
	if err != nil {
		t.Fatalf("analyze go imports: %v", err)
	}
	if !relevant {
		t.Fatal("expected go analyzer to be relevant")
	}
	if len(response.Replacements) != 1 {
		t.Fatalf("expected 1 replacement, got %#v", response.Replacements)
	}

	replacement := response.Replacements[0]
	if replacement.File != "internal/languages/php/parser.go" {
		t.Fatalf("expected parser.go replacement, got %q", replacement.File)
	}
	if replacement.Replacement != "example.com/project/internal/parsing/treesitter" {
		t.Fatalf("expected new import path, got %q", replacement.Replacement)
	}
	if len(response.PathMappings) != 1 {
		t.Fatalf("expected 1 path mapping, got %#v", response.PathMappings)
	}
}

func TestAnalyzerIgnoresNonGoMoves(t *testing.T) {
	root := t.TempDir()
	writeFile(t, root, "go.mod", "module example.com/project\n")

	_, relevant, err := NewAnalyzer().Analyze(root, planning.MovePlan{
		Moves: []planning.FileMove{{
			OldPath: "README.md",
			NewPath: "docs/README.md",
		}},
	})
	if err != nil {
		t.Fatalf("analyze go imports: %v", err)
	}
	if relevant {
		t.Fatal("expected non-Go move to be ignored")
	}
}

func writeFile(t *testing.T, root string, relativePath string, content string) {
	t.Helper()

	absolutePath := filepath.Join(root, filepath.FromSlash(relativePath))
	if err := os.MkdirAll(filepath.Dir(absolutePath), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(absolutePath, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
}
