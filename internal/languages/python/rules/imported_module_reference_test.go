//go:build cgo

package rules_test

import (
	"testing"

	"refactorlah/internal/languages/python"
	"refactorlah/internal/languages/python/rules"
)

func TestImportedModuleReferenceRuleUpdatesUnaliasedImportedModuleUsage(t *testing.T) {
	source := []byte(`from collector.assembly.cache_files import snapshot_manifest

manifest = snapshot_manifest.load()
`)
	document, err := python.Parse(source)
	if err != nil {
		t.Fatalf("parse python: %v", err)
	}
	defer document.Close()

	replacements := rules.ImportedModuleReferenceRule{}.Collect(document, rules.ImportedModuleReferenceInput{
		File:      "src/collector/assembly/cache_files/loader.py",
		Package:   "collector.assembly.cache_files",
		Source:    source,
		OldModule: "collector.assembly.cache_files.snapshot_manifest",
		NewModule: "collector.assembly.cache_files.summary_manifest",
	})

	if len(replacements) != 1 {
		t.Fatalf("expected 1 replacement, got %#v", replacements)
	}
	replacement := replacements[0]
	if string(source[replacement.Start:replacement.End]) != "snapshot_manifest" {
		t.Fatalf("replacement range points to %q", string(source[replacement.Start:replacement.End]))
	}
	if replacement.Replacement != "summary_manifest" {
		t.Fatalf("expected summary_manifest replacement, got %q", replacement.Replacement)
	}
}

func TestImportedModuleReferenceRuleSupportsRelativeImports(t *testing.T) {
	source := []byte("from . import snapshot_manifest\n\nvalue = snapshot_manifest.load()\n")
	document, err := python.Parse(source)
	if err != nil {
		t.Fatalf("parse python: %v", err)
	}
	defer document.Close()

	replacements := rules.ImportedModuleReferenceRule{}.Collect(document, rules.ImportedModuleReferenceInput{
		File:      "src/collector/assembly/cache_files/loader.py",
		Package:   "collector.assembly.cache_files",
		Source:    source,
		OldModule: "collector.assembly.cache_files.snapshot_manifest",
		NewModule: "collector.assembly.cache_files.summary_manifest",
	})

	if len(replacements) != 1 {
		t.Fatalf("expected 1 replacement, got %#v", replacements)
	}
}

func TestImportedModuleReferenceRuleSkipsAliasedImports(t *testing.T) {
	source := []byte("from collector.assembly.cache_files import snapshot_manifest as manifest\n\nvalue = manifest.load()\n")
	document, err := python.Parse(source)
	if err != nil {
		t.Fatalf("parse python: %v", err)
	}
	defer document.Close()

	replacements := rules.ImportedModuleReferenceRule{}.Collect(document, rules.ImportedModuleReferenceInput{
		File:      "src/collector/assembly/cache_files/loader.py",
		Package:   "collector.assembly.cache_files",
		Source:    source,
		OldModule: "collector.assembly.cache_files.snapshot_manifest",
		NewModule: "collector.assembly.cache_files.summary_manifest",
	})

	if len(replacements) != 0 {
		t.Fatalf("expected no replacements, got %#v", replacements)
	}
}
