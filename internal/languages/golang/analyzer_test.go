package golang

import (
	"os"
	"path/filepath"
	"sort"
	"strings"
	"testing"

	adapterproto "refactorlah/internal/adapters"
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

func TestAnalyzerUpdatesPackageDeclarationsAndQualifiersForRenamedPackageDirectory(t *testing.T) {
	root := t.TempDir()
	writeFile(t, root, "go.mod", "module example.com/project\n")
	oldSource := `package oldpkg

func Build() {}
`
	consumer := `package consumer

import "example.com/project/internal/oldpkg"

func Use() {
	oldpkg.Build()
}
`
	writeFile(t, root, "internal/oldpkg/service.go", oldSource)
	writeFile(t, root, "internal/consumer/use.go", consumer)

	response, relevant, err := NewAnalyzer().Analyze(root, planning.MovePlan{
		Moves: []planning.FileMove{{
			OldPath: "internal/oldpkg/service.go",
			NewPath: "internal/newpkg/service.go",
		}},
	})
	if err != nil {
		t.Fatalf("analyze go package: %v", err)
	}
	if !relevant {
		t.Fatal("expected go analyzer to be relevant")
	}

	updatedSource := applyGoReplacements(oldSource, response.Replacements, "internal/oldpkg/service.go")
	if !strings.Contains(updatedSource, "package newpkg") {
		t.Fatalf("expected moved package declaration to change, got:\n%s", updatedSource)
	}

	updatedConsumer := applyGoReplacements(consumer, response.Replacements, "internal/consumer/use.go")
	if !strings.Contains(updatedConsumer, `"example.com/project/internal/newpkg"`) {
		t.Fatalf("expected import path rewrite, got:\n%s", updatedConsumer)
	}
	if !strings.Contains(updatedConsumer, "newpkg.Build()") {
		t.Fatalf("expected package qualifier rewrite, got:\n%s", updatedConsumer)
	}
	if len(response.SymbolMappings) != 1 {
		t.Fatalf("expected package symbol mapping, got %#v", response.SymbolMappings)
	}
}

func TestAnalyzerPreservesCustomPackageNames(t *testing.T) {
	root := t.TempDir()
	writeFile(t, root, "go.mod", "module example.com/project\n")
	oldSource := `package custom

func Build() {}
`
	consumer := `package consumer

import "example.com/project/internal/oldpkg"

func Use() {
	custom.Build()
}
`
	writeFile(t, root, "internal/oldpkg/service.go", oldSource)
	writeFile(t, root, "internal/consumer/use.go", consumer)

	response, _, err := NewAnalyzer().Analyze(root, planning.MovePlan{
		Moves: []planning.FileMove{{
			OldPath: "internal/oldpkg/service.go",
			NewPath: "internal/newpkg/service.go",
		}},
	})
	if err != nil {
		t.Fatalf("analyze go package: %v", err)
	}

	updatedSource := applyGoReplacements(oldSource, response.Replacements, "internal/oldpkg/service.go")
	if strings.Contains(updatedSource, "package newpkg") {
		t.Fatalf("did not expect custom package name rewrite, got:\n%s", updatedSource)
	}

	updatedConsumer := applyGoReplacements(consumer, response.Replacements, "internal/consumer/use.go")
	if !strings.Contains(updatedConsumer, `"example.com/project/internal/newpkg"`) {
		t.Fatalf("expected import path rewrite, got:\n%s", updatedConsumer)
	}
	if !strings.Contains(updatedConsumer, "custom.Build()") {
		t.Fatalf("expected custom qualifier to remain, got:\n%s", updatedConsumer)
	}
}

func TestAnalyzerUpdatesExternalTestPackageDeclaration(t *testing.T) {
	root := t.TempDir()
	writeFile(t, root, "go.mod", "module example.com/project\n")
	oldSource := `package oldpkg_test

func TestBuild() {}
`
	writeFile(t, root, "internal/oldpkg/service_test.go", oldSource)

	response, relevant, err := NewAnalyzer().Analyze(root, planning.MovePlan{
		Moves: []planning.FileMove{{
			OldPath: "internal/oldpkg/service_test.go",
			NewPath: "internal/newpkg/service_test.go",
		}},
	})
	if err != nil {
		t.Fatalf("analyze go package: %v", err)
	}
	if !relevant {
		t.Fatal("expected go analyzer to be relevant")
	}

	updatedSource := applyGoReplacements(oldSource, response.Replacements, "internal/oldpkg/service_test.go")
	if !strings.Contains(updatedSource, "package newpkg_test") {
		t.Fatalf("expected external test package declaration to change, got:\n%s", updatedSource)
	}
}

func TestAnalyzerWarnsAndSkipsSemanticRewritesForPartialPackageMoves(t *testing.T) {
	root := t.TempDir()
	writeFile(t, root, "go.mod", "module example.com/project\n")
	writeFile(t, root, "internal/oldpkg/a.go", "package oldpkg\n")
	writeFile(t, root, "internal/oldpkg/b.go", "package oldpkg\n")
	writeFile(t, root, "internal/consumer/use.go", `package consumer

import "example.com/project/internal/oldpkg"
`)

	response, relevant, err := NewAnalyzer().Analyze(root, planning.MovePlan{
		Moves: []planning.FileMove{{
			OldPath: "internal/oldpkg/a.go",
			NewPath: "internal/newpkg/a.go",
		}},
	})
	if err != nil {
		t.Fatalf("analyze go package: %v", err)
	}
	if !relevant {
		t.Fatal("expected go analyzer to be relevant")
	}
	if len(response.Replacements) != 0 {
		t.Fatalf("expected no semantic replacements for partial package move, got %#v", response.Replacements)
	}
	if len(response.Warnings) != 1 {
		t.Fatalf("expected one warning, got %#v", response.Warnings)
	}
}

func TestAnalyzerWarnsAndSkipsAmbiguousGoPackageNames(t *testing.T) {
	root := t.TempDir()
	writeFile(t, root, "go.mod", "module example.com/project\n")
	writeFile(t, root, "internal/oldpkg/a.go", "package oldpkg\n")
	writeFile(t, root, "internal/oldpkg/b.go", "package alternate\n")
	writeFile(t, root, "internal/consumer/use.go", `package consumer

import "example.com/project/internal/oldpkg"
`)

	response, relevant, err := NewAnalyzer().Analyze(root, planning.MovePlan{
		Moves: []planning.FileMove{
			{OldPath: "internal/oldpkg/a.go", NewPath: "internal/newpkg/a.go"},
			{OldPath: "internal/oldpkg/b.go", NewPath: "internal/newpkg/b.go"},
		},
	})
	if err != nil {
		t.Fatalf("analyze go package: %v", err)
	}
	if !relevant {
		t.Fatal("expected go analyzer to be relevant")
	}
	if len(response.Replacements) != 0 {
		t.Fatalf("expected no replacements for ambiguous packages, got %#v", response.Replacements)
	}
	if len(response.Warnings) != 1 {
		t.Fatalf("expected one warning, got %#v", response.Warnings)
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

func applyGoReplacements(content string, replacements []adapterproto.Replacement, file string) string {
	fileReplacements := make([]adapterproto.Replacement, 0, len(replacements))
	for _, replacement := range replacements {
		if replacement.File == file {
			fileReplacements = append(fileReplacements, replacement)
		}
	}
	sort.Slice(fileReplacements, func(left int, right int) bool {
		return fileReplacements[left].Start > fileReplacements[right].Start
	})

	result := []byte(content)
	for _, replacement := range fileReplacements {
		next := make([]byte, 0, len(result)-replacement.End+replacement.Start+len(replacement.Replacement))
		next = append(next, result[:replacement.Start]...)
		next = append(next, []byte(replacement.Replacement)...)
		next = append(next, result[replacement.End:]...)
		result = next
	}
	return string(result)
}
