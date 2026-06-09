package golang

import (
	"os"
	"path/filepath"
	"sort"
	"strings"
	"testing"

	adapterproto "github.com/shiplah/refactorlah/internal/adapters/contract"
	"github.com/shiplah/refactorlah/internal/config"
	"github.com/shiplah/refactorlah/internal/planning"
)

func TestGoCandidateQueryIncludesMovedFilesAndReferenceNeedles(t *testing.T) {
	query := goCandidateQuery([]packageMoveMapping{{
		OldImport:  "example.com/project/internal/oldpkg",
		OldPackage: "oldpkg",
		FilePackages: []filePackageMapping{{
			OldPath: "internal/oldpkg/service.go",
		}},
	}}, []symbolMoveMapping{{
		OldPath:    "internal/models/old_thing.go",
		OldImport:  "example.com/project/internal/models",
		OldPackage: "models",
		OldSymbol:  "OldThing",
	}})

	for _, expected := range []string{"internal/oldpkg/service.go", "internal/models/old_thing.go"} {
		if !containsGoString(query.IncludePaths, expected) {
			t.Fatalf("expected include path %q in %#v", expected, query.IncludePaths)
		}
	}
	for _, expected := range []string{"example.com/project/internal/oldpkg", "oldpkg", "example.com/project/internal/models", "models", "OldThing"} {
		if !containsGoString(query.Needles, expected) {
			t.Fatalf("expected needle %q in %#v", expected, query.Needles)
		}
	}
}

