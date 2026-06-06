//go:build cgo

package treesitter

import "testing"

func TestPHPParserFindsRefactorRelevantNodes(t *testing.T) {
	source := []byte(`<?php

namespace App\Http\Controllers;

use App\Services\Billing\InvoiceService;

final class InvoiceController
{
    public function show(InvoiceService $service): \App\Services\Billing\InvoiceService
    {
        return $service;
    }
}
`)

	document, err := Parse(source, PHP())
	if err != nil {
		t.Fatalf("parse php: %v", err)
	}
	defer document.Close()

	assertNodeText(t, document, "namespace_definition", "namespace App\\Http\\Controllers;")
	assertNodeText(t, document, "namespace_use_declaration", "use App\\Services\\Billing\\InvoiceService;")
	assertNodeText(t, document, "class_declaration", "final class InvoiceController\n{\n    public function show(InvoiceService $service): \\App\\Services\\Billing\\InvoiceService\n    {\n        return $service;\n    }\n}")
	assertNodeText(t, document, "namespace_name", "App\\Http\\Controllers")
	assertNodeText(t, document, "qualified_name", "App\\Services\\Billing\\InvoiceService")
	assertNodeText(t, document, "qualified_name", "\\App\\Services\\Billing\\InvoiceService")
}

func TestPythonParserFindsRefactorRelevantNodes(t *testing.T) {
	source := []byte(`from collector.assembly.cache_files.snapshot_manifest import SnapshotManifest
from .snapshot_manifest import SnapshotManifest as LocalSnapshotManifest
import collector.source_bundle_helpers as sidecars

def load() -> SnapshotManifest:
    return LocalSnapshotManifest()
`)

	document, err := Parse(source, Python())
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

func assertNodeText(t *testing.T, document *Document, kind string, expected string) {
	t.Helper()

	for _, node := range document.NodesByKind(kind) {
		if node.Text == expected {
			if node.StartByte < 0 || node.EndByte <= node.StartByte {
				t.Fatalf("%s node %q has invalid byte range %d..%d", kind, expected, node.StartByte, node.EndByte)
			}
			if string(document.source[node.StartByte:node.EndByte]) != expected {
				t.Fatalf("%s node range %d..%d does not resolve to %q", kind, node.StartByte, node.EndByte, expected)
			}
			return
		}
	}

	t.Fatalf("missing %s node %q; found %#v", kind, expected, document.NodesByKind(kind))
}
