//go:build cgo

package python

import (
	"testing"

	"github.com/NickSdot/refactorlah/internal/parsing/treesitter"
)

func TestParserFindsRefactorRelevantNodes(t *testing.T) {
	source := []byte(`from collector.assembly.cache_files.snapshot_manifest import SnapshotManifest
from .snapshot_manifest import SnapshotManifest as LocalSnapshotManifest
import collector.source_bundle_helpers as sidecars

def load() -> SnapshotManifest:
    return LocalSnapshotManifest()
`)

	document, err := Parse(source)
	if err != nil {
		t.Fatalf("parse python: %v", err)
	}
	defer document.Close()

	assertNodeText(t, document, "import_from_statement", "from collector.assembly.cache_files.snapshot_manifest import SnapshotManifest")
	assertNodeText(t, document, "import_from_statement", "from .snapshot_manifest import SnapshotManifest as LocalSnapshotManifest")
	assertNodeText(t, document, "import_statement", "import collector.source_bundle_helpers as sidecars")
	assertNodeText(t, document, "dotted_name", "collector.assembly.cache_files.snapshot_manifest")
	assertNodeText(t, document, "relative_import", ".snapshot_manifest")
}

func assertNodeText(t *testing.T, document *treesitter.Document, kind string, expected string) {
	t.Helper()

	for _, node := range document.NodesByKind(kind) {
		if node.Text == expected {
			if node.StartByte < 0 || node.EndByte <= node.StartByte {
				t.Fatalf("%s node %q has invalid byte range %d..%d", kind, expected, node.StartByte, node.EndByte)
			}
			return
		}
	}

	t.Fatalf("missing %s node %q; found %#v", kind, expected, document.NodesByKind(kind))
}
