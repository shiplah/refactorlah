//go:build cgo

package php

import (
	"testing"

	"github.com/shiplah/refactorlah/internal/parsing/treesitter"
)

func TestParserFindsRefactorRelevantNodes(t *testing.T) {
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

	document, err := Parse(source)
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
