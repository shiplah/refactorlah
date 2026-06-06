//go:build cgo

package rules_test

import (
	"testing"

	"refactorlah/internal/languages/python"
	"refactorlah/internal/languages/python/rules"
)

func TestQualifiedModuleReferenceRuleUpdatesQualifiedUsage(t *testing.T) {
	source := []byte(`def load():
    return collector.assembly.cache_files.snapshot_manifest.load()
`)
	document, err := python.Parse(source)
	if err != nil {
		t.Fatalf("parse python: %v", err)
	}
	defer document.Close()

	replacements := rules.QualifiedModuleReferenceRule{}.Collect(document, rules.QualifiedModuleReferenceInput{
		File:      "src/collector/assembly/cache_files/loader.py",
		OldModule: "collector.assembly.cache_files.snapshot_manifest",
		NewModule: "collector.assembly.cache_files.summary_manifest",
	})

	if len(replacements) != 1 {
		t.Fatalf("expected 1 replacement, got %#v", replacements)
	}
	replacement := replacements[0]
	if string(source[replacement.Start:replacement.End]) != "collector.assembly.cache_files.snapshot_manifest" {
		t.Fatalf("replacement range points to %q", string(source[replacement.Start:replacement.End]))
	}
}

func TestQualifiedModuleReferenceRuleSkipsImportStatements(t *testing.T) {
	source := []byte("import collector.assembly.cache_files.snapshot_manifest\n")
	document, err := python.Parse(source)
	if err != nil {
		t.Fatalf("parse python: %v", err)
	}
	defer document.Close()

	replacements := rules.QualifiedModuleReferenceRule{}.Collect(document, rules.QualifiedModuleReferenceInput{
		File:      "src/collector/assembly/cache_files/loader.py",
		OldModule: "collector.assembly.cache_files.snapshot_manifest",
		NewModule: "collector.assembly.cache_files.summary_manifest",
	})

	if len(replacements) != 0 {
		t.Fatalf("expected no replacements, got %#v", replacements)
	}
}

func TestQualifiedModuleReferenceRuleDoesNotRewriteLongerSimilarModule(t *testing.T) {
	source := []byte(`def load():
    return collector.assembly.cache_files.snapshot_manifest_extra.load()
`)
	document, err := python.Parse(source)
	if err != nil {
		t.Fatalf("parse python: %v", err)
	}
	defer document.Close()

	replacements := rules.QualifiedModuleReferenceRule{}.Collect(document, rules.QualifiedModuleReferenceInput{
		File:      "src/collector/assembly/cache_files/loader.py",
		OldModule: "collector.assembly.cache_files.snapshot_manifest",
		NewModule: "collector.assembly.cache_files.summary_manifest",
	})

	if len(replacements) != 0 {
		t.Fatalf("expected no replacements, got %#v", replacements)
	}
}