func TestAnalyzerUpdatesGoImportPathsForMovedPackageDirectory(t *testing.T) {
	root := t.TempDir()
	writeFile(t, root, "go.mod", "module example.com/project\n")
	writeFile(t, root, "internal/adapters/php/parser.go", `package php

import "example.com/project/internal/parsing/treesitter"
`)
	writeFile(t, root, "internal/parsing/treesitter/document.go", "package treesitter\n")

	response, relevant, err := analyzeGo(t, root, planning.MovePlan{
		Moves: []planning.FileMove{{
			OldPath: "internal/parsing/treesitter/document.go",
			NewPath: "internal/parsing/document/document.go",
		}},
	})
	if err != nil {
		t.Fatalf("analyze go imports: %v", err)
	}
	if !relevant {
		t.Fatal("expected go analyzer to be relevant")
	}
	replacement, found := findReplacement(response.Replacements, "internal/adapters/php/parser.go", "go-import-path")
	if !found {
		t.Fatalf("expected parser import replacement, got %#v", response.Replacements)
	}
	if replacement.File != "internal/adapters/php/parser.go" {
		t.Fatalf("expected parser.go replacement, got %q", replacement.File)
	}
	if replacement.Replacement != "example.com/project/internal/parsing/document" {
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

	response, relevant, err := analyzeGo(t, root, planning.MovePlan{
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

func TestAnalyzerUpdatesGoSymbolRenameFromFileBasename(t *testing.T) {
	root := t.TempDir()
	writeFile(t, root, "go.mod", "module example.com/project\n")
	oldSource := `package models

type OldThing struct{}

func (thing OldThing) Clone() OldThing {
	return OldThing{}
}
`
	samePackageConsumer := `package models

func Build(value OldThing) OldThing {
	return OldThing{}
}
`
	externalConsumer := `package consumer

import "example.com/project/internal/models"

func Build() models.OldThing {
	return models.OldThing{}
}
`
	writeFile(t, root, "internal/models/old_thing.go", oldSource)
	writeFile(t, root, "internal/models/use.go", samePackageConsumer)
	writeFile(t, root, "internal/consumer/use.go", externalConsumer)

	response, relevant, err := analyzeGo(t, root, planning.MovePlan{
		Moves: []planning.FileMove{{
			OldPath: "internal/models/old_thing.go",
			NewPath: "internal/models/new_thing.go",
		}},
	})
	if err != nil {
		t.Fatalf("analyze go symbol rename: %v", err)
	}
	if !relevant {
		t.Fatal("expected go analyzer to be relevant")
	}

	updatedSource := applyGoReplacements(oldSource, response.Replacements, "internal/models/old_thing.go")
	for _, expected := range []string{"type NewThing struct{}", "func (thing NewThing) Clone() NewThing", "return NewThing{}"} {
		if !strings.Contains(updatedSource, expected) {
			t.Fatalf("expected %q in moved source, got:\n%s", expected, updatedSource)
		}
	}

	updatedSamePackageConsumer := applyGoReplacements(samePackageConsumer, response.Replacements, "internal/models/use.go")
	for _, expected := range []string{"func Build(value NewThing) NewThing", "return NewThing{}"} {
		if !strings.Contains(updatedSamePackageConsumer, expected) {
			t.Fatalf("expected %q in same-package consumer, got:\n%s", expected, updatedSamePackageConsumer)
		}
	}

	updatedExternalConsumer := applyGoReplacements(externalConsumer, response.Replacements, "internal/consumer/use.go")
	for _, expected := range []string{"models.NewThing", "return models.NewThing{}"} {
		if !strings.Contains(updatedExternalConsumer, expected) {
			t.Fatalf("expected %q in external consumer, got:\n%s", expected, updatedExternalConsumer)
		}
	}
	if !hasGoSymbolMapping(response.SymbolMappings, "go-type", "example.com/project/internal/models.OldThing", "example.com/project/internal/models.NewThing") {
		t.Fatalf("expected go type symbol mapping, got %#v", response.SymbolMappings)
	}
}

func TestAnalyzerDoesNotRewriteExcludedGoLocalReferences(t *testing.T) {
	root := t.TempDir()
	writeFile(t, root, "go.mod", "module example.com/project\n")
	oldSource := `package models

type OldThing struct{}
`
	excludedConsumer := `package models

func Build(value OldThing) OldThing {
	return OldThing{}
}
`
	writeFile(t, root, "internal/models/old_thing.go", oldSource)
	writeFile(t, root, "internal/models/use.go", excludedConsumer)

	response, _, err := analyzeGoWithConfig(t, root, planning.MovePlan{
		Moves: []planning.FileMove{{
			OldPath: "internal/models/old_thing.go",
			NewPath: "internal/models/new_thing.go",
		}},
	}, config.Config{Exclude: []string{"internal/models/use.go"}})
	if err != nil {
		t.Fatalf("analyze go symbol rename: %v", err)
	}

	updatedSource := applyGoReplacements(oldSource, response.Replacements, "internal/models/old_thing.go")
	if !strings.Contains(updatedSource, "type NewThing struct{}") {
		t.Fatalf("expected moved source declaration replacement, got:\n%s", updatedSource)
	}
	if _, found := findReplacement(response.Replacements, "internal/models/use.go", "go-local-symbol-reference"); found {
		t.Fatalf("did not expect replacement in excluded Go file, got %#v", response.Replacements)
	}
}

func TestAnalyzerEmitsConsequenceWarningForSkippedGoCandidatesWithoutParserDetails(t *testing.T) {
	root := t.TempDir()
	writeFile(t, root, "go.mod", "module example.com/project\n")
	writeFile(t, root, "internal/oldpkg/service.go", `package oldpkg

func Build() {}
`)
	writeFile(t, root, "internal/consumer/use.go", `package consumer

import "example.com/project/internal/oldpkg"

func Use( {
	oldpkg.Build()
}
`)

	response, _, err := analyzeGo(t, root, planning.MovePlan{
		Moves: []planning.FileMove{{
			OldPath: "internal/oldpkg/service.go",
			NewPath: "internal/newpkg/service.go",
		}},
	})
	if err != nil {
		t.Fatalf("analyze go package: %v", err)
	}

	assertGoWarning(t, response.Warnings, "internal/consumer/use.go", "This file could not be checked for Go package qualifier changes; matching references may be unchanged.")
	assertNoGoWarningContains(t, response.Warnings, "parsed")
	assertNoGoWarningContains(t, response.Warnings, "not analysed")
}

func TestAnalyzerUpdatesGoFunctionRenameFromFileBasename(t *testing.T) {
	root := t.TempDir()
	writeFile(t, root, "go.mod", "module example.com/project\n")
	oldSource := `package tasks

func oldThing() int {
	return 1
}
`
	consumer := `package tasks

func Use() int {
	return oldThing()
}
`
	writeFile(t, root, "internal/tasks/old_thing.go", oldSource)
	writeFile(t, root, "internal/tasks/use.go", consumer)

	response, _, err := analyzeGo(t, root, planning.MovePlan{
		Moves: []planning.FileMove{{
			OldPath: "internal/tasks/old_thing.go",
			NewPath: "internal/tasks/new_thing.go",
		}},
	})
	if err != nil {
		t.Fatalf("analyze go function rename: %v", err)
	}

	updatedSource := applyGoReplacements(oldSource, response.Replacements, "internal/tasks/old_thing.go")
	if !strings.Contains(updatedSource, "func newThing() int") {
		t.Fatalf("expected function declaration rename, got:\n%s", updatedSource)
	}
	updatedConsumer := applyGoReplacements(consumer, response.Replacements, "internal/tasks/use.go")
	if !strings.Contains(updatedConsumer, "return newThing()") {
		t.Fatalf("expected function reference rename, got:\n%s", updatedConsumer)
	}
}

func TestAnalyzerUpdatesGoPackageAndSymbolRenameTogether(t *testing.T) {
	root := t.TempDir()
	writeFile(t, root, "go.mod", "module example.com/project\n")
	oldSource := `package oldpkg

type OldThing struct{}
`
	consumer := `package consumer

import "example.com/project/internal/oldpkg"

func Build() oldpkg.OldThing {
	return oldpkg.OldThing{}
}
`
	writeFile(t, root, "internal/oldpkg/old_thing.go", oldSource)
	writeFile(t, root, "internal/consumer/use.go", consumer)

	response, _, err := analyzeGo(t, root, planning.MovePlan{
		Moves: []planning.FileMove{{
			OldPath: "internal/oldpkg/old_thing.go",
			NewPath: "internal/newpkg/new_thing.go",
		}},
	})
	if err != nil {
		t.Fatalf("analyze go package and symbol rename: %v", err)
	}

	updatedSource := applyGoReplacements(oldSource, response.Replacements, "internal/oldpkg/old_thing.go")
	for _, expected := range []string{"package newpkg", "type NewThing struct{}"} {
		if !strings.Contains(updatedSource, expected) {
			t.Fatalf("expected %q in moved source, got:\n%s", expected, updatedSource)
		}
	}

	updatedConsumer := applyGoReplacements(consumer, response.Replacements, "internal/consumer/use.go")
	for _, expected := range []string{`"example.com/project/internal/newpkg"`, "newpkg.NewThing"} {
		if !strings.Contains(updatedConsumer, expected) {
			t.Fatalf("expected %q in consumer, got:\n%s", expected, updatedConsumer)
		}
	}
}

func TestAnalyzerSkipsGoSymbolRenameWhenDeclarationDoesNotMatchBasename(t *testing.T) {
	root := t.TempDir()
	writeFile(t, root, "go.mod", "module example.com/project\n")
	writeFile(t, root, "internal/models/old_thing.go", `package models

type Service struct{}
`)

	response, _, err := analyzeGo(t, root, planning.MovePlan{
		Moves: []planning.FileMove{{
			OldPath: "internal/models/old_thing.go",
			NewPath: "internal/models/new_thing.go",
		}},
	})
	if err != nil {
		t.Fatalf("analyze go symbol skip: %v", err)
	}
	if len(response.Replacements) != 0 {
		t.Fatalf("expected no replacements when declaration does not match basename, got %#v", response.Replacements)
	}
	if len(response.SymbolMappings) != 0 {
		t.Fatalf("expected no symbol mapping, got %#v", response.SymbolMappings)
	}
}

func TestAnalyzerWarnsAndSkipsGoSymbolRenameWhenTargetSymbolExists(t *testing.T) {
	root := t.TempDir()
	writeFile(t, root, "go.mod", "module example.com/project\n")
	writeFile(t, root, "internal/models/old_thing.go", `package models

type OldThing struct{}
`)
	writeFile(t, root, "internal/models/existing.go", `package models

type NewThing struct{}
`)

	response, _, err := analyzeGo(t, root, planning.MovePlan{
		Moves: []planning.FileMove{{
			OldPath: "internal/models/old_thing.go",
			NewPath: "internal/models/new_thing.go",
		}},
	})
	if err != nil {
		t.Fatalf("analyze go symbol conflict: %v", err)
	}
	if len(response.Replacements) != 0 {
		t.Fatalf("expected no replacements when target symbol exists, got %#v", response.Replacements)
	}
	if len(response.Warnings) != 1 {
		t.Fatalf("expected one warning, got %#v", response.Warnings)
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

	response, _, err := analyzeGo(t, root, planning.MovePlan{
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

	response, relevant, err := analyzeGo(t, root, planning.MovePlan{
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

	response, relevant, err := analyzeGo(t, root, planning.MovePlan{
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

	response, relevant, err := analyzeGo(t, root, planning.MovePlan{
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

	_, relevant, err := analyzeGo(t, root, planning.MovePlan{
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

func findReplacement(replacements []adapterproto.Replacement, file string, reason string) (adapterproto.Replacement, bool) {
	for _, replacement := range replacements {
		if replacement.File == file && replacement.Reason == reason {
			return replacement, true
		}
	}
	return adapterproto.Replacement{}, false
}

func containsGoString(values []string, expected string) bool {
	for _, value := range values {
		if value == expected {
			return true
		}
	}
	return false
}

func hasGoSymbolMapping(mappings []adapterproto.SymbolMapping, kind string, oldSymbol string, newSymbol string) bool {
	for _, mapping := range mappings {
		if mapping.Kind == kind && mapping.OldSymbol == oldSymbol && mapping.NewSymbol == newSymbol {
			return true
		}
	}
	return false
}

func assertNoGoWarningContains(t *testing.T, warnings []adapterproto.Warning, needle string) {
	t.Helper()

	for _, warning := range warnings {
		if strings.Contains(warning.Message, needle) {
			t.Fatalf("did not expect warning containing %q, got %#v", needle, warning)
		}
	}
}

func assertGoWarning(t *testing.T, warnings []adapterproto.Warning, file string, message string) {
	t.Helper()

	for _, warning := range warnings {
		if warning.File == file && warning.Message == message {
			return
		}
	}
	t.Fatalf("expected warning in %s: %s, got %#v", file, message, warnings)
}
