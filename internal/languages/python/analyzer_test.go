//go:build cgo

package python

import (
	"testing"

	adapterproto "refactorlah/internal/adapters"
	"refactorlah/internal/planning"
)

func TestAnalyzerUpdatesAbsoluteAndRelativeImports(t *testing.T) {
	root := t.TempDir()
	writePythonFixture(t, root, "src/collector/__init__.py", "")
	writePythonFixture(t, root, "src/collector/assembly/__init__.py", "")
	writePythonFixture(t, root, "src/collector/assembly/cache_files/__init__.py", "")
	writePythonFixture(t, root, "src/collector/assembly/cache_files/snapshot_manifest.py", "class SnapshotManifest: pass\n")
	writePythonFixture(t, root, "src/collector/assembly/cache_files/loader.py", `from collector.assembly.cache_files.snapshot_manifest import SnapshotManifest
from .snapshot_manifest import SnapshotManifest as LocalSnapshotManifest
from . import snapshot_manifest as manifest
`)

	response, relevant, err := NewAnalyzer().Analyze(root, planning.MovePlan{
		Moves: []planning.FileMove{{
			OldPath: "src/collector/assembly/cache_files/snapshot_manifest.py",
			NewPath: "src/collector/assembly/cache_files/summary_manifest.py",
		}},
	})
	if err != nil {
		t.Fatalf("analyze python: %v", err)
	}
	if !relevant {
		t.Fatal("expected python analyzer to be relevant")
	}
	if len(response.SymbolMappings) != 1 {
		t.Fatalf("expected 1 symbol mapping, got %#v", response.SymbolMappings)
	}

	assertPythonReplacement(t, response.Replacements, "src/collector/assembly/cache_files/loader.py", "collector.assembly.cache_files.summary_manifest")
	assertPythonReplacement(t, response.Replacements, "src/collector/assembly/cache_files/loader.py", "summary_manifest")
}

func TestAnalyzerWarnsForPythonFileOutsideSourceRoots(t *testing.T) {
	root := t.TempDir()
	writePythonFixture(t, root, "tools/snapshot_manifest.py", "class SnapshotManifest: pass\n")

	response, relevant, err := NewAnalyzer().Analyze(root, planning.MovePlan{
		Moves: []planning.FileMove{{
			OldPath: "tools/snapshot_manifest.py",
			NewPath: "other/summary_manifest.py",
		}},
	})
	if err != nil {
		t.Fatalf("analyze python: %v", err)
	}
	if !relevant {
		t.Fatal("expected python analyzer to be relevant")
	}
	if len(response.Warnings) != 0 {
		t.Fatalf("fallback source root should derive modules without warnings, got %#v", response.Warnings)
	}
}

func assertPythonReplacement(t *testing.T, replacements []adapterproto.Replacement, file string, newText string) {
	t.Helper()

	for _, replacement := range replacements {
		if replacement.File == file && replacement.Replacement == newText {
			return
		}
	}
	t.Fatalf("expected replacement in %s to %q, got %#v", file, newText, replacements)
}
