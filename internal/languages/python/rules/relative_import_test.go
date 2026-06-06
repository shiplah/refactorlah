//go:build cgo

package rules_test

import (
	"testing"

	"refactorlah/internal/languages/python"
	"refactorlah/internal/languages/python/rules"
)

func TestRelativeImportRuleUpdatesExplicitRelativeModule(t *testing.T) {
	source := []byte("from .snapshot_manifest import SnapshotManifest\n")
	document, err := python.Parse(source)
	if err != nil {
		t.Fatalf("parse python: %v", err)
	}
	defer document.Close()

	replacements := rules.RelativeImportRule{}.Collect(document, rules.RelativeImportInput{
		File:      "collector/src/collector/assembly/cache_files/loader.py",
		Package:   "collector.assembly.cache_files",
		OldModule: "collector.assembly.cache_files.snapshot_manifest",
		NewModule: "collector.assembly.cache_files.summary_manifest",
	})

	if len(replacements) != 1 {
		t.Fatalf("expected 1 replacement, got %#v", replacements)
	}
	replacement := replacements[0]
	if string(source[replacement.Start:replacement.End]) != ".snapshot_manifest" {
		t.Fatalf("replacement range points to %q", string(source[replacement.Start:replacement.End]))
	}
	if replacement.Replacement != "collector.assembly.cache_files.summary_manifest" {
		t.Fatalf("expected absolute replacement module, got %q", replacement.Replacement)
	}
}

func TestRelativeImportRuleUpdatesPackageRelativeImportedName(t *testing.T) {
	source := []byte("from . import snapshot_manifest as manifest\n")
	document, err := python.Parse(source)
	if err != nil {
		t.Fatalf("parse python: %v", err)
	}
	defer document.Close()

	replacements := rules.RelativeImportRule{}.Collect(document, rules.RelativeImportInput{
		File:      "collector/src/collector/assembly/cache_files/loader.py",
		Package:   "collector.assembly.cache_files",
		OldModule: "collector.assembly.cache_files.snapshot_manifest",
		NewModule: "collector.assembly.cache_files.summary_manifest",
	})

	if len(replacements) != 2 {
		t.Fatalf("expected 2 replacements, got %#v", replacements)
	}

	updated := string(source)
	for index := len(replacements) - 1; index >= 0; index-- {
		replacement := replacements[index]
		updated = updated[:replacement.Start] + replacement.Replacement + updated[replacement.End:]
	}
	expected := "from collector.assembly.cache_files import summary_manifest as manifest\n"
	if updated != expected {
		t.Fatalf("unexpected updated source:\n%s", updated)
	}
}

func TestRelativeImportRuleSkipsUnrelatedRelativeImport(t *testing.T) {
	source := []byte("from .other_manifest import SnapshotManifest\n")
	document, err := python.Parse(source)
	if err != nil {
		t.Fatalf("parse python: %v", err)
	}
	defer document.Close()

	replacements := rules.RelativeImportRule{}.Collect(document, rules.RelativeImportInput{
		File:      "collector/src/collector/assembly/cache_files/loader.py",
		Package:   "collector.assembly.cache_files",
		OldModule: "collector.assembly.cache_files.snapshot_manifest",
		NewModule: "collector.assembly.cache_files.summary_manifest",
	})

	if len(replacements) != 0 {
		t.Fatalf("expected no replacements, got %#v", replacements)
	}
}
