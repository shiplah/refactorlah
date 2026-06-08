//go:build cgo

package rules_test

import (
	"testing"

	"github.com/NickSdot/refactorlah/internal/adapters/python"
	"github.com/NickSdot/refactorlah/internal/adapters/python/rules"
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
		OldModule: "collector.assembly.cache_files.snapshot_manifest",
		NewModule: "collector.assembly.cache_files.summary_manifest",
	})

	if len(replacements) != 0 {
		t.Fatalf("expected no replacements, got %#v", replacements)
	}
}

func TestStringAnnotationRuleSkipsDictionaryValueStrings(t *testing.T) {
	source := []byte(`config = {
    "handler": "collector.assembly.cache_files.snapshot_manifest.SnapshotManifest",
}
`)
	document, err := python.Parse(source)
	if err != nil {
		t.Fatalf("parse python: %v", err)
	}
	defer document.Close()

	replacements := rules.StringAnnotationRule{}.Collect(document, rules.StringAnnotationInput{
		File:      "src/collector/assembly/cache_files/loader.py",
		OldModule: "collector.assembly.cache_files.snapshot_manifest",
		NewModule: "collector.assembly.cache_files.summary_manifest",
	})

	if len(replacements) != 0 {
		t.Fatalf("expected no replacements, got %#v", replacements)
	}
}

func TestStringAnnotationRuleSkipsBytesAndFStringAnnotations(t *testing.T) {
	source := []byte(`def load(
    raw: b"collector.assembly.cache_files.snapshot_manifest.SnapshotManifest",
    interpolated: f"collector.assembly.cache_files.snapshot_manifest.{name}",
) -> "collector.assembly.cache_files.snapshot_manifest.SnapshotManifest":
    return raw or interpolated
`)
	document, err := python.Parse(source)
	if err != nil {
		t.Fatalf("parse python: %v", err)
	}
	defer document.Close()

	replacements := rules.StringAnnotationRule{}.Collect(document, rules.StringAnnotationInput{
		File:      "src/collector/assembly/cache_files/loader.py",
		OldModule: "collector.assembly.cache_files.snapshot_manifest",
		NewModule: "collector.assembly.cache_files.summary_manifest",
	})

	if len(replacements) != 1 {
		t.Fatalf("expected only the plain return annotation to be replaced, got %#v", replacements)
	}
	if string(source[replacements[0].Start:replacements[0].End]) != "collector.assembly.cache_files.snapshot_manifest" {
		t.Fatalf("replacement range points to %q", string(source[replacements[0].Start:replacements[0].End]))
	}
}
