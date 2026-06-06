//go:build cgo

package rules_test

import (
	"testing"

	"refactorlah/internal/languages/python"
	"refactorlah/internal/languages/python/rules"
)

func TestStringAnnotationRuleUpdatesAnnotationStrings(t *testing.T) {
	source := []byte(`def load(
    manifest: "collector.assembly.cache_files.snapshot_manifest.SnapshotManifest",
) -> "collector.assembly.cache_files.snapshot_manifest.SnapshotManifest":
    return manifest
`)
	document, err := python.Parse(source)
	if err != nil {
		t.Fatalf("parse python: %v", err)
	}
	defer document.Close()

	replacements := rules.StringAnnotationRule{}.Collect(document, rules.StringAnnotationInput{
		File:      "src/collector/assembly/cache_files/loader.py",
		Source:    source,
		OldModule: "collector.assembly.cache_files.snapshot_manifest",
		NewModule: "collector.assembly.cache_files.summary_manifest",
	})

	if len(replacements) != 2 {
		t.Fatalf("expected 2 replacements, got %#v", replacements)
	}
	for _, replacement := range replacements {
		if replacement.Replacement != "collector.assembly.cache_files.summary_manifest" {
			t.Fatalf("unexpected replacement %q", replacement.Replacement)
		}
	}
}

func TestStringAnnotationRuleSkipsOrdinaryStrings(t *testing.T) {
	source := []byte(`value = "collector.assembly.cache_files.snapshot_manifest.SnapshotManifest"
`)
	document, err := python.Parse(source)
	if err != nil {
		t.Fatalf("parse python: %v", err)
	}
	defer document.Close()

	replacements := rules.StringAnnotationRule{}.Collect(document, rules.StringAnnotationInput{
		File:      "src/collector/assembly/cache_files/loader.py",
		Source:    source,
		OldModule: "collector.assembly.cache_files.snapshot_manifest",
		NewModule: "collector.assembly.cache_files.summary_manifest",
	})

	if len(replacements) != 0 {
		t.Fatalf("expected no replacements, got %#v", replacements)
	}
}
